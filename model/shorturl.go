package model

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"merchant/contrib/helper"
)

func ShortURLSet(uri string) error {

	key := meta.Prefix + ":shorturl:domain"
	cmd := meta.MerchantRedis.Set(ctx, key, uri, -1)
	fmt.Println(cmd.String())
	err := cmd.Err()
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	return nil
}

func ShortURLGet() (string, error) {

	key := meta.Prefix + ":shorturl:domain"
	cmd := meta.MerchantRedis.Get(ctx, key)
	fmt.Println(cmd.String())
	resc, err := cmd.Result()
	if err != nil && err != redis.Nil {
		return "", pushLog(err, helper.RedisErr)
	}

	if err == redis.Nil {
		return "", redis.Nil
	}

	return resc, nil
}
