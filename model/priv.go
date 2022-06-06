package model

import (
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
	"merchant/contrib/helper"
)

func PrivList(gid string) (string, error) {

	//privAllKey := fmt.Sprintf("%s:priv:PrivAll", meta.Prefix)
	gKey := fmt.Sprintf("%s:priv:list:GM%d", meta.Prefix, gid)
	cmd := meta.MerchantRedis.Get(ctx, gKey)
	fmt.Println(cmd.String())
	val, err := cmd.Result()
	if err != nil && err != redis.Nil {
		return val, pushLog(err, helper.RedisErr)
	}

	return val, nil
}

func PrivCheck(uri, gid string) error {

	key := fmt.Sprintf("%s:priv:PrivMap", meta.Prefix)
	cmd := meta.MerchantRedis.HGet(ctx, key, uri)
	fmt.Println(cmd.String())
	privId, err := cmd.Result()
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	id := fmt.Sprintf("%s:priv:GM%s", meta.Prefix, gid)
	hcmd := meta.MerchantRedis.HExists(ctx, id, privId)
	fmt.Println(hcmd.String())
	exists := hcmd.Val()
	if !exists {
		return errors.New("404")
	}

	return nil
}

/**
 * @Description: 刷新缓存
 * @Author: carl
 */
func LoadPrivs() error {

	var records []Priv

	query, _, _ := dialect.From("tbl_admin_priv").
		Select("pid", "state", "id", "name", "sortlevel", "module").Where(g.Ex{"prefix": meta.Prefix}).Order(g.C("sortlevel").Asc()).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Select(&records, query)
	if err != nil {
		return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	recs, err := helper.JsonMarshal(records)
	if err != nil {
		return errors.New(helper.FormatErr)
	}

	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()

	privMapKey := fmt.Sprintf("%s:priv:PrivMap", meta.Prefix)
	privAllKey := fmt.Sprintf("%s:priv:PrivAll", meta.Prefix)
	pipe.Unlink(ctx, privAllKey)
	pipe.Unlink(ctx, privMapKey)
	pipe.Set(ctx, privAllKey, string(recs), 0)

	for _, val := range records {
		id := fmt.Sprintf("%d", val.ID)
		pipe.HSet(ctx, privMapKey, val.Module, id)
	}
	pipe.Persist(ctx, privAllKey)
	pipe.Persist(ctx, privMapKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	return nil
}
