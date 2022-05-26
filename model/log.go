package model

import (
	"errors"
	"fmt"
	"github.com/olivere/elastic/v7"
	"merchant2/contrib/helper"
)

func AdminLoginLog(start, end string, page, pageSize int, query *elastic.BoolQuery) (AdminLoginLogData, error) {

	data := AdminLoginLogData{}

	if start != "" && end != "" {

		startAt, err := helper.TimeToLoc(start, loc)
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		endAt, err := helper.TimeToLoc(end, loc)
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		query.Filter(
			elastic.NewRangeQuery("created_at").Gte(startAt).Lte(endAt),
		)
	}
	query.Filter(elastic.NewTermQuery("prefix", meta.Prefix))
	fields := []string{"uid", "name", "ip", "device", "flag", "created_at", "prefix"}
	total, result, _, err := EsQuerySearch(esPrefixIndex("admin_login_log"), "@timestamp", page, pageSize, fields, query, nil)
	if err != nil {
		return data, err
	}

	data.T = total
	data.S = pageSize

	for _, v := range result {

		log := adminLoginLog{}
		if err = helper.JsonUnmarshal(v.Source, &log); err != nil {
			return data, errors.New(helper.FormatErr)
		}

		log.Id = v.Id
		data.D = append(data.D, log)
	}

	return data, nil
}

// 系统日志
func SystemLog(start, end string, page, pageSize int, query *elastic.BoolQuery) (SystemLogData, error) {

	//data := SystemLogData{}
	//
	//if start != "" && end != "" {
	//
	//	startAt, err := helper.TimeToLoc(start, loc)
	//	if err != nil {
	//		return data, errors.New(helper.DateTimeErr)
	//	}
	//
	//	endAt, err := helper.TimeToLoc(end, loc)
	//	if err != nil {
	//		return data, errors.New(helper.DateTimeErr)
	//	}
	//
	//	query.Filter(
	//		elastic.NewRangeQuery("created_at").Gte(startAt).Lte(endAt),
	//	)
	//}
	//query.Filter(elastic.NewTermQuery("prefix", meta.Prefix))
	//
	//fields := []string{"uid", "name", "title", "ip", "content", "created_at", "prefix"}
	//total, result, _, err := EsQuerySearch(esPrefixIndex("system_log"), "@timestamp", page, pageSize, fields, query, nil)
	//if err != nil {
	//	return data, err
	//}
	//
	//data.T = total
	//data.S = pageSize
	//
	//for _, v := range result {
	//
	//	log := systemLog{}
	//	if err = helper.JsonUnmarshal(v.Source, &log); err != nil {
	//		return data, errors.New(helper.FormatErr)
	//	}
	//
	//	log.Id = v.Id
	//	data.D = append(data.D, log)
	//}

	return data, nil
}

func EsQuerySearch(index, sortField string, page, pageSize int,
	fields []string, boolQuery *elastic.BoolQuery, agg map[string]*elastic.SumAggregation) (int64, []*elastic.SearchHit, elastic.Aggregations, error) {

	fsc := elastic.NewFetchSourceContext(true).Include(fields...)
	offset := (page - 1) * pageSize
	//打印es查询json
	esService := meta.ES.Search().FetchSourceContext(fsc).Query(boolQuery).From(offset).Size(pageSize).TrackTotalHits(true).Sort(sortField, false)
	for k, v := range agg {
		esService = esService.Aggregation(k, v)
	}
	resOrder, err := esService.Index(index).Do(ctx)
	if err != nil {
		fmt.Println(err)
		return 0, nil, nil, pushLog(err, helper.ESErr)
	}

	if resOrder.Status != 0 || resOrder.Hits.TotalHits.Value <= int64(offset) {
		return resOrder.Hits.TotalHits.Value, nil, nil, nil
	}

	return resOrder.Hits.TotalHits.Value, resOrder.Hits.Hits, resOrder.Aggregations, nil
}
