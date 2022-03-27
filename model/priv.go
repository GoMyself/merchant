package model

import (
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
	"merchant2/contrib/helper"
)

func PrivList() (string, error) {

	val, err := meta.MerchantRedis.Get(ctx, "PrivAll").Result()
	if err != nil && err != redis.Nil {
		return val, pushLog(err, helper.RedisErr)
	}

	return val, nil
}

func PrivCheck(uri, gid string) error {

	privId, err := meta.MerchantRedis.HGet(ctx, "PrivMap", uri).Result()
	if err != nil {
		return err
	}

	id := fmt.Sprintf("GM%s", gid)
	exists := meta.MerchantRedis.HExists(ctx, id, privId).Val()
	if !exists {
		return errors.New("404")
	}

	return nil
}

/**
 * @Description: 刷新缓存
 * @Author: carl
 */
func PrivRefresh() error {

	var records []Priv

	// select * from tbl_admin_priv order by `sortlevel` asc;
	query, _, _ := dialect.From("tbl_admin_priv").
		Select("pid", "state", "id", "name", "sortlevel", "module").Where(g.Ex{"prefix": meta.Prefix}).Order(g.C("sortlevel").Asc()).ToSQL()

	err := meta.MerchantDB.Select(&records, query)
	if err != nil {
		return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	recs, err := helper.JsonMarshal(records)
	if err != nil {
		return errors.New(helper.FormatErr)
	}

	pipe := meta.MerchantRedis.TxPipeline()
	pipe.Unlink(ctx, "PrivAll", "PrivMap")
	pipe.Set(ctx, "PrivAll", string(recs), 0)

	for _, val := range records {
		id := fmt.Sprintf("%d", val.ID)
		pipe.HSet(ctx, "PrivMap", val.Module, id)
	}
	pipe.Persist(ctx, "PrivAll")
	pipe.Persist(ctx, "PrivMap")

	_, err = pipe.Exec(ctx)
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	err = PrivLevelCache(records)
	if err != nil {
		return err
	}

	return nil
}

//PrivLevelCache 处理权限分级保存至 redis
func PrivLevelCache(values []Priv) error {

	data := make(map[string]*PrivTree)
	privFormatByPid(0, "", values, data)

	pipe := meta.MerchantRedis.TxPipeline()
	pipe.Unlink(ctx, "priv_tree")
	// 去除前两级 存入redis hash
	for i, v := range data {
		if v.Parent == nil || v.Parent.Parent == nil {
			continue
		}
		pipe.HSet(ctx, "priv_tree", i, v)
	}

	pipe.Persist(ctx, "priv_tree")

	_, err := pipe.Exec(ctx)
	if err != nil {
		fmt.Println(err)
		return pushLog(err, helper.RedisErr)
	}

	return nil
}

func privFormatByPid(pid int64, index string, privs []Priv, data map[string]*PrivTree) {

	for i, v := range privs {
		if pid == v.Pid {
			data[v.Module] = &PrivTree{
				Priv:   &privs[i],
				Parent: nil,
			}

			if d, ok := data[index]; ok {
				data[v.Module].Parent = d
			}
			// 获取子级
			privFormatByPid(v.ID, v.Module, privs, data)
		}
	}
}

func (c *PrivTree) MarshalBinary() ([]byte, error) {
	return helper.JsonMarshal(c)
}
