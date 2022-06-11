package model

import (
	"github.com/go-redis/redis/v8"
	"merchant/contrib/helper"
)

func ShortURLSet(uri string) error {

	key := meta.Prefix + ":shorturl:domain"
	err := meta.MerchantRedis.Set(ctx, key, uri, -1).Err()
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	return nil
}

func ShortURLGet() (string, error) {

	key := meta.Prefix + ":shorturl:domain"
	resc, err := meta.MerchantRedis.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return "", pushLog(err, helper.RedisErr)
	}

	if err == redis.Nil {
		return "", redis.Nil
	}

	return resc, nil
}
