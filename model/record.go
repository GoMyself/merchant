package model

import (
	"database/sql"
	"errors"
	"fmt"
	"merchant/contrib/helper"
	"merchant/contrib/validator"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"

	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/olivere/elastic/v7"
	"github.com/valyala/fastjson"
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

func RecordTransaction(page, pageSize int, startTime, endTime string, ex g.Ex) (TransactionData, error) {

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

	t := dialect.From("tbl_balance_transaction")
	if page == 1 {
		query, _, _ := t.Select(g.COUNT("id")).Where(ex).ToSQL()
		fmt.Println(query)
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
	//fmt.Println(query)
	err := meta.MerchantDB.Get(&data.Agg, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	query, _, _ = t.Select(colsTransaction...).Where(ex).
		Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("created_at").Desc()).ToSQL()
	fmt.Println(query)
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
		fmt.Println(query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}

		if data.T == 0 {
			return data, nil
		}

		query, _, _ = t.Select(g.SUM("amount").As("agg")).Where(ex).ToSQL()
		err = meta.MerchantDB.Get(&data.Agg, query)
		fmt.Println(query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}
	}

	offset := pageSize * (page - 1)
	query, _, _ := t.Select(colsTransfer...).Where(ex).
		Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("created_at").Desc()).ToSQL()
	fmt.Println(query)
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
		"rebate_amount":        "rebate_amount",
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

		data, err = recordGameESQuery(pullPrefixIndex("tbl_game_record"), "bet_time", false, page, pageSize, param, rangeParam, aggParam)
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
		data, err = recordGameESQuery(pullPrefixIndex("tbl_game_record"), rangeField, false, page, pageSize, param, rangeParam, aggParam)
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

	data, err = recordGameESQuery(pullPrefixIndex("tbl_game_record"), rangeField, false, page, pageSize, param, rangeParam, aggParam)
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
		"rebate_amount":        "rebate_amount",
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
	fmt.Println("boolQuery:", boolQuery)
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

func recordGameESQuery(index, sortField string, ascending bool, page, pageSize int,
	param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (GameRecordData, error) {

	data := GameRecordData{Agg: map[string]string{}}
	param["tester"] = "1"
	fmt.Println("param:", param)
	total, esData, aggData, err := esSearch(index, sortField, ascending, page, pageSize, gameRecordFields, param, rangeParam, aggField)
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

func RecordAdminGame(flag, startTime, endTime string, page, pageSize int, query *elastic.BoolQuery) (GameRecordData, error) {

	data := GameRecordData{}

	startAt, err := helper.TimeToLocMs(startTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	endAt, err := helper.TimeToLocMs(endTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	if startAt >= endAt {
		return data, errors.New(helper.QueryTimeRangeErr)
	}

	query.Filter(elastic.NewTermQuery("prefix", meta.Prefix),
		elastic.NewRangeQuery(betTimeFlags[flag]).Gte(startAt).Lte(endAt))

	t, esResult, _, err := EsQuerySearch(pullPrefixIndex("tbl_game_record"), "bet_time", page, pageSize, gameRecordFields, query, nil)
	if err != nil {
		return data, err
	}

	data.T = t
	var names []string
	for _, v := range esResult {
		record := GameRecord{}
		record.ID = v.Id
		_ = helper.JsonUnmarshal(v.Source, &record)
		record.ApiTypes = fmt.Sprintf("%d", record.ApiType)
		data.D = append(data.D, record)
		names = append(names, record.ParentName)
	}

	return data, nil
}

func RecordLoginLog(page, pageSize int, startTime, endTime string, query *elastic.BoolQuery) (MemberLoginLogData, error) {

	data := MemberLoginLogData{}
	/*
		if startTime != "" && endTime != "" {

			startAt, err := helper.TimeToLoc(startTime, loc)
			if err != nil {
				return data, errors.New(helper.TimeTypeErr)
			}

			endAt, err := helper.TimeToLoc(endTime, loc)
			if err != nil {
				return data, errors.New(helper.TimeTypeErr)
			}

			query.Filter(elastic.NewRangeQuery("date").Gte(startAt).Lte(endAt))
		}

		t, esResult, _, err := EsQuerySearch(
			esPrefixIndex("memberlogin"), "date", page, pageSize, loginLogFields, query, nil)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}

		var names []string
		data.S = pageSize
		data.T = t
		for _, v := range esResult {

			log := MemberLoginLog{}
			_ = helper.JsonUnmarshal(v.Source, &log)
			data.D = append(data.D, log)
			names = append(names, log.Parents)
		}
	*/
	return data, nil
}

func RecordDeposit(page, pageSize int, startTime, endTime string, query *elastic.BoolQuery) (FDepositData, error) {

	data := FDepositData{}
	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.TimeToLoc(endTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		query.Filter(elastic.NewRangeQuery("created_at").Gte(startAt).Lte(endAt))
	}

	query.Filter(elastic.NewTermQuery("prefix", meta.Prefix))
	t, esResult, _, err := EsQuerySearch(
		esPrefixIndex("tbl_deposit"), "created_at", page, pageSize, depositFields, query, nil)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	var names []string
	data.T = t
	for _, v := range esResult {

		record := Deposit{}
		_ = helper.JsonUnmarshal(v.Source, &record)
		record.ID = v.Id
		data.D = append(data.D, record)
		names = append(names, record.ParentName)
	}

	return data, nil
}

func RecordDividend(page, pageSize int, startTime, endTime string, query *elastic.BoolQuery) (DividendEsData, error) {

	data := DividendEsData{}
	query.Filter(elastic.NewTermQuery("state", DividendReviewPass))

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

		query.Filter(elastic.NewRangeQuery("review_at").Gte(startAt).Lte(endAt))
	}

	query.Filter(elastic.NewTermQuery("prefix", meta.Prefix))
	t, esResult, _, err := EsQuerySearch(
		esPrefixIndex("tbl_member_dividend"), "review_at", page, pageSize, dividendFields, query, nil)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	var names []string
	data.T = t
	for _, v := range esResult {

		record := Dividend{}
		//fmt.Println(string(v.Source))
		_ = helper.JsonUnmarshal(v.Source, &record)
		record.ID = v.Id
		//fmt.Println(record)
		data.D = append(data.D, record)
		names = append(names, record.ParentName)
	}

	return data, nil
}

func RecordRebate(page, pageSize int, startTime, endTime string, ex g.Ex) (RebateData, error) {

	data := RebateData{}
	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
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
	t := dialect.From("tbl_commission_transaction")
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
	query, _, _ := t.Select(colsCommissionTransfer...).Where(ex).
		Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("created_at").Desc()).ToSQL()
	err := meta.MerchantDB.Select(&data.D, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func RecordAdjust(page, pageSize int, startTime, endTime string, query *elastic.BoolQuery) (AdjustData, error) {

	data := AdjustData{}
	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.TimeToLoc(endTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		query.Filter(elastic.NewRangeQuery("review_at").Gte(startAt).Lte(endAt))
	}

	query.Filter(elastic.NewTermQuery("prefix", meta.Prefix))
	t, esResult, _, err := EsQuerySearch(
		esPrefixIndex("tbl_member_adjust"), "apply_at", page, pageSize, adjustFields, query, nil)
	if err != nil {
		return data, err
	}

	data.T = t
	var names []string
	for _, v := range esResult {

		record := MemberAdjust{}
		_ = helper.JsonUnmarshal(v.Source, &record)
		record.ID = v.Id
		data.D = append(data.D, record)
		names = append(names, record.ParentName)
	}

	return data, nil
}

// 代理管理-记录管理-提款
func RecordWithdraw(page, pageSize int, startTime, endTime, applyStartTime, applyEndTime string, query *elastic.BoolQuery) (FWithdrawData, error) {

	data := FWithdrawData{}
	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.TimeToLoc(endTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		query.Filter(elastic.NewRangeQuery("withdraw_at").Gte(startAt).Lte(endAt))
	}

	if applyStartTime != "" && applyEndTime != "" {

		startAt, err := helper.TimeToLoc(applyStartTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.TimeToLoc(applyEndTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		query.Filter(elastic.NewRangeQuery("created_at").Gte(startAt).Lte(endAt))
	}

	query.Filter(elastic.NewTermQuery("prefix", meta.Prefix))
	t, esResult, _, err := EsQuerySearch(
		esPrefixIndex("tbl_withdraw"), "created_at", page, pageSize, withdrawFields, query, nil)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	data.T = t
	for _, v := range esResult {

		record := Withdraw{}
		_ = helper.JsonUnmarshal(v.Source, &record)
		record.ID = v.Id
		data.D = append(data.D, record)
	}

	return data, nil
}

// 处理 提款订单返回数据
func WithdrawDealListData(data FWithdrawData) (WithdrawListData, error) {

	result := WithdrawListData{
		T:   data.T,
		Agg: data.Agg,
	}

	if len(data.D) == 0 {
		return result, nil
	}

	var (
		bids []string
		uids []string
	)

	encFields := []string{"realname"}

	for _, v := range data.D {
		bids = append(bids, v.BID)
		uids = append(uids, v.UID)

		encFields = append(encFields, "bankcard"+v.BID)
	}

	bankcards, err := bankcardListDBByIDs(bids)
	if err != nil {
		return result, pushLog(err, helper.DBErr)
	}

	//fmt.Println("bids = ", bids)
	//fmt.Println("uids = ", uids)

	recs, err := grpc_t.DecryptAll(uids, true, encFields)
	if err != nil {
		fmt.Println("grpc_t.Decrypt err = ", err)
		return result, errors.New(helper.GetRPCErr)
	}

	/*
		d1, err := grpc_t.DecryptAll(uids, true, []string{"realname"})
		if err != nil {
			fmt.Println("grpc_t.Decrypt err = ", err)
			return result, errors.New(helper.GetRPCErr)
		}

		d2, err := grpc_t.DecryptAll(bids, true, []string{"bankcard"})
		if err != nil {
			fmt.Println("grpc_t.Decrypt err = ", err)
			return result, errors.New(helper.GetRPCErr)
		}
	*/
	// 处理返回前端的数据
	for _, v := range data.D {
		w := withdrawCols{
			Withdraw:           v,
			MemberBankNo:       recs[v.UID]["bankcard"+v.BID],
			MemberBankRealName: recs[v.UID]["realname"],
			MemberRealName:     recs[v.UID]["realname"],
		}

		card, ok := bankcards[v.BID]
		if ok {
			w.MemberBankID = card.BankID
			w.MemberBankAddress = card.BankAddress
		}

		result.D = append(result.D, w)
	}

	return result, nil
}

func bankcardListDBByIDs(ids []string) (map[string]BankCard_t, error) {

	data := make(map[string]BankCard_t)
	if len(ids) == 0 {
		return nil, errors.New(helper.UsernameErr)
	}

	ex := g.Ex{"id": ids}
	bankcards, _, err := BankcardsList(ex)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	for _, v := range bankcards {
		data[v.ID] = v
	}

	return data, nil
}

func BankcardsList(ex g.Ex) ([]BankCard_t, string, error) {

	var data []BankCard_t
	t := dialect.From("tbl_member_bankcard")
	query, _, _ := t.Select(colsBankcard...).Where(ex).Order(g.C("created_at").Desc()).ToSQL()
	err := meta.MerchantDB.Select(&data, query)
	if err != nil && err != sql.ErrNoRows {
		return data, "db", err
	}

	return data, "", nil
}

func RecordGroup(page, pageSize int, startTime, endTime string, ex g.Ex, parentName string) (AgencyTransferRecordData, error) {

	data := AgencyTransferRecordData{}
	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.TimeToLoc(endTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		ex["updated_at"] = g.Op{"between": g.Range(startAt, endAt)}
	}
	orEx := g.Or()
	if parentName != "" {

		orEx = g.Or(
			g.Ex{"after_name": parentName},
			g.Ex{"before_name": parentName},
		)
	}
	ex["prefix"] = meta.Prefix
	t := dialect.From("tbl_agency_transfer_record")
	if page == 1 {
		query, _, _ := t.Select(g.COUNT(1)).Where(g.And(ex, orEx)).ToSQL()
		err := meta.MerchantDB.Get(&data.T, query)
		if err != nil {
			return data, pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
		}

		if data.T == 0 {
			return data, nil
		}
	}

	offset := (page - 1) * pageSize
	query, _, _ := t.Select(colsAgencyTransferRecord...).Where(g.And(ex, orEx)).
		Order(g.C("updated_at").Desc()).Offset(uint(offset)).Limit(uint(pageSize)).ToSQL()
	err := meta.MerchantDB.Select(&data.D, query)
	if err != nil && err != sql.ErrNoRows {
		return data, pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	return data, nil
}

func RecordIssuse(id string) ([]string, error) {

	tableName := "tbl_vncp_plan_issues"
	var result []string

	ex := g.Ex{
		"plan_id": id,
	}
	build := dialect.From(tableName).Where(ex)

	build = build.Select(
		"id",
	).Order(g.C("created_at").Desc())
	query, _, _ := build.ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Select(&result, query)
	if err != nil {
		return result, err
	}
	return result, nil
}

func RecordOrder(page, pageSize int, ex g.Ex) (OrderData, error) {

	data := OrderData{}
	ex["prefix"] = meta.Prefix
	t := dialect.From("tbl_vncp_orders")
	if page == 1 {
		query, _, _ := t.Select(g.COUNT("id")).Where(ex).ToSQL()
		fmt.Println(query)
		err := meta.MerchantDB.Get(&data.T, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}

		if data.T == 0 {
			return data, nil
		}
	}

	offset := pageSize * (page - 1)
	query, _, _ := t.Select(g.C("username"), g.C("pay_amount"), g.C("bonus").As("bonus")).Where(ex).
		Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("created_at").Desc()).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Select(&data.D, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	for i := 0; i < len(data.D); i++ {
		pay, _ := decimal.NewFromString(data.D[i].PayAmount)
		bonus, _ := decimal.NewFromString(data.D[i].Bonus)
		data.D[i].NetAmount = bonus.Sub(pay).StringFixed(4)
	}
	return data, nil
}
