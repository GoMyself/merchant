package model

import (
	"database/sql"
	"errors"
	"fmt"
	"merchant2/contrib/helper"

	g "github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
)

type Platform struct {
	ID            string `db:"id" json:"id"`
	Name          string `db:"name" json:"name"`
	Code          string `db:"code" json:"code"`
	State         int    `db:"state" json:"state"`
	Seq           int64  `db:"seq" json:"seq"`
	GameCode      string `db:"game_code" json:"game_code"`
	CreatedAt     int32  `db:"created_at" json:"created_at"`
	UpdatedAt     int32  `db:"updated_at" json:"updated_at"`
	MaintainBegin int64  `db:"maintain_begin" json:"maintain_begin"`
	MaintainEnd   int64  `db:"maintain_end" json:"maintain_end"`
	Wallet        int    `db:"wallet" json:"wallet"`
	GameType      int    `db:"game_type" json:"game_type"`
}

type PlatRate struct {
	PID  string  `db:"pid" json:"pid"`
	Name string  `db:"name" json:"name"`
	Rate float64 `db:"rate" json:"rate"`
}

type platJson struct {
	ID       string `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	Code     string `db:"code" json:"code"`
	GameType int64  `db:"game_type" json:"game_type"`
	State    int64  `db:"state" json:"state"`
	Seq      int64  `db:"seq" json:"seq"`
	Wallet   int64  `db:"wallet" json:"wallet"`
	GameCode string `db:"game_code" json:"game_code"`
}

type PlatformData struct {
	T int64      `json:"t"`
	D []Platform `json:"d"`
	S uint       `json:"s"`
}

type navJson struct {
	Cate
	L []platJson `json:"l"`
}

/**
 * @Description: 更新场馆(状态,锁定钱包,排序)
 * flag=0,更新场馆排序
 * flag=1,更新场馆上线状态
 * flag=2,更新场馆钱包状态
 * @Author: parker
 * @Date: 2021/4/3 10:43
 * @LastEditTime: 2021/4/3 17:43
 * @LastEditors: parker
 */
func PlatformUpdate(id uint64, flag uint8, state, seq int, lastTime uint32) error {

	record, err := platformRecord(id, flag, state, seq, lastTime)
	if err != nil {
		return err
	}

	query, _, _ := dialect.Update("tbl_platforms").Set(record).Where(g.Ex{"id": id}).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return PlatToMinio()
}

/**
 * @Description: platformRecord 更新内容构建
 * flag==1更新场馆状态
 * 同时根据state==1时修改第三方维护开始时间,同时将state状态修改为下线,反之修改第三方维护结束时间,state为上线状态
 * flag==2更新场馆锁定钱包,wallet状态为0设置为1,状态为1时设置为0
 * @Author: parker
 * @Date: 2021/4/3 10:43
 * @LastEditTime: 2021/4/3 17:43
 * @LastEditors: parker
 */
func platformRecord(id uint64, flag uint8, state, seq int, lastTime uint32) (g.Record, error) {

	record := g.Record{}
	if flag == PlatformFlagEdit {

		var sid string
		query, _, _ := dialect.From("tbl_platforms").
			Select("id").Where(g.Ex{"seq": seq, "prefix": meta.Prefix}).Limit(1).ToSQL()
		err := meta.MerchantDB.Get(&sid, query)
		if err != nil {
			if err != sql.ErrNoRows {
				return nil, pushLog(err, helper.DBErr)
			}
		}

		if len(sid) > 0 && sid != fmt.Sprintf("%d", id) {
			return nil, errors.New(helper.RecordExistErr)
		}

		record["seq"] = seq
	}

	data, err := PlatformFindOne(g.Ex{"id": id})
	if err != nil {
		return record, err
	}

	if flag == PlatformFlagState {

		if data.State == state {
			return nil, errors.New(helper.NoDataUpdate)
		}

		record["state"] = state

		if data.State == 0 {
			record["maintain_end"] = lastTime
		}

		if data.State == 1 {
			record["maintain_begin"] = lastTime
		}
	}

	if flag == PlatformFlagWallet {

		if data.Wallet == state {
			return nil, errors.New(helper.NoDataUpdate)
		}

		record["wallet"] = state
	}

	record["updated_at"] = lastTime

	return record, nil
}

/**
 * @Description: 查询单个场馆
 * @Author: parker
 * @Date: 2021/4/3 10:43
 * @LastEditTime: 2021/4/3 17:43
 * @LastEditors: parker
 */
func PlatformFindOne(ex g.Ex) (Platform, error) {

	data := Platform{}
	ex["prefix"] = meta.Prefix
	query, _, _ := dialect.From("tbl_platforms").Select(colsPlatform...).Where(ex).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&data, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

/**
 * @Description: 下拉选项,读取redis中数据
 * @Author: parker
 * @Date: 2021/4/3 10:43
 * @LastEditTime: 2021/4/3 17:43
 * @LastEditors: parker
 */
func PlatListRedis() string {

	res, err := meta.MerchantRedis.Get(ctx, "plat").Result()
	if err == redis.Nil || err != nil {
		return "[]"
	}

	return res
}

/**
 * @Description: 场馆列表查询及分页
 * @Author: parker
 * @Date: 2021/4/3 10:43
 * @LastEditTime: 2021/4/3 17:43
 * @LastEditors: parker
 */
func PlatformList(ex g.Ex, pageSize, page uint) (PlatformData, error) {

	data := PlatformData{
		S: pageSize,
	}
	ex["prefix"] = meta.Prefix
	offset := (page - 1) * pageSize
	t := dialect.From("tbl_platforms")
	if page == 1 {

		query, _, _ := t.Select(g.COUNT("id")).Where(ex).ToSQL()
		err := meta.MerchantDB.Get(&data.T, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}

		if data.T == 0 {
			return data, nil
		}
	}
	query, _, _ := t.Select(colsPlatform...).Where(ex).Order(g.C("created_at").Asc()).Offset(offset).Limit(pageSize).ToSQL()
	err := meta.MerchantDB.Select(&data.D, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func PlatToMinio() error {

	var data []platJson

	ex := g.Ex{
		"state":  1,
		"prefix": meta.Prefix,
	}
	query, _, _ := dialect.From("tbl_platforms").
		Select(colsPlatJson...).Where(ex).Order(g.C("seq").Asc()).ToSQL()
	err := meta.MerchantDB.Select(&data, query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	if len(data) == 0 {
		return errors.New(helper.RecordNotExistErr)
	}

	b, err := helper.JsonMarshal(data)
	if err != nil {
		return errors.New(helper.FormatErr)
	}

	navJ, err := NavMinio()
	if err != nil {
		return err
	}

	navB, err := helper.JsonMarshal(navJ)
	if err != nil {
		return errors.New(helper.FormatErr)
	}

	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()

	for _, val := range data {
		k := fmt.Sprintf("plat:%s", val.ID)
		b1, err := helper.JsonMarshal(val)
		if err != nil {
			fmt.Println("PlatToMinio error = ", err)
			continue
		}
		pipe.Unlink(ctx, k)
		pipe.Set(ctx, k, string(b1), 0)
		pipe.Persist(ctx, k)
	}

	pipe.Unlink(ctx, "nav", "plat")
	pipe.Set(ctx, "nav", string(navB), 0)
	pipe.Persist(ctx, "nav")
	pipe.Set(ctx, "plat", string(b), 0)
	pipe.Persist(ctx, "plat")
	_, err = pipe.Exec(ctx)
	if err != nil {
		return pushLog(err, "redis")
	}

	return nil
}

func NavMinio() ([]navJson, error) {

	var top []Cate
	query, _, _ := dialect.From("tbl_tree").
		Where(g.C("level").ILike("0010%"), g.C("prefix").Eq(meta.Prefix)).Order(g.C("sort").Asc()).ToSQL()
	err := meta.MerchantDB.Select(&top, query)
	if err != nil {
		return nil, pushLog(err, helper.DBErr)
	}

	topLen := len(top)
	if topLen == 0 {
		fmt.Println("NavMinio query = ", query)
		return nil, errors.New(helper.RecordNotExistErr)
	}

	data := make([]navJson, topLen)
	for k, v := range top {

		data[k].Cate = v
		ex := g.Ex{
			"state":     1,
			"game_type": v.ID,
			"prefix":    meta.Prefix,
		}

		query, _, _ = dialect.From("tbl_platforms").Select(colsPlatJson...).Where(ex).Order(g.C("seq").Asc()).ToSQL()
		err = meta.MerchantDB.Select(&data[k].L, query)
		if err != nil {
			fmt.Println("platform query = ", query)
			continue
		}
	}

	return data, nil
}

func PlatformRate() ([]PlatRate, error) {

	var data []PlatRate
	query, _, _ := dialect.From("tbl_platform_rate").Select("name", "pid", "rate").Where(g.Ex{"prefix": meta.Prefix}).ToSQL()
	err := meta.MerchantDB.Select(&data, query)
	if err != nil {
		return nil, pushLog(err, helper.DBErr)
	}

	return data, nil
}
