package model

import (
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
	"merchant/contrib/helper"
	"time"
)

func ShortURLSet(uri string) error {

	key := meta.Prefix + ":shorturl:domain"
	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()

	pipe.Set(ctx, key, uri, 100*time.Hour)
	pipe.Persist(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	return nil
}

func ShortURLGet() (string, error) {

	key := meta.Prefix + ":shorturl:domain"
	cmd := meta.MerchantRedis.Get(ctx, key)
	//fmt.Println(cmd.String())
	resc, err := cmd.Result()
	if err != nil && err != redis.Nil {
		return "", pushLog(err, helper.RedisErr)
	}

	if err == redis.Nil {
		return "", redis.Nil
	}

	return resc, nil
}

func ShortURLInitNoAd() error {

	var tss []string
	query, _, _ := dialect.From("shorturl").Select("ts").Where(g.Ex{}).ToSQL()
	fmt.Println(query)
	err := meta.MerchantTD.Select(&tss, query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	p := len(tss) / 100
	l := len(tss) % 100
	if l > 0 {
		p += 1
	}

	for i := 0; i < p; i++ {
		offset := i * 100
		d := tss[offset:]
		if i < p-1 {
			d = tss[offset : offset+100]
		}
		var records []g.Record
		for _, ts := range d {
			fmt.Println(ts)
			t, err := time.ParseInLocation("2006-01-02T15:04:05.999999+07:00", ts, loc)
			if err != nil {
				return pushLog(err, helper.DateTimeErr)
			}

			record := g.Record{
				"ts":    t.UnixMicro(),
				"no_ad": 0,
			}
			records = append(records, record)
		}
		query, _, _ = dialect.Insert("shorturl").Rows(records).ToSQL()
		fmt.Println(query)
		_, err = meta.MerchantTD.Exec(query)
		if err != nil {
			return pushLog(err, helper.DBErr)
		}
	}

	return nil
}
