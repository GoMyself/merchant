package model

import (
	"errors"
	"github.com/olivere/elastic/v7"
	"merchant2/contrib/helper"
)

func SmsRecordList(page, pageSize int, startTime, endTime string, query *elastic.BoolQuery) (SmsRecordData, error) {

	data := SmsRecordData{}

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
