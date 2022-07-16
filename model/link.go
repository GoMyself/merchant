package model

import (
	"database/sql"
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
	"merchant/contrib/helper"
	"time"
)

func LinkList(page, pageSize uint, ex g.Ex) (LinkData, error) {

	data := LinkData{}
	key := meta.Prefix + ":shorturl:domain"
	shortDomain, err := meta.MerchantRedis.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return data, pushLog(err, helper.RedisErr)
	}

	if err == redis.Nil || shortDomain == "" {
		shortDomain = "https://s.p3vn.co/"
	}

	ex["prefix"] = meta.Prefix
	t := dialect.From("tbl_member_link")
	if page == 1 {
		query, _, _ := t.Select(g.COUNT(1)).Where(ex).ToSQL()
		err := meta.MerchantDB.Get(&data.T, query)
		if err != nil {
			return data, pushLog(fmt.Errorf("query : %s, error : %s", query, err.Error()), helper.DBErr)
		}
	}

	offset := pageSize * (page - 1)
	query, _, _ := t.Select(colsLink...).Where(ex).Order(g.C("created_at").Desc()).Offset(offset).Limit(pageSize).ToSQL()
	err = meta.MerchantDB.Select(&data.D, query)
	if err != nil {
		return data, pushLog(fmt.Errorf("query : %s, error : %s", query, err.Error()), helper.DBErr)
	}

	for k := range data.D {
		data.D[k].ShortURL = shortDomain + data.D[k].ShortURL
	}

	return data, nil
}

type shortUR struct {
	TS   string `db:"ts"`
	NoAd string `db:"no_ad"`
}

func LinkSetNoAd(shortCode, noAd string) error {

	data := shortUR{}
	ex := g.Ex{
		"short_url": shortCode,
	}
	query, _, _ := dialect.From("shorturl").Select("ts", "no_ad").Where(ex).ToSQL()
	fmt.Println(query)
	err := meta.MerchantTD.Get(&data, query)
	if err != nil && err != sql.ErrNoRows {
		return pushLog(err, helper.DBErr)
	}

	if err == sql.ErrNoRows {
		return errors.New(helper.RecordNotExistErr)
	}

	if data.NoAd == noAd {
		return errors.New(helper.NoDataUpdate)
	}

	l := len(data.TS)
	if l < 26 {
		return errors.New(helper.DateTimeErr)
	}
	data.TS = data.TS[:l-6] + "+" + data.TS[l-5:]
	t, err := time.ParseInLocation("2006-01-02T15:04:05.999999+07:00", data.TS, loc)
	if err != nil {
		return pushLog(err, helper.DateTimeErr)
	}

	record := g.Record{
		"ts":    t.UnixMicro(),
		"no_ad": noAd,
	}
	query, _, _ = dialect.Insert("shorturl").Rows(record).ToSQL()
	fmt.Println(query)
	_, err = meta.MerchantTD.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	query, _, _ = dialect.Update("tbl_member_link").Set(g.Record{"no_ad": noAd}).Where(ex).ToSQL()
	fmt.Println(query)
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

func LinkDelete(uid, id string) error {

	query, _, _ := dialect.Delete("tbl_member_link").Where(g.Ex{
		"id":     id,
		"uid":    uid,
		"prefix": meta.Prefix,
	}).ToSQL()
	_, err := meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	LoadMemberLinks(uid)

	return nil
}

func LoadMemberLinks(uid string) {

	ex := g.Ex{
		"uid":    uid,
		"prefix": meta.Prefix,
	}
	var data []Link_t
	query, _, _ := dialect.From("tbl_member_link").Where(ex).Select(colsLink...).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Select(&data, query)
	if err != nil {
		if err != sql.ErrNoRows {
			_ = pushLog(err, helper.DBErr)
		}
		return
	}

	links := make(map[string]Link_t)
	for _, v := range data {
		links["$"+v.ID] = v
	}

	value, err := helper.JsonMarshal(&links)
	if err != nil {
		_ = pushLog(err, helper.FormatErr)
		return
	}

	key := fmt.Sprintf("%s:lk:%s", meta.Prefix, uid)
	pipe := meta.MerchantRedis.TxPipeline()
	pipe.Unlink(ctx, key)
	pipe.Do(ctx, "JSON.SET", key, ".", string(value))
	pipe.Persist(ctx, key)

	_, err = pipe.Exec(ctx)
	if err != nil {
		fmt.Println(key, string(value), err)
		_ = pushLog(err, helper.RedisErr)
		return
	}

	_ = pipe.Close()
}
