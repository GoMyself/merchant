package model

import (
	"database/sql"
	"errors"
	"fmt"
	"merchant2/contrib/helper"
	"strings"

	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/olivere/elastic/v7"
	"github.com/valyala/fasthttp"
	"github.com/wI2L/jettison"
)

// 黑名单列表
func BlacklistList(page, pageSize uint, startTime, endTime string, ty int, ex g.Ex) (BlacklistData, error) {

	data := BlacklistData{}

	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		endAt, err := helper.TimeToLoc(endTime, loc)
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}
	ex["prefix"] = meta.Prefix
	t := dialect.From("tbl_blacklist")
	if page == 1 {

		query, _, _ := t.Select(g.COUNT(1)).Where(ex).ToSQL()
		err := meta.MerchantDB.Get(&data.T, query)
		if err != nil {
			return data, pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
		}

		if data.T == 0 {
			return data, nil
		}
	}

	data.S = pageSize
	offset := (page - 1) * pageSize
	query, _, _ := t.Select(colsBlacklist...).Where(ex).
		Order(g.C("created_at").Desc()).Offset(offset).Limit(pageSize).ToSQL()
	err := meta.MerchantDB.Select(&data.D, query)
	if err != nil {
		return data, pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	return data, nil
}

// 黑名单添加
func BlacklistInsert(fctx *fasthttp.RequestCtx, ty int, value string, record g.Record) error {

	var (
		data []BankCard_t
		key  string
	)

	user, err := AdminToken(fctx)
	if err != nil {
		return errors.New(helper.AccessTokenExpires)
	}

	ex := g.Ex{
		"ty":    ty,
		"value": value,
	}
	if BlacklistExist(ex) {
		return errors.New(helper.RecordExistErr)
	}

	record["created_at"] = fctx.Time().In(loc).Unix()
	record["created_uid"] = user["id"]
	record["created_name"] = user["name"]
	record["prefix"] = meta.Prefix

	query, _, _ := dialect.Insert("tbl_blacklist").Rows(record).ToSQL()

	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		//fmt.Println("BlacklistInsert Exec err = ", err.Error())
		return errors.New(helper.DBErr)
	}

	switch ty {
	case TyDevice:
		key = fmt.Sprintf("%s:merchant:device_blacklist", meta.Prefix)
	case TyIP:
		key = fmt.Sprintf("%s:merchant:ip_blacklist", meta.Prefix)
	case TyPhone:
		key = fmt.Sprintf("%s:merchant:phone_blacklist", meta.Prefix)
	case TyBankcard:
		key = fmt.Sprintf("%s:merchant:bankcard_blacklist", meta.Prefix)
	}

	meta.MerchantRedis.Do(ctx, "CF.ADD", key, value).Val()
	valueHash := MurmurHash(value, 0)

	ex = g.Ex{
		"prefix":         meta.Prefix,
		"bank_card_hash": valueHash,
	}
	recs := g.Record{
		"state": "3",
	}
	query, _, _ = dialect.Update("tbl_member_bankcard").Set(recs).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return errors.New(helper.DBErr)
	}

	t := dialect.From("tbl_member_bankcard")
	query, _, _ = t.Select(colsBankcard...).Where(ex).ToSQL()
	err = meta.MerchantDB.Select(&data, query)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("BankcardUpdateCache err = ", err)
		return err
	}

	for _, v := range data {
		BankcardUpdateCache(v.Username)
	}

	return nil
}

// 黑名单更新
func BlacklistUpdate(ex g.Ex, record g.Record) error {

	ex["prefix"] = meta.Prefix
	query, _, _ := dialect.Update("tbl_blacklist").Set(record).Where(ex).ToSQL()
	_, err := meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	return nil
}

// 删除记录
func BlacklistDelete(id string) error {

	ex := g.Ex{
		"id":     id,
		"prefix": meta.Prefix,
	}

	data := Blacklist{}
	query, _, _ := dialect.From("tbl_blacklist").Select(colsBlacklist...).Where(ex).ToSQL()
	err := meta.MerchantDB.Get(&data, query)
	if err != nil {
		return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	query, _, _ = dialect.Delete("tbl_blacklist").Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	BlacklistLoadCache(data.Ty)

	return nil
}

// 满足条件的黑名单数量
func BlacklistExist(ex g.Ex) bool {

	var id string
	ex["prefix"] = meta.Prefix
	t := dialect.From("tbl_blacklist")
	query, _, _ := t.Select("id").Where(ex).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&id, query)
	return err != sql.ErrNoRows
}

func BlacklistLoadCache(ty int) error {

	var data []Blacklist

	if ty != 0 {
		ex := g.Ex{"ty": ty}
		query, _, _ := dialect.From("tbl_blacklist").Select(colsBlacklist...).Where(ex).ToSQL()
		fmt.Println(query)
		err := meta.MerchantDB.Select(&data, query)
		if err != nil {
			return err
		}
	} else {
		query, _, _ := dialect.From("tbl_blacklist").Select(colsBlacklist...).ToSQL()
		fmt.Println(query)
		err := meta.MerchantDB.Select(&data, query)
		if err != nil {
			return err
		}
	}

	pipe := meta.MerchantRedis.Pipeline()
	defer pipe.Close()

	deviceKey := fmt.Sprintf("%s:merchant:device_blacklist", meta.Prefix)
	ipKey := fmt.Sprintf("%s:merchant:ip_blacklist", meta.Prefix)
	phoneKey := fmt.Sprintf("%s:merchant:phone_blacklist", meta.Prefix)
	bankcardKey := fmt.Sprintf("%s:merchant:bankcard_blacklist", meta.Prefix)

	if ty != 0 {
		key := ""
		switch ty {
		case TyDevice:
			key = deviceKey
		case TyIP:
			key = ipKey
		case TyPhone:
			key = phoneKey
		case TyBankcard:
			key = bankcardKey
		}
		pipe.Unlink(ctx, key)
	} else {
		pipe.Unlink(ctx, deviceKey, ipKey, phoneKey, bankcardKey)
	}
	for _, v := range data {
		key := ""
		switch v.Ty {
		case TyDevice:
			key = deviceKey
		case TyIP:
			key = ipKey
		case TyPhone:
			key = phoneKey
		case TyBankcard:
			key = bankcardKey
		}

		pipe.Do(ctx, "CF.ADD", key, v.Value)
	}

	_, _ = pipe.Exec(ctx)

	return nil
}

func MemberAssocLoginLogList(page, pageSize int, aggField string, q *elastic.BoolQuery) (string, error) {

	fields := []string{"username", "ips", "device", "device_no", "date", "parents"}
	//q = q.Filter(q.Query, elastic.NewTermQuery("prefix", meta.Prefix))
	fsc := elastic.NewFetchSourceContext(true).Include(fields...)

	collapseBuilder := elastic.NewCollapseBuilder("username.keyword").InnerHit(elastic.NewInnerHit().
		Name(aggField).Size(100000).Collapse(elastic.NewCollapseBuilder(aggField)).FetchSourceContext(fsc))

	agg := elastic.NewCardinalityAggregation().Field("username.keyword")

	offset := pageSize * (page - 1)

	searchRes, err := meta.ES.Search().Index(esPrefixIndex("memberlogin")).
		FetchSourceContext(elastic.NewFetchSourceContext(true).Include("username")).
		Aggregation("count", agg).
		Size(pageSize).From(offset).Query(q).Sort("date", false).Collapse(collapseBuilder).Do(ctx)
	if err != nil {
		return `{"t":0,"d":[]}`, pushLog(err, helper.ESErr)
	}

	total, found := searchRes.Aggregations.Cardinality("count")
	if !found {
		return `{"t":0,"d":[]}`, nil
	}

	var usernames []string
	usernameMap := make(map[string]bool)
	data := MemberAssocLoginLogData{}
	data.S = pageSize
	data.T = int64(*total.Value)

	for _, v := range searchRes.Hits.Hits {
		for _, hits := range v.InnerHits {
			for _, value := range hits.Hits.Hits {
				log := MemberAssocLoginLog{}
				err = helper.JsonUnmarshal(value.Source, &log)
				if err != nil {
					return "", errors.New(helper.FormatErr)
				}

				data.D = append(data.D, log)

				if _, ok := usernameMap[log.Username]; !ok {
					usernames = append(usernames, log.Username)
				}

			}
		}
	}

	var tags []memberTags

	mCache, err := memberFindBatch(usernames)
	if err != nil {
		return `{"t":0,"d":[]}`, err
	}

	var uids []string
	mpMb := map[string]string{}
	for _, v := range mCache {
		if v.UID != "" {
			uids = append(uids, v.UID)
			mpMb[v.UID] = v.Username
		}
	}

	if len(uids) > 0 {

		t := dialect.From("tbl_member_tags")
		query, _, _ := t.Select("uid", "tag_id", "tag_name").Where(g.Ex{"uid": uids, "prefix": meta.Prefix}).ToSQL()
		err = meta.MerchantDB.Select(&tags, query)
		if err != nil {
			return `{"t":0,"d":[]}`, pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
		}

		mpTags := map[string][]string{}
		for _, v := range tags {
			// 存在追加，不存在新增
			if _, ok := mpTags[v.Uid]; ok {
				mpTags[mpMb[v.Uid]] = append(mpTags[mpMb[v.Uid]], v.TagName)
			} else {
				mpTags[mpMb[v.Uid]] = []string{v.TagName}
			}
		}

		for k := range data.D {
			if tgs, ok := mpTags[data.D[k].Username]; ok {
				data.D[k].Tags = strings.Join(tgs, ",")
			}
		}
	}

	b, err := jettison.Marshal(data)
	if err != nil {
		return "", errors.New(helper.FormatErr)
	}

	return string(b), nil
}
