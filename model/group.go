package model

import (
	"database/sql"
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
	"merchant2/contrib/helper"
	"strings"
)

type Group struct {
	CreateAt   int32  `db:"create_at" rule:"none" json:"create_at"`                                                            //创建时间
	Gid        int64  `db:"gid" rule:"none" json:"gid"`                                                                        //
	Gname      string `db:"gname" name:"gname" rule:"chn" min:"2" max:"8" msg:"gname error[2-8]" json:"gname"`                 //组名
	Lft        int64  `db:"lft" rule:"none" json:"lft"`                                                                        //节点左值
	Lvl        int64  `db:"lvl" rule:"none" json:"lvl"`                                                                        //
	Noted      string `db:"noted" name:"noted" rule:"none" default:"" min:"0" max:"511" msg:"noted error[0-511]" json:"noted"` //备注信息
	Permission string `db:"permission" name:"permission" rule:"sDigit" min:"2" msg:"permission error[2-]" json:"permission"`   //权限模块ID
	Rgt        int64  `db:"rgt" rule:"none" json:"rgt"`                                                                        //节点右值
	Pid        string `db:"pid" rule:"none" json:"pid"`                                                                        //父节点
	State      int    `db:"state" json:"state" name:"state" rule:"digit" min:"0" max:"1" msg:"state error"`                    //0:关闭1:开启
	Prefix     string `db:"prefix" rule:"none" json:"prefix"`
}

func GroupUpdate(id string, data Group) error {

	var gid string
	query, _, _ := dialect.From("tbl_admin_group").Select("gid").Where(g.Ex{"gname": data.Gname, "prefix": meta.Prefix}).ToSQL()
	err := meta.MerchantDB.Get(&gid, query)
	if err != nil && err != sql.ErrNoRows {
		body := fmt.Errorf("%s,[%s]", err.Error(), query)
		return pushLog(body, helper.DBErr)
	}

	if gid != id && gid != "" {
		return errors.New(helper.RecordExistErr)
	}

	record := g.Record{
		"gname":      data.Gname,
		"noted":      data.Noted,
		"state":      data.State,
		"permission": data.Permission,
	}
	query, _, _ = dialect.Update("tbl_admin_group").Set(record).Where(g.Ex{"gid": id}).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		body := fmt.Errorf("%s,[%s]", err.Error(), query)
		return pushLog(body, helper.DBErr)
	}

	return GroupRefresh()
}

func GroupRefresh() error {

	var records []Group
	cols := []interface{}{"noted", "gid", "gname", "permission", "create_at", "state", "lft", "rgt", "lvl", "pid"}
	ex := g.Ex{
		"prefix": meta.Prefix,
	}
	query, _, _ := dialect.From("tbl_admin_group").Select(cols...).Where(ex).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Select(&records, query)
	if err != nil {
		body := fmt.Errorf("%s,[%s]", err.Error(), query)
		return pushLog(body, helper.DBErr)
	}

	recs, err := helper.JsonMarshal(records)
	if err != nil {
		return errors.New(helper.FormatErr)
	}

	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()

	key := fmt.Sprintf("%s:priv:GroupAll", meta.Prefix)
	pipe.Unlink(ctx, key)
	pipe.Set(ctx, key, string(recs), 0)
	pipe.Persist(ctx, key)

	for _, val := range records {

		id := fmt.Sprintf("%s:priv:GM%d", meta.Prefix, val.Gid)
		pipe.Unlink(ctx, id)
		// 只保存开启状态的分组
		if val.State == 1 {
			data := strings.Split(val.Permission, ",")
			for _, v := range data {
				pipe.HSet(ctx, id, v, "1")
			}
			pipe.Persist(ctx, id)
		}
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	return nil
}

func GroupInsert(pid string, data Group) error {

	var gid string
	query, _, _ := dialect.From("tbl_admin_group").Select("gid").Where(g.Ex{"gname": data.Gname, "prefix": meta.Prefix}).ToSQL()
	err := meta.MerchantDB.Get(&gid, query)
	if err != nil && err != sql.ErrNoRows {
		body := fmt.Errorf("%s,[%s]", err.Error(), query)
		return pushLog(body, helper.DBErr)
	}

	if gid != "" {
		return errors.New(helper.RecordExistErr)
	}

	parent := Group{}

	err = meta.MerchantDB.Get(&parent, "SELECT `lvl`,`lft`,`rgt` FROM `tbl_admin_group` WHERE gid = ? and prefix =?;", pid, meta.Prefix)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	_, err = tx.Exec("UPDATE `tbl_admin_group` SET lft = lft + 2 WHERE lft > ? and prefix =?", parent.Lft, meta.Prefix)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	_, err = tx.Exec("UPDATE `tbl_admin_group` SET rgt = rgt + 2 WHERE rgt > ? and prefix =?", parent.Lft, meta.Prefix)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	data.Lvl = parent.Lvl + 1
	data.Lft = parent.Lft + 1
	data.Rgt = parent.Lft + 2
	data.Pid = pid
	data.Prefix = meta.Prefix
	query, _, _ = dialect.Insert("tbl_admin_group").Rows(data).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		body := fmt.Errorf("%s,[%s]", err.Error(), query)
		return pushLog(body, helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		body := fmt.Errorf("%s,[%s]", err.Error(), query)
		return pushLog(body, helper.DBErr)
	}

	return GroupRefresh()
}

func GroupList() (string, error) {

	key := fmt.Sprintf("%s:priv:GroupAll", meta.Prefix)
	val, err := meta.MerchantRedis.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return val, pushLog(err, helper.RedisErr)
	}

	return val, nil
}

func GroupFindOne(ex g.Ex) (Group, error) {

	data := Group{}
	query, _, _ := dialect.From("tbl_group").Select(colsGroup...).Where(ex).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&data, query)
	if err != nil {
		body := fmt.Errorf("%s,[%s]", err.Error(), query)
		return data, pushLog(body, helper.DBErr)
	}

	return data, nil
}

func GroupFindAll(ex g.Ex) ([]Group, error) {

	var data []Group
	query, _, _ := dialect.From("tbl_group").Select(colsGroup...).Where(ex).ToSQL()
	err := meta.MerchantDB.Select(&data, query)
	if err != nil {
		body := fmt.Errorf("%s,[%s]", err.Error(), query)
		return data, pushLog(body, helper.DBErr)
	}

	return data, nil
}
