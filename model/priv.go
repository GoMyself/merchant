package model

import (
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
	"merchant/contrib/helper"
)

func PrivList() (string, error) {

	privAllKey := fmt.Sprintf("%s:priv:PrivAll", meta.Prefix)
	val, err := meta.MerchantRedis.Get(ctx, privAllKey).Result()
	if err != nil && err != redis.Nil {
		return val, pushLog(err, helper.RedisErr)
	}

	return val, nil
}

func PrivCheck(uri, gid string) error {

	key := fmt.Sprintf("%s:priv:PrivMap", meta.Prefix)
	privId, err := meta.MerchantRedis.HGet(ctx, key, uri).Result()
	if err != nil {
		return err
	}

	id := fmt.Sprintf("%s:priv:GM%s", meta.Prefix, gid)
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

func (c *PrivTree) MarshalBinary() ([]byte, error) {
	return helper.JsonMarshal(c)
}
