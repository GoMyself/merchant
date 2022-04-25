package model

import (
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/go-redis/redis/v8"
	"github.com/olivere/elastic/v7"
	"github.com/shopspring/decimal"
	"github.com/valyala/fastjson"
	"merchant2/contrib/helper"
	"net/url"
	"strings"
	"time"
)

// 推广域名统计分析

func TimeToLoc(s, format string, loc *time.Location) (string, error) {

	st, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return "", err
	}

	return st.In(loc).Format(format), nil
}

// 推广域名统计信息
func PromoteInfoList(ty int, urls []string, startTime, endTime string, page, pageSize int) (PromoteData, error) {

	ex := g.Ex{}
	query := elastic.NewBoolQuery()
	depositEx := g.Ex{"state": DepositSuccess}

	if ty != 4 {
		return PromoteData{}, nil
	}

	//rangeParam := map[string][]interface{}{}
	if startTime != "" && endTime != "" {
		startAt, err := helper.TimeToLocMs(startTime, loc) // 毫秒级时间戳
		if err != nil {
			return PromoteData{}, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.TimeToLocMs(endTime, loc) // 毫秒级时间戳
		if err != nil {
			return PromoteData{}, errors.New(helper.TimeTypeErr)
		}

		if startAt >= endAt {
			return PromoteData{}, errors.New(helper.QueryTimeRangeErr)
		}

		startAtStr, _ := TimeToLoc(startTime, TGDateFormat, loc)
		endAtStr, _ := TimeToLoc(endTime, TGDateFormat, loc)
		//es日期查询区间
		query.Must(elastic.NewRangeQuery("time_iso8601").Gte(startAtStr).Lte(endAtStr))

		startAt, _ = helper.TimeToLoc(startTime, loc)
		endAt, _ = helper.TimeToLoc(endTime, loc)
		//mysql日期查询区间
		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
		depositEx["confirm_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
		//rangeParam["created_at"] = []interface{}{startAt, endAt}
	}

	var defUrls []string
	// 默认查询地推域名的数据
	if len(urls) == 0 {

		domains, ok := meta.PromoteConfig["local"]["domains"].([]interface{})
		if ok {
			for _, v := range domains {
				defUrls = append(defUrls, v.(string))
			}
		}
	} else {
		for _, v := range urls {
			uri, _, err := parseURL(v)
			if err != nil {
				continue
			}

			defUrls = append(defUrls, uri)
		}
	}

	//获取redis中域名中代理id
	result := PromoteData{}
	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()

	shouldQuery := elastic.NewBoolQuery()
	cmds := map[string]*redis.SliceCmd{}
	for _, v := range defUrls {

		cmd := pipe.HMGet(ctx, v, "uid", "name")
		cmds[v] = cmd

		//查询域名,采用模糊匹配
		shouldQuery.Should(elastic.NewTermQuery("hostdomain", v))
	}

	query.Must(shouldQuery)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return result, pushLog(err, helper.RedisErr)
	}

	//处理获取的代理id
	var uids []string
	agencyUids := map[string]LinkAgency{}
	for k, v := range cmds {

		if v.Err() != nil {
			continue
		}
		item := LinkAgency{}
		err := v.Scan(&item)
		if err != nil {
			return result, pushLog(err, helper.RedisErr)
		}

		if len(item.Name) == 0 {
			continue
		}

		agencyUids[k] = item
		uids = append(uids, item.UID)
	}

	if len(uids) > 0 {
		ex["parent_uid"] = uids
		depositEx["parent_uid"] = uids
	}

	//构建同代理会员注册数
	var data []UrlAgencyCount
	sql, _, _ := dialect.From("tbl_members").
		Select("parent_uid",
			g.COUNT("uid").As("num"),
			g.SUM("first_deposit_amount").As("first_deposit_amount"),
			g.SUM(
				g.Case().When(g.C("first_deposit_at").Gt(0), 1)).As("first_deposit_num"),
		).Where(ex).GroupBy("parent_uid").ToSQL()

	fmt.Printf("members sql=%s\n", sql)
	err = meta.MerchantDB.Select(&data, sql)
	if err != nil {
		return result, pushLog(fmt.Errorf("%s,[%s]", err.Error(), sql), helper.DBErr)
	}

	sql, _, _ = dialect.From("tbl_deposit").
		Select("parent_uid",
			g.COUNT("id").As("num"),
			g.SUM("amount").As("amount")).
		GroupBy("parent_uid").Where(depositEx).ToSQL()

	fmt.Printf("members sql=%s\n", sql)
	var ddata []depositData
	err = meta.MerchantDB.Select(&ddata, sql)
	if err != nil {
		return result, pushLog(fmt.Errorf("%s,[%s]", err.Error(), sql), helper.DBErr)
	}

	groupIp := "remote_addr"
	//聚合 先按域名聚合访问数，在按ip去重统计
	fieldIpAgg := elastic.NewTermsAggregation().
		Field("hostdomain").Size(100).
		SubAggregation("remote_addr",
			elastic.NewCardinalityAggregation().Field("remote_addr"))

	searchAggs := map[string]*elastic.TermsAggregation{
		groupIp: fieldIpAgg,
	}

	fmt.Println(meta.PromoteConfig)
	index, ok := meta.PromoteConfig["local"]["index"].(string)
	if !ok {
		return result, pushLog(errors.New(helper.ESErr), "es")
	}

	//es查询
	resOrder, flg, err := EsQueryAggTerms(meta.AccessEs, index, query, searchAggs)
	if err != nil {
		return result, pushLog(err, flg)
	}

	//分析es查询结果
	var p fastjson.Parser
	agg, _ := resOrder.Aggregations[groupIp].MarshalJSON()
	buckets, err := p.ParseBytes(agg)
	if err != nil {
		return result, errors.New(helper.FormatErr)
	}

	items, err := buckets.Get("buckets").Array()
	if err != nil {
		return result, errors.New(helper.FormatErr)
	}

	//构建返回body
	var detail []Promote
	for _, v := range items {

		uri := string(v.GetStringBytes("key"))
		item := Promote{
			URL:     uri,
			CallNum: v.GetInt64("doc_count"),
			IpNum:   v.GetInt64("remote_addr", "value"),
			RegNum:  0,
		}

		//获取域名注册数量
		agency, ok := agencyUids[uri]
		if ok {

			item.Username = agency.Name
			item.UID = agency.UID

			for _, vv := range data {
				if vv.ParentUID == agency.UID {
					item.RegNum = vv.Num.Int64
					item.FirstDepositAmount = vv.FirstDepositAmount.Float64
					item.FirstDepositNum = vv.FirstDepositNum.Int64
				}
			}

			for _, vv := range ddata {
				if vv.ParentUID == agency.UID {
					item.DepositNum = vv.Num.Int64
					item.DepositAmount = vv.Amount.Float64
				}
			}
		}

		regNum := decimal.NewFromInt(item.RegNum)
		ipNum := decimal.NewFromInt(item.IpNum)
		item.RegRatio = 0
		if ipNum.Cmp(decimal.Zero) != 0 {
			item.RegRatio, _ = regNum.Div(ipNum).Truncate(4).Float64()
		}

		fdNum := decimal.NewFromInt(item.FirstDepositNum)
		if regNum.Cmp(decimal.Zero) != 0 {
			item.FirstDepositRatio, _ = fdNum.Div(regNum).Truncate(4).Float64()
		}

		detail = append(detail, item)
	}

	unm := len(detail)
	result.T = int64(unm)
	result.D = detail

	return result, nil
}

func parseURL(s string) (string, string, error) {

	fmt.Printf("注册地址 : %s\n", s)

	u, err := url.Parse(s)
	if err != nil {
		return "", "", err
	}

	hosts := strings.SplitN(u.Host, ".", 5)
	n := len(hosts)
	if n < 2 || n > 3 {
		return "", "", errors.New("url error")
	}
	return hosts[n-2] + "." + hosts[n-1], u.Host, nil
}

// 推广域名关联ip列表
func PromoteIPList(url string, startTime, endTime string, page, pageSize int) (TgIpData, error) {

	result := TgIpData{}
	query := elastic.NewBoolQuery()
	if startTime != "" && endTime != "" {
		startAt, err := helper.TimeToLocMs(startTime, loc) // 毫秒级时间戳
		if err != nil {
			return result, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.TimeToLocMs(endTime, loc) // 毫秒级时间戳
		if err != nil {
			return result, errors.New(helper.TimeTypeErr)
		}

		if startAt >= endAt {
			return result, errors.New(helper.QueryTimeRangeErr)
		}

		startAtStr, _ := TimeToLoc(startTime, TGDateFormat, loc)
		endAtStr, _ := TimeToLoc(endTime, TGDateFormat, loc)
		//es日期查询区间
		query.Must(elastic.NewRangeQuery("time_iso8601").Gte(startAtStr).Lte(endAtStr))
	}

	uri, _, err := parseURL(url)
	if err != nil {
		return result, errors.New(helper.URLErr)
	}

	cmd := meta.MerchantRedis.HMGet(ctx, uri, "uid", "name")
	if cmd.Err() != nil {
		return result, nil
	}

	agency := LinkAgency{}
	err = cmd.Scan(&agency)
	if err != nil {
		return result, nil
	}

	index, ok := meta.PromoteConfig["local"]["index"].(string)
	if !ok {
		return result, errors.New(helper.ESErr)
	}

	//查询域名
	query.Must(elastic.NewTermQuery("hostdomain", uri))

	fields := []string{"remote_addr", "hostdomain", "time_iso8601", "request_uri"}
	//	TODO
	total, esResult, _, err := EsQuerySearch(esPrefixIndex(index), "@timestamp", page, pageSize, fields, query, nil)

	if err != nil {
		return result, pushLog(err, helper.ESErr)
	}

	result.S = uint(pageSize)
	result.T = total
	for _, v := range esResult {
		item := TgIp{}
		item.Id = v.Id
		item.UID = agency.UID
		item.Username = agency.Name
		_ = helper.JsonUnmarshal(v.Source, &item)
		result.D = append(result.D, item)

	}

	return result, nil
}

// 推广域名关联会员列表
func PromoteMemberList(url string, startTime, endTime string, page, pageSize uint) (TgMemberData, error) {

	result := TgMemberData{}
	ex := g.Ex{}
	//rangeParam := map[string][]interface{}{}
	if startTime != "" && endTime != "" {
		startAt, err := helper.TimeToLoc(startTime, loc) // 毫秒级时间戳
		if err != nil {
			return result, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.TimeToLoc(endTime, loc) // 毫秒级时间戳
		if err != nil {
			return result, errors.New(helper.TimeTypeErr)
		}

		if startAt >= endAt {
			return result, errors.New(helper.QueryTimeRangeErr)
		}

		//mysql日期查询区间
		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}

	}

	result.S = pageSize
	result.D = []TgMember{}

	uri, _, err := parseURL(url)
	if err != nil {
		return result, errors.New(helper.URLErr)
	}

	cmd := meta.MerchantRedis.HGet(ctx, uri, "uid")
	if err := cmd.Err(); err != nil {
		return result, nil
	}

	ex["parent_uid"] = cmd.Val()
	t := dialect.From("tbl_members")

	if page == 1 {

		var total int64
		sql, _, _ := t.Select(g.COUNT(1).As("t")).Where(ex).ToSQL()
		err := meta.MerchantDB.Get(&total, sql)
		if err != nil {
			return result, pushLog(err, helper.DBErr)
		}

		result.T = total
	}

	offset := (page - 1) * pageSize
	cols := []interface{}{"uid", "username", "parent_uid", "parent_name", "created_at", "regip"}
	sql, _, _ := t.Select(cols...).Where(ex).Offset(offset).Limit(pageSize).ToSQL()
	err = meta.MerchantDB.Select(&result.D, sql)
	if err != nil {
		return result, pushLog(err, helper.DBErr)
	}

	return result, nil
}
