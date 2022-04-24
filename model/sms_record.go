package model

import (
	"errors"
	"github.com/olivere/elastic/v7"
	"merchant2/contrib/helper"
)

func SmsRecordList(page, pageSize int, startTime, endTime string, query *elastic.BoolQuery) (SmsRecordData, error) {

	data := SmsRecordData{}
	//// 没有查询条件  startTime endTime 必填
	//if len(ex) == 0 && (startTime == "" || endTime == "") {
	//	return data, errors.New(helper.QueryTermsErr)
	//}
	//
	//if startTime != "" && endTime != "" {
	//
	//	startAt, err := helper.TimeToLocMs(startTime, loc)
	//	if err != nil {
	//		return data, errors.New(helper.DateTimeErr)
	//	}
	//
	//	endAt, err := helper.TimeToLocMs(endTime, loc)
	//	if err != nil {
	//		return data, errors.New(helper.TimeTypeErr)
	//	}
	//
	//	if startAt >= endAt {
	//		return data, errors.New(helper.QueryTimeRangeErr)
	//	}
	//
	//	ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	//}
	//
	//ex["prefix"] = meta.Prefix
	//t := dialect.From("tbl_sms_record")
	//if page == 1 {
	//	query, _, _ := t.Select(g.COUNT("id")).Where(ex).ToSQL()
	//	err := meta.MerchantDB.Get(&data.T, query)
	//	if err != nil {
	//		return data, pushLog(err, helper.DBErr)
	//	}
	//
	//	if data.T == 0 {
	//		return data, nil
	//	}
	//}
	//
	//offset := pageSize * (page - 1)
	//query, _, _ := t.Select(colsDividend...).Where(ex).
	//	Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("created_at").Desc()).ToSQL()
	//err := meta.MerchantDB.Select(&data.D, query)
	//if err != nil {
	//	return data, pushLog(err, helper.DBErr)
	//}
	//
	//return data, nil

	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.TimeToLoc(endTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		query.Filter(elastic.NewRangeQuery("create_at").Gte(startAt).Lte(endAt))
	}

	t, esResult, _, err := EsQuerySearch(
		esPrefixIndex("sms_log"), "create_at", page, pageSize, loginLogFields, query, nil)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	data.T = t
	for _, v := range esResult {

		log := SmsRecord{}
		_ = helper.JsonUnmarshal(v.Source, &log)
		data.D = append(data.D, log)
	}

	return data, nil
}
