package model

import (
	"database/sql"
	"errors"
	"fmt"
	"merchant2/contrib/helper"
	"strconv"
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

	if ty == TyBankcard {

		var (
			ids []string
		)
		for _, v := range data.D {
			ids = append(ids, v.ID)
		}

		d, err := grpc_t.DecryptAll(ids, true, []string{"bankcard"})
		if err != nil {
			fmt.Println("grpc_t.DecryptAll err = ", err)
			return data, errors.New(helper.GetRPCErr)
		}

		for k, v := range data.D {
			data.D[k].Value = d[v.ID]["bankcard"]
		}
	}

	return data, nil
}

// 黑名单添加
func BlacklistInsert(fctx *fasthttp.RequestCtx, ty int, value string, record g.Record) error {

	key := ""
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
		fmt.Println("BlacklistInsert Exec err = ", err.Error())
		return errors.New(helper.DBErr)
	}

	switch ty {
	case TyIP:
		key = "ip_blacklist"
	case TyPhone:
		key = "phone_blacklist"
	case TyBankcard:
		key = "bankcard_blacklist"
	}

	meta.MerchantRedis.Do(ctx, "CF.ADD", key, value).Err()

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

	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	query, _, _ = dialect.Delete("tbl_blacklist").Where(ex).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	idx := MurmurHash(data.Value, 0) % 10
	key := fmt.Sprintf("bl:dev%d", idx)
	switch data.Ty {
	case TyDevice:
	case TyIP:
		key = fmt.Sprintf("bl:ip%d", idx)
	case TyEmail:
		key = fmt.Sprintf("bl:em%d", idx)
	case TyPhone:
		key = fmt.Sprintf("bl:ph%d", idx)
	case TyBankcard:
		bc, err := BankCardFindOne(ex)
		if err != nil {
			return err
		}

		// 不是删除也不是冻结
		if bc.State != 2 && bc.State != 3 {
			return errors.New(helper.OperateFailed)
		}

		// 冻结状态直接恢复
		if bc.State == 3 {
			query, _, _ = dialect.Update("tbl_member_bankcard").Set(g.Record{"state": 1}).Where(g.Ex{"id": id, "prefix": meta.Prefix}).ToSQL()
			_, err = tx.Exec(query)
			if err != nil {
				_ = tx.Rollback()
				return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
			}
		}

		hash, _ := strconv.ParseUint(data.Value, 10, 64)
		idx = hash % 10
		key = fmt.Sprintf("bl:bc%d", idx)
	case TyVirtualAccount:
		key = fmt.Sprintf("bl:va%d", idx)
	default:
	}

	err = tx.Commit()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	// 删除value set
	_, err = meta.MerchantRedis.SRem(ctx, key, data.Value).Result()
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	return nil
}

// 获取黑名单会员
/*
ty 1 设备号 2 ip
value 设备号或者ip的值
*/
func BlacklistFindUsers(ty int, value string) []string {

	if ty != TyDevice && ty != TyIP {
		return nil
	}

	key := "device_no.keyword"
	if ty == TyIP {
		key = "ips.keyword"
	}
	query := elastic.NewBoolQuery().Must(elastic.NewTermQuery(key, value), elastic.NewTermQuery("prefix", meta.Prefix))
	agg := elastic.NewTermsAggregation().Field("username.keyword").Size(10000).
		SubAggregation("username", elastic.NewCardinalityAggregation().Field("username.keyword"))
	resOrder, err := meta.ES.Search().Index(esPrefixIndex("memberlogin")).
		Query(query).Size(0).Aggregation("group", agg).Do(ctx)
	if err != nil {
		return nil
	}

	terms, ok := resOrder.Aggregations.Terms("group")
	if !ok {
		return nil
	}

	var data []string
	for _, v := range terms.Buckets {

		key, ok := v.Key.(string)
		if !ok {
			continue
		}

		data = append(data, key)
	}

	return data
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

func BlacklistLoadCache() error {

	t := dialect.From("tbl_blacklist")
	for i := TyDevice; i <= TyVirtualAccount; i++ {

		keyFmt := "bl:dev%d"
		switch i {
		case TyDevice:
		case TyIP:
			keyFmt = "bl:ip%d"
		case TyEmail:
			keyFmt = "bl:em%d"
		case TyPhone:
			keyFmt = "bl:ph%d"
		case TyBankcard:
			keyFmt = "bl:bc%d"
		case TyVirtualAccount:
			keyFmt = "bl:va%d"
		default:
		}

		var values []string
		ex := g.Ex{
			"ty":     i,
			"prefix": meta.Prefix,
		}
		query, _, _ := t.Select("value").Where(ex).ToSQL()
		err := meta.MerchantDB.Select(&values, query)
		if err != nil {
			return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
		}

		pipe := meta.MerchantRedis.TxPipeline()
		for _, v := range values {
			idx := MurmurHash(v, 0) % 10
			if i == TyBankcard {
				hash, _ := strconv.ParseUint(v, 10, 64)
				idx = hash % 10
			}
			key := fmt.Sprintf(keyFmt, idx)
			pipe.SAdd(ctx, key, v)
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			return pushLog(err, helper.RedisErr)
		}
	}

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
