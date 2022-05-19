package model

import (
	"database/sql"
	"errors"
	"fmt"
	"merchant2/contrib/helper"

	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"

	"github.com/olivere/elastic/v7"
)

func MemberRemarkLogList(uid, adminName, startTime, endTime string, page, pageSize int) (MemberRemarkLogData, error) {

	ex := g.Ex{}

	if uid != "" {
		ex["uid"] = uid
	}

	if adminName != "" {
		ex["created_name"] = adminName
	}

	data := MemberRemarkLogData{}

	if len(ex) == 0 && (startTime == "" || endTime == "") {
		return data, errors.New(helper.QueryTermsErr)
	}
	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		endAt, err := helper.TimeToLoc(endTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}
	ex["prefix"] = meta.Prefix

	t := dialect.From("member_remarks_log")

	if page == 1 {
		query, _, _ := t.Select(g.COUNT("*")).Where(ex).ToSQL()

		fmt.Println(query)

		err := meta.MerchantTD.Get(&data.T, query)
		if err == sql.ErrNoRows {
			return data, nil
		}

		if err != nil {
			fmt.Println("Member Remarks Log err = ", err.Error())
			fmt.Println("Member Remarks Log query = ", query)
			body := fmt.Errorf("%s,[%s]", err.Error(), query)
			return data, pushLog(body, helper.DBErr)
		}
		if data.T == 0 {
			return data, nil
		}
	}

	offset := (page - 1) * pageSize
	query, _, _ := t.Select("id", "uid", "username", "msg", "file", "created_name", "created_at", "prefix").Where(ex).Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("ts").Desc()).ToSQL()
	fmt.Println("Member Remarks Log query = ", query)

	err := meta.MerchantTD.Select(&data.D, query)
	if err != nil {
		fmt.Println("Member Remarks Log err = ", err.Error())
		fmt.Println("Member Remarks Log query = ", query)
		body := fmt.Errorf("%s,[%s]", err.Error(), query)
		return data, pushLog(body, helper.DBErr)
	}

	data.S = pageSize

	return data, nil
}

func MemberLoginLogList(startTime, endTime string, page, pageSize int, ex g.Ex) (MemberLoginLogData, error) {

	data := MemberLoginLogData{}
	if len(ex) == 0 && (startTime == "" || endTime == "") {
		return data, errors.New(helper.QueryTermsErr)
	}
	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		endAt, err := helper.TimeToLoc(endTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		ex["create_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}
	ex["prefix"] = meta.Prefix

	t := dialect.From("member_login_log")
	fmt.Println(ex)
	if page == 1 {
		query, _, _ := t.Select(g.COUNT("*")).Where(ex).ToSQL()
		err := meta.MerchantTD.Get(&data.T, query)
		if err == sql.ErrNoRows {
			return data, nil
		}

		if err != nil {
			fmt.Println("Member Login Log err = ", err.Error())
			fmt.Println("Member Login Log query = ", query)
			body := fmt.Errorf("%s,[%s]", err.Error(), query)
			return data, pushLog(body, helper.DBErr)
		}
		if data.T == 0 {
			return data, nil
		}
	}
	offset := (page - 1) * pageSize
	query, _, _ := t.Select("username", "ip", "device", "device_no", "top_name", "parent_name", "create_at").Where(ex).Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("ts").Desc()).ToSQL()
	fmt.Println("Member Login Log query = ", query)

	err := meta.MerchantTD.Select(&data.D, query)
	if err != nil {
		fmt.Println("Member Login Log err = ", err.Error())
		fmt.Println("Member Login Log query = ", query)
		body := fmt.Errorf("%s,[%s]", err.Error(), query)
		return data, pushLog(body, helper.DBErr)
	}

	data.S = pageSize
	return data, nil
}

//func MemberLoginLogList(startTime, endTime string, page, pageSize int, param map[string]interface{}) (MemberLoginLogData, error) {
//
//	data := MemberLoginLogData{}
//	rangeParam := make(map[string][]interface{})
//	if startTime != "" && endTime != "" {
//
//		startAt, err := helper.TimeToLoc(startTime, loc)
//		if err != nil {
//			return data, errors.New(helper.DateTimeErr)
//		}
//
//		endAt, err := helper.TimeToLoc(endTime, loc)
//		if err != nil {
//			return data, errors.New(helper.DateTimeErr)
//		}
//
//		if startAt >= endAt {
//			return data, errors.New(helper.QueryTimeRangeErr)
//		}
//
//		rangeParam["date"] = []interface{}{startAt, endAt}
//	}
//
//	result, err := memberLoginLogList(meta.EsPrefix, page, pageSize, param, rangeParam)
//	if err != nil {
//		return data, err
//	}
//
//	err = helper.JsonUnmarshal([]byte(result), &data)
//	if err != nil {
//		return data, errors.New(helper.FormatErr)
//	}
//
//	// var agencyNames []string
//	// for _, log := range data.D {
//	// 	if log.Parents != "" && log.Parents != "root" {
//	// 		agencyNames = append(agencyNames, log.Parents)
//	// 	}
//	// }
//
//	//riskMap, err := AgencyIsRiskNameMap(agencyNames)
//	//if err != nil {
//	//	return data, err
//	//}
//	//
//	//for i, log := range data.D {
//	//	if risk, ok := riskMap[log.Parents]; ok {
//	//		data.D[i].IsRisk = risk
//	//	}
//	//}
//
//	return data, nil
//}

//func memberRemarkLogList(startTime, endTime string, page, pageSize int, ex g.Ex) (MemberRemarkLogData, error) {
//
//	data := MemberRemarkLogData{}
//
//	if len(ex) == 0 && (startTime == "" || endTime == "") {
//		return data, errors.New(helper.QueryTermsErr)
//	}
//	if startTime != "" && endTime != "" {
//
//		startAt, err := helper.TimeToLoc(startTime, loc)
//		if err != nil {
//			return data, errors.New(helper.DateTimeErr)
//		}
//
//		endAt, err := helper.TimeToLoc(endTime, loc)
//		if err != nil {
//			return data, errors.New(helper.TimeTypeErr)
//		}
//
//		if startAt >= endAt {
//			return data, errors.New(helper.QueryTimeRangeErr)
//		}
//
//		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
//	}
//	ex["prefix"] = meta.Prefix
//
//	t := dialect.From("member_remarks_log")
//
//	if page == 1 {
//		query, _, _ := t.Select(g.COUNT("*")).Where(ex).ToSQL()
//		err := meta.MerchantTD.Get(&data.T, query)
//		if err == sql.ErrNoRows {
//			return data, nil
//		}
//
//		if err != nil {
//			fmt.Println("Member Remarks Log err = ", err.Error())
//			fmt.Println("Member Remarks Log query = ", query)
//			body := fmt.Errorf("%s,[%s]", err.Error(), query)
//			return data, pushLog(body, helper.DBErr)
//		}
//		if data.T == 0 {
//			return data, nil
//		}
//	}
//
//	offset := (page - 1) * pageSize
//	query, _, _ := t.Select("id", "uid", "username", "msg", "file", "created_name", "created_at").Where(ex).Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("ts").Desc()).ToSQL()
//	fmt.Println("Member Remarks Log query = ", query)
//
//	err := meta.MerchantTD.Select(&data.D, query)
//	if err != nil {
//		fmt.Println("Member Remarks Log err = ", err.Error())
//		fmt.Println("Member Remarks Log query = ", query)
//		body := fmt.Errorf("%s,[%s]", err.Error(), query)
//		return data, pushLog(body, helper.DBErr)
//	}
//
//	data.S = pageSize
//
//	return data, nil
//}

//func memberLoginLogList(esPrefix string, page, pageSize int, param map[string]interface{}, rangeParam map[string][]interface{}) (string, error) {
//
//	fields := []string{"username", "ips", "device", "device_no", "date", "parents"}
//	param["prefix"] = meta.Prefix
//	total, esData, _, err := esSearch(esPrefix+"memberlogin", "date", page, pageSize, fields, param, rangeParam, map[string]string{})
//	if err != nil {
//		return `{"t":0,"d":[]}`, err
//	}
//
//	data := MemberLoginLogData{}
//	data.S = pageSize
//	data.T = total
//	for _, v := range esData {
//		log := MemberLoginLog{}
//		_ = helper.JsonUnmarshal(v.Source, &log)
//		data.D = append(data.D, log)
//	}
//
//	b, err := jettison.Marshal(data)
//	if err != nil {
//		return "", errors.New(helper.FormatErr)
//	}
//
//	return string(b), nil
//}

//ES查询转账记录
func esSearch(index, sortField string, page, pageSize int, fields []string,
	param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (int64, []*elastic.SearchHit, elastic.Aggregations, error) {

	boolQuery := elastic.NewBoolQuery()
	terms := make([]elastic.Query, 0)
	filters := make([]elastic.Query, 0)

	if len(rangeParam) > 0 {
		for k, v := range rangeParam {
			if v == nil {
				continue
			}

			if len(v) == 2 {

				if v[0] == nil && v[1] == nil {
					continue
				}
				if val, ok := v[0].(string); ok {
					switch val {
					case "gt":
						rg := elastic.NewRangeQuery(k).Gt(v[1])
						filters = append(filters, rg)
					case "gte":
						rg := elastic.NewRangeQuery(k).Gte(v[1])
						filters = append(filters, rg)
					case "lt":
						rg := elastic.NewRangeQuery(k).Lt(v[1])
						filters = append(filters, rg)
					case "lte":
						rg := elastic.NewRangeQuery(k).Lte(v[1])
						filters = append(filters, rg)
					}
					continue
				}

				rg := elastic.NewRangeQuery(k).Gte(v[0]).Lte(v[1])
				if v[0] == nil {
					rg.IncludeLower(false)
				}

				if v[1] == nil {
					rg.IncludeUpper(false)
				}

				filters = append(filters, rg)
			}
		}
	}

	if len(param) > 0 {
		for k, v := range param {
			if v == nil {
				continue
			}

			if vv, ok := v.([]interface{}); ok {
				filters = append(filters, elastic.NewTermsQuery(k, vv...))
				continue
			}

			terms = append(terms, elastic.NewTermQuery(k, v))
		}
	}

	boolQuery.Filter(filters...)
	boolQuery.Must(terms...)
	fsc := elastic.NewFetchSourceContext(true).Include(fields...)
	offset := (page - 1) * pageSize
	//打印es查询json
	esService := meta.ES.Search().FetchSourceContext(fsc).Query(boolQuery).From(offset).Size(pageSize).TrackTotalHits(true).Sort(sortField, false)

	// 聚合条件
	if len(aggField) > 0 {
		for k, v := range aggField {
			esService = esService.Aggregation(k, elastic.NewSumAggregation().Field(v))
		}
	}

	resOrder, err := esService.Index(index).Do(ctx)
	if err != nil {
		return 0, nil, nil, pushLog(err, helper.ESErr)
	}

	if resOrder.Status != 0 || resOrder.Hits.TotalHits.Value <= int64(offset) {
		return resOrder.Hits.TotalHits.Value, nil, nil, nil
	}

	return resOrder.Hits.TotalHits.Value, resOrder.Hits.Hits, resOrder.Aggregations, nil
}

func EsQueryAggTerms(esCli *elastic.Client, index string, boolQuery *elastic.BoolQuery, agg map[string]*elastic.TermsAggregation) (*elastic.SearchResult, string, error) {

	fsc := elastic.NewFetchSourceContext(true)

	//打印es查询json
	esService := esCli.Search().FetchSourceContext(fsc).Query(boolQuery).Size(0)
	for k, v := range agg {
		esService = esService.Aggregation(k, v)
	}
	resOrder, err := esService.Index(index).Do(ctx)
	if err != nil {
		fmt.Println(err)
		return nil, "es", err
	}
	return resOrder, "", nil
}
