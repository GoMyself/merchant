package model

import (
	"database/sql"
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/olivere/elastic/v7"
	"github.com/valyala/fastjson"
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"strconv"
	"strings"
)

var (
	betTimeFlags = map[string]string{
		"1": "bet_time",
		"2": "settle_time",
		"3": "start_time",
	}
)

type GameGroupData struct {
	Agg map[string]string   `json:"agg"`
	D   []map[string]string `json:"d"`
	T   int                 `json:"t"`
	S   int                 `json:"s"`
}

func RecordTransaction(page, pageSize int, startTime, endTime, table string, ex g.Ex) (TransactionData, error) {

	data := TransactionData{}
	ex["prefix"] = meta.Prefix
	if startTime != "" && endTime != "" {
		startAt, err := helper.TimeToLocMs(startTime, loc) // 毫秒级时间戳
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.TimeToLocMs(endTime, loc) // 毫秒级时间戳
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}

	t := dialect.From(table)
	if page == 1 {
		query, _, _ := t.Select(g.COUNT("id")).Where(ex).ToSQL()
		err := meta.MerchantDB.Get(&data.T, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}

		if data.T == 0 {
			return data, nil
		}
	}

	offset := pageSize * (page - 1)
	query, _, _ := t.Select(g.SUM("amount").As("agg")).Where(ex).ToSQL()
	err := meta.MerchantDB.Get(&data.Agg, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	query, _, _ = t.Select(colsTransaction...).Where(ex).
		Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("created_at").Desc()).ToSQL()
	err = meta.MerchantDB.Select(&data.D, query)
	if err != nil && err != sql.ErrNoRows {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func RecordTransfer(page, pageSize int, startTime, endTime string, ex g.Ex) (TransferData, error) {

	data := TransferData{}
	if startTime != "" && endTime != "" {
		//判断日期
		startAt, err := helper.TimeToLocMs(startTime, loc) // 毫秒级时间戳
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}
		endAt, err := helper.TimeToLocMs(endTime, loc) // 毫秒级时间戳
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}
	ex["prefix"] = meta.Prefix
	t := dialect.From("tbl_member_transfer")
	if page == 1 {
		query, _, _ := t.Select(g.COUNT("id")).Where(ex).ToSQL()
		err := meta.MerchantDB.Get(&data.T, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}

		if data.T == 0 {
			return data, nil
		}

		query, _, _ = t.Select(g.SUM("amount").As("agg")).Where(ex).ToSQL()
		err = meta.MerchantDB.Get(&data.Agg, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}
	}

	offset := pageSize * (page - 1)
	query, _, _ := t.Select(colsTransfer...).Where(ex).
		Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("created_at").Desc()).ToSQL()
	err := meta.MerchantDB.Select(&data.D, query)
	if err != nil && err != sql.ErrNoRows {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func Game(ty, pageSize, page int, params map[string]string) (GameRecordData, error) {

	data := GameRecordData{}
	//判断日期
	startAt, err := helper.TimeToLocMs(params["start_time"], loc) // 毫秒级时间戳
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}
	endAt, err := helper.TimeToLocMs(params["end_time"], loc) // 毫秒级时间戳
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	if startAt >= endAt {
		return data, errors.New(helper.QueryTimeRangeErr)
	}

	aggParam := map[string]string{
		"bet_amount_agg":       "bet_amount",
		"net_amount_agg":       "net_amount",
		"valid_bet_amount_agg": "valid_bet_amount",
	}

	if ty == GameTyValid {

		param := map[string]interface{}{
			"name":   params["username"],
			"flag":   "1",
			"prefix": meta.Prefix,
		}

		rangeParam := map[string][]interface{}{
			"bet_time": {startAt, endAt},
		}

		data, err = recordGameESQuery(pullPrefixIndex("tbl_game_record"), "bet_time", page, pageSize, param, rangeParam, aggParam)
		if err != nil {
			return data, err
		}

		return data, nil
	}
	//查询条件
	param := map[string]interface{}{
		"prefix": meta.Prefix,
	}

	if params["pid"] != "" {
		if strings.Contains(params["pid"], ",") {
			pids := strings.Split(params["pid"], ",")

			var ids []interface{}
			for _, v := range pids {
				if validator.CtypeDigit(v) {
					ids = append(ids, v)
				}
			}

			param["api_type"] = ids
		}

		if !strings.Contains(params["pid"], ",") {
			if validator.CtypeDigit(params["pid"]) {
				param["api_type"] = params["pid"]
			}
		}
	}

	if params["flag"] != "" {
		param["flag"] = params["flag"]
	}

	rangeField := ""
	if params["time_flag"] != "" {
		rangeField = betTimeFlags[params["time_flag"]]
	}

	if rangeField == "" {
		return data, errors.New(helper.QueryTermsErr)
	}

	rangeParam := map[string][]interface{}{
		rangeField: {startAt, endAt},
	}

	if ty == GameMemberWinOrLose {

		param["name"] = params["username"]
		data, err = recordGameESQuery(pullPrefixIndex("tbl_game_record"), rangeField, page, pageSize, param, rangeParam, aggParam)
		if err != nil {
			return data, err
		}

		return data, nil
	}

	if !validator.CtypeDigit(params["time_flag"]) {
		return data, errors.New(helper.QueryTermsErr)
	}

	if params["bet_min"] == "" && params["bet_max"] != "" {
		max, _ := strconv.ParseFloat(params["bet_max"], 64)
		rangeParam["bet_amount"] = []interface{}{nil, max}
	}

	if params["bet_min"] != "" && params["bet_max"] == "" {
		min, _ := strconv.ParseFloat(params["bet_min"], 64)
		rangeParam["bet_amount"] = []interface{}{min, nil}
	}

	if params["bet_min"] != "" && params["bet_max"] != "" {
		min, _ := strconv.ParseFloat(params["bet_min"], 64)
		max, _ := strconv.ParseFloat(params["bet_max"], 64)
		if max < min {
			return data, errors.New(helper.BetAmountRangeErr)
		}

		rangeParam["bet_amount"] = []interface{}{min, max}
	}

	if params["uid"] != "" {
		param["uid"] = params["uid"]
	}

	if params["plat_type"] != "" {
		param["game_type"] = params["plat_type"]
	}

	if params["game_name"] != "" {
		param["game_name"] = params["game_name"]
	}

	if params["username"] != "" {
		param["name"] = params["username"]
	}

	if params["bill_no"] != "" {
		param["bill_no"] = params["bill_no"]
	}

	if params["pre_settle"] != "" {
		early, _ := strconv.Atoi(params["pre_settle"])
		param["presettle"] = early
	}

	if params["resettle"] != "" {
		second, _ := strconv.Atoi(params["resettle"])
		param["resettle"] = second
	}

	if params["parent_name"] != "" {
		param["parent_name"] = params["parent_name"]
	}

	if params["top_name"] != "" {
		param["top_name"] = params["top_name"]
	}

	data, err = recordGameESQuery(pullPrefixIndex("tbl_game_record"), rangeField, page, pageSize, param, rangeParam, aggParam)
	if err != nil {
		return data, err
	}

	return data, nil
}

func GameGroup(ty, pageSize, page int, params map[string]string) (GameGroupData, error) {

	data := GameGroupData{}
	//判断日期
	startAt, err := helper.TimeToLocMs(params["start_time"], loc) // 毫秒级时间戳
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}
	endAt, err := helper.TimeToLocMs(params["end_time"], loc) // 毫秒级时间戳
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	if startAt >= endAt {
		return data, errors.New(helper.QueryTimeRangeErr)
	}

	aggParam := map[string]string{
		"bet_amount_agg":       "bet_amount",
		"net_amount_agg":       "net_amount",
		"valid_bet_amount_agg": "valid_bet_amount",
	}
	//查询条件
	param := map[string]interface{}{
		"prefix": meta.Prefix,
	}
	if params["pid"] != "" {
		if strings.Contains(params["pid"], ",") {
			pids := strings.Split(params["pid"], ",")

			var ids []interface{}
			for _, v := range pids {
				if validator.CtypeDigit(v) {
					ids = append(ids, v)
				}
			}

			param["api_type"] = ids
		}

		if !strings.Contains(params["pid"], ",") {
			if validator.CtypeDigit(params["pid"]) {
				param["api_type"] = params["pid"]
			}
		}
	}

	if params["flag"] != "" {
		param["flag"] = params["flag"]
	}

	rangeField := ""
	if params["time_flag"] != "" {
		rangeField = betTimeFlags[params["time_flag"]]
	}

	if rangeField == "" {
		return data, errors.New(helper.QueryTermsErr)
	}

	rangeParam := map[string][]interface{}{
		rangeField: {startAt, endAt},
	}
	if ty == GameMemberDayGroup {
		param["name"] = params["username"]
		other := map[string]string{
			"index":           "tbl_game_record",
			"range_field":     rangeField,
			"agg_group_field": rangeField,
			"interal":         "1d",
		}

		return groupByEs(page, pageSize, other, param, rangeParam, aggParam)
	}

	if ty == GameMemberTransferGroup {
		param["name"] = params["username"]

		other := map[string]string{
			"index":           "tbl_game_record",
			"range_field":     rangeField,
			"agg_group_field": "api_type",
			"interal":         "",
		}

		return groupByEs(page, pageSize, other, param, rangeParam, aggParam)
	}

	return GameGroupData{}, errors.New("ty error")
}

func groupByEs(page, pageSize int, other map[string]string, param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (GameGroupData, error) {

	boolQuery := elastic.NewBoolQuery()
	terms := make([]elastic.Query, 0)
	filters := make([]elastic.Query, 0)

	if len(rangeParam) > 0 {
		for k, v := range rangeParam {
			if v == nil {
				continue
			}

			if len(v) == 2 {
				filters = append(filters, elastic.NewRangeQuery(k).Gte(v[0]).Lte(v[1]))
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
	offset := (page - 1) * pageSize
	//打印es查询json
	esService := meta.ES.Search().Query(boolQuery).TrackTotalHits(true).Size(0).Sort(other["range_field"], false)

	if len(other["interal"]) > 0 {
		//"1d"
		timeAgg := elastic.NewDateHistogramAggregation().Field(other["agg_group_field"]).FixedInterval(other["interal"]).MinDocCount(1) //.Keyed(true) //.Format("yyyy-MM-dd")
		timeAgg.SubAggregation("total", elastic.NewValueCountAggregation().Field(other["agg_group_field"]))
		// 聚合条件
		if len(aggField) > 0 {
			for k, v := range aggField {
				timeAgg = timeAgg.SubAggregation(k, elastic.NewSumAggregation().Field(v))
			}
		}
		esService = esService.Aggregation(other["agg_group_field"], timeAgg)
	}

	if len(other["interal"]) == 0 {
		fieldAgg := elastic.NewTermsAggregation().Field(other["agg_group_field"]).Size(1000)
		fieldAgg.SubAggregation("total", elastic.NewValueCountAggregation().Field(other["agg_group_field"]))
		// 聚合条件
		if len(aggField) > 0 {
			for k, v := range aggField {
				fieldAgg = fieldAgg.SubAggregation(k, elastic.NewSumAggregation().Field(v))
			}
		}

		esService = esService.Aggregation(other["agg_group_field"], fieldAgg)
	}

	resOrder, err := esService.Index(pullPrefixIndex(other["index"])).Do(ctx)
	if err != nil {
		return GameGroupData{}, pushLog(err, helper.ESErr)
	}

	if resOrder.Status != 0 {
		return GameGroupData{}, errors.New(helper.RecordNotExistErr)
	}

	agg, _ := resOrder.Aggregations[other["agg_group_field"]].MarshalJSON()

	var p fastjson.Parser
	buckets, err := p.ParseBytes(agg)
	if err != nil {
		return GameGroupData{}, pushLog(err, helper.ESErr)
	}

	var data []map[string]string
	bucketItems, err := buckets.Get("buckets").Array()
	if err != nil {
		return GameGroupData{}, pushLog(err, helper.ESErr)
	}

	var count uint64
	total := map[string]float64{}

	for k, v := range bucketItems {
		item := map[string]string{}

		item["key"] = fmt.Sprintf("%d", v.GetUint64("key"))
		count += v.GetUint64("total", "value")
		item["total"] = fmt.Sprintf("%d", v.GetUint64("total", "value"))

		for ak, av := range aggField {
			if _, ok := total[av]; !ok {
				total[av] = 0
			}

			total[av] += v.GetFloat64(ak, "value")
			item[av] = fmt.Sprintf("%0.4f", v.GetFloat64(ak, "value"))
		}

		if k >= offset {
			data = append(data, item)
		}
	}

	totalData := map[string]string{"total": fmt.Sprintf("%d", count)}
	for k, v := range total {
		totalData[k] = fmt.Sprintf("%0.4f", v)
	}

	result := GameGroupData{
		Agg: totalData,
		D:   data,
		T:   len(bucketItems),
		S:   pageSize,
	}

	return result, nil
}

func recordGameESQuery(index, sortField string, page, pageSize int,
	param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (GameRecordData, error) {

	data := GameRecordData{Agg: map[string]string{}}
	total, esData, aggData, err := esSearch(index, sortField, page, pageSize, gameRecordFields, param, rangeParam, aggField)
	if err != nil {
		return data, err
	}

	for k, v := range aggField {
		amount, _ := aggData.Sum(k)
		if amount != nil {
			data.Agg[v] = fmt.Sprintf("%.4f", *amount.Value)
		}
	}

	data.T = total
	for _, v := range esData {
		record := GameRecord{}
		record.ID = v.Id
		_ = helper.JsonUnmarshal(v.Source, &record)
		record.ApiTypes = fmt.Sprintf("%d", record.ApiType)
		data.D = append(data.D, record)
	}

	return data, nil
}
