package model

import (
	"database/sql"
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/olivere/elastic/v7"
	"github.com/shopspring/decimal"
	"merchant2/contrib/helper"
	"strings"
	"time"
)

type Inspection struct {
	T int64            `json:"t"`
	D []InspectionData `json:"d"`
}

type PromoRecord struct {
	Id           string  `json:"id" db:"id"`
	Uid          string  `json:"uid" db:"uid"`
	Username     string  `json:"username" db:"username"`
	Level        int     `json:"level" db:"level"`
	TopUid       int     `json:"top_uid" db:"top_uid"`
	TopName      string  `json:"top_name" db:"top_name"`
	ParentUid    string  `json:"parent_uid" db:"parent_uid"`
	ParentName   string  `json:"parent_name" db:"parent_name"`
	Pid          string  `json:"pid" db:"pid"`
	Title        string  `json:"title" db:"title"`
	Amount       float64 `json:"amount" db:"amount"`
	BonusType    int     `json:"bonus_type" db:"bonus_type"`
	BonusRate    int     `json:"bonus_rate" db:"bonus_rate"`
	Bonus        float64 `json:"bonus" db:"bonus"`
	Flow         float64 `json:"flow" db:"flow"`
	Multiple     int     `json:"multiple" db:"multiple"`
	State        int     `json:"state" db:"state"`
	CreatedAt    int64   `json:"created_at" db:"created_at"`
	ReviewAt     int     `json:"review_at" db:"review_at"`
	ReviewUid    int     `json:"review_uid" db:"review_uid"`
	ReviewName   string  `json:"review_name" db:"review_name"`
	InspectAt    int     `json:"inspect_at" db:"inspect_at"`
	InspectUid   int     `json:"inspect_uid" db:"inspect_uid"`
	InspectName  string  `json:"inspect_name" db:"inspect_name"`
	InspectState int     `json:"inspect_state" db:"inspect_state"`
}

type PromoData struct {
	Id          string `json:"id" db:"id"`
	Prefix      string `json:"prefix" db:"prefix"`
	Title       string `json:"title" db:"title"`
	Period      int    `json:"period" db:"period"`
	Sort        int    `json:"sort" db:"sort"`
	Flag        string `json:"flag" db:"flag"`
	State       int    `json:"state" db:"state"`
	StartAt     int64  `json:"start_at" db:"start_at"`
	EndAt       int    `json:"end_at" db:"end_at"`
	ShowAt      int    `json:"show_at" db:"show_at"`
	CreatedAt   int64  `json:"created_at" db:"created_at"`
	CreatedUid  int64  `json:"created_uid" db:"created_uid"`
	CreatedName string `json:"created_name" db:"created_name"`
	UpdatedAt   int    `json:"updated_at" db:"updated_at"`
	UpdatedUid  int64  `json:"updated_uid" db:"updated_uid"`
	UpdatedName string `json:"updated_name" db:"updated_name"`
	ApplyTotal  int    `json:"apply_total" db:"apply_total"`
	ApplyDaily  int    `json:"apply_daily" db:"apply_daily"`
	Platforms   string `json:"platforms" db:"platforms"`
}

type InspectionData struct {
	No               string `json:"no"`
	Username         string `json:"username"`
	Level            string `json:"level"`
	TopName          string `json:"top_name"`
	Title            string `json:"title"`
	Amount           string `json:"amount"`
	RewardAmount     string `json:"reward_amount"`
	FlowMultiple     string `json:"flow_multiple"`
	FlowAmount       string `json:"flow_amount"`
	FinishedAmount   string `json:"finished_amount"`
	UnfinishedAmount string `json:"unfinished_amount"`
	CreatedAt        int64  `json:"created_at"`
	ReviewName       string `json:"review_name"`
	Ty               string `json:"ty"`
	Pid              string `json:"pid"`
	Platforms        string `json:"platforms"`
	RecordId         string `json:"recordId"`
}

type PromoInspection struct {
	Id               string `json:"id" db:"id"`
	Uid              string `json:"uid" db:"uid"`
	Username         string `json:"username" db:"username"`
	TopUid           string `json:"top_uid" db:"top_uid"`
	TopName          string `json:"top_name" db:"top_name"`
	ParentUid        string `json:"parent_uid" db:"parent_uid"`
	ParentName       string `json:"parent_name" db:"parent_name"`
	Level            int    `json:"level" db:"level"`
	Pid              string `json:"pid" db:"pid"`
	Title            string `json:"title" db:"title"`
	Platforms        string `json:"platforms" db:"platforms"`
	State            string `json:"state" db:"state"`
	CapitalAmount    string `json:"capital_amount" db:"capital_amount"`
	DividendAmount   string `json:"dividend_amount" db:"dividend_amount"`
	FlowMultiple     string `json:"flow_multiple" db:"flow_multiple"`
	FlowAmount       string `json:"flow_amount" db:"flow_amount"`
	FinishedAmount   string `json:"finished_amount" db:"finished_amount"`
	UnfinishedAmount string `json:"unfinished_amount" db:"unfinished_amount"`
	ReviewAt         int64  `json:"review_at" db:"review_at"`
	ReviewUid        string `json:"review_uid" db:"review_uid"`
	ReviewName       string `json:"review_name" db:"review_name"`
	InspectAt        int64  `json:"inspect_at" db:"inspect_at"`
	InspectUid       string `json:"inspect_uid" db:"inspect_uid"`
	InspectName      string `json:"inspect_name" db:"inspect_name"`
	Ty               string `json:"ty" db:"ty"`
	BillNo           string `json:"bill_no" db:"bill_no"`
	Remark           string `json:"remark" db:"remark"`
}

type PagePromoInspection struct {
	D []PromoInspection `json:"d"`
	T int64             `json:"t"`
}

func InspectionList(username string) (Inspection, Member, error) {

	var data Inspection
	i := 1
	now := time.Now().Unix()
	//查用户
	mb, err := MemberFindOne(username)
	if err != nil || mb.Username == "" {
		return data, mb, errors.New(helper.UsernameErr)
	}
	//上一次提现成功
	var cutTime int64
	lastWithdraw, err := getWithdrawLast(username)
	fmt.Println(lastWithdraw)
	if err != nil && err != sql.ErrNoRows {
		return data, mb, errors.New(helper.DBErr)
	}
	if err != sql.ErrNoRows {
		cutTime = lastWithdraw.CreatedAt
	}

	lastInspection, err := getInspectionLast(username)
	if cutTime < lastInspection.InspectAt {
		cutTime = lastInspection.InspectAt
	}

	//查活动记录
	recordList, err := promoRecrodList(username)
	promoMap := map[string]PromoData{}
	if err != nil {
		return data, mb, errors.New(helper.DBErr)
	}
	//查活动记录对应的活动 理论上会有多个活
	var pids []string
	for _, v := range recordList {
		pids = append(pids, v.Pid)
	}
	//参加的活动
	promolist, err := promoDataList(pids)
	if err != nil {
		return data, mb, errors.New(helper.DBErr)
	}

	for _, v := range promolist {
		promoMap[v.Id] = v
	}
	//上次提现至今的流水
	totalVaild, err := EsPlatValidBet(username, "", cutTime, now)
	if err != nil {
		return data, mb, errors.New(helper.DBErr)
	}
	//查升级红利
	dividendAmount, err := EsDividend(username, cutTime, now, []int{DividendUpgrade, DividendBirthday, DividendMonthly, DividendRedPacket})
	if err != nil {
		return data, mb, errors.New(helper.DBErr)
	}
	//组装红利的流水稽查
	data.D = append(data.D, InspectionData{
		No:               fmt.Sprintf(`%d`, i),
		Username:         username,
		Level:            fmt.Sprintf(`%d`, mb.Level),
		TopName:          mb.TopName,
		Title:            "红利/礼金",
		Amount:           "0",
		RewardAmount:     dividendAmount.StringFixed(4),
		ReviewName:       "系统自动发送",
		FlowMultiple:     "1",
		FlowAmount:       dividendAmount.StringFixed(4),
		FinishedAmount:   totalVaild.StringFixed(4),
		UnfinishedAmount: dividendAmount.Sub(totalVaild).StringFixed(4),
		CreatedAt:        0,
		Ty:               "2",
		Pid:              "0",
	})
	i++

	//查调整
	adjustAmount, err := EsAdjust(username, cutTime, now)
	if err != nil {
		return data, mb, errors.New(helper.DBErr)
	}
	//组装vip礼金的流水稽查
	data.D = append(data.D, InspectionData{
		No:               fmt.Sprintf(`%d`, i),
		Username:         username,
		Level:            fmt.Sprintf(`%d`, mb.Level),
		TopName:          mb.TopName,
		Title:            "调整（分数调整和输赢调整）",
		Amount:           "0",
		RewardAmount:     adjustAmount.StringFixed(4),
		ReviewName:       "",
		FlowMultiple:     "1",
		FlowAmount:       adjustAmount.StringFixed(4),
		FinishedAmount:   totalVaild.StringFixed(4),
		UnfinishedAmount: adjustAmount.Sub(totalVaild).StringFixed(4),
		CreatedAt:        0,
		Ty:               "4",
		Pid:              "0",
	})
	i++

	//查存款
	depostAmount, err := EsDepost(username, cutTime, now)
	if err != nil {
		return data, mb, errors.New(helper.DBErr)
	}
	//组装存款的流水稽查
	data.D = append(data.D, InspectionData{
		No:               fmt.Sprintf(`%d`, i),
		Username:         username,
		Level:            fmt.Sprintf(`%d`, mb.Level),
		TopName:          mb.TopName,
		Title:            "存款",
		Amount:           depostAmount.StringFixed(4),
		RewardAmount:     "0",
		ReviewName:       "无",
		FlowMultiple:     "1",
		FlowAmount:       depostAmount.StringFixed(4),
		FinishedAmount:   totalVaild.StringFixed(4),
		UnfinishedAmount: depostAmount.Sub(totalVaild).StringFixed(4),
		CreatedAt:        0,
		Ty:               "1",
		Pid:              "0",
	})
	i++

	//查活动对应场馆的流水总和
	for _, v := range recordList {
		validBetAmount, err := EsPlatValidBet(username, promoMap[v.Pid].Platforms, promoMap[v.Pid].StartAt, now)
		if err != nil {
			return data, mb, errors.New(helper.ESErr)
		}
		//组装活动的流水稽查
		data.D = append(data.D, InspectionData{
			No:               fmt.Sprintf(`%d`, i),
			Username:         username,
			Level:            fmt.Sprintf(`%d`, mb.Level),
			TopName:          mb.TopName,
			Title:            v.Title,
			Amount:           fmt.Sprintf(`%f`, v.Amount),
			RewardAmount:     fmt.Sprintf(`%f`, v.Bonus),
			ReviewName:       promoMap[v.Pid].UpdatedName,
			FlowMultiple:     fmt.Sprintf(`%d`, v.Multiple),
			FlowAmount:       fmt.Sprintf(`%f`, v.Flow),
			FinishedAmount:   validBetAmount.StringFixed(4),
			UnfinishedAmount: decimal.NewFromFloat(v.Flow).Sub(validBetAmount).StringFixed(4),
			CreatedAt:        v.CreatedAt,
			Ty:               "3",
			Pid:              v.Id,
			Platforms:        promoMap[v.Pid].Platforms,
			RecordId:         v.Id,
		})
		i++
	}
	data.T = int64(i)
	return data, mb, nil
}

func InspectionReview(username, inspectState, billNo, remark string, admin map[string]string) (bool, error) {

	inspection, mb, err := InspectionList(username)
	if err != nil {
		return false, err
	}

	//有提交订单号的去校验订单号是否这个用户的提款订单
	if len(billNo) > 0 {
		ex := g.Ex{
			"username": username,
			"id":       billNo,
		}
		w := WithdrawRecord{}

		query, _, _ := dialect.From("tbl_withdraw").Select(colWithdrawRecord...).Where(ex).Order(g.C("created_at").Desc()).Limit(1).ToSQL()
		fmt.Println(query)
		err := meta.MerchantDB.Get(&w, query)
		if err != nil {
			return false, errors.New(helper.OrderNotExist)
		}

	}

	//开启事务
	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return false, pushLog(err, helper.DBErr)
	}
	for _, v := range inspection.D {
		data := &PromoInspection{
			Id:               helper.GenId(),
			Uid:              mb.UID,
			Username:         mb.Username,
			TopUid:           mb.TopUid,
			TopName:          mb.TopName,
			ParentUid:        mb.ParentUid,
			ParentName:       mb.ParentName,
			Level:            mb.Level,
			Pid:              v.Pid,
			Title:            v.Title,
			Platforms:        v.Platforms,
			State:            inspectState,
			CapitalAmount:    v.Amount,
			DividendAmount:   v.RewardAmount,
			FlowMultiple:     v.FlowMultiple,
			FlowAmount:       v.FlowAmount,
			FinishedAmount:   v.FinishedAmount,
			UnfinishedAmount: v.UnfinishedAmount,
			ReviewAt:         time.Now().Unix(),
			ReviewUid:        admin["id"],
			ReviewName:       admin["name"],
			InspectAt:        time.Now().Unix(),
			InspectUid:       admin["id"],
			InspectName:      admin["name"],
			Ty:               v.Ty,
			BillNo:           billNo,
			Remark:           remark,
		}

		// 插入稽查历史
		queryInsert, _, _ := dialect.Insert("tbl_promo_inspection").Rows(data).ToSQL()
		_, err = tx.Exec(queryInsert)
		if err != nil {
			_ = tx.Rollback()
			return false, pushLog(fmt.Errorf("%s,[%s]", err.Error(), queryInsert), helper.DBErr)
		}
		//是活动的要更新活动记录稽查状态
		if v.Ty == "3" {
			ex := g.Ex{
				"id": v.RecordId,
			}
			record := g.Record{
				"state": inspectState,
			}
			query, _, _ := dialect.Update("tbl_promo_record").Set(record).Where(ex).ToSQL()
			_, err = tx.Exec(query)
			if err != nil {
				_ = tx.Rollback()
				return false, pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
			}
		}
	}

	tx.Commit()
	return true, nil
}

func InspectionHistory(ex g.Ex, page, pageSize int) (PagePromoInspection, error) {

	var data PagePromoInspection
	t := dialect.From("tbl_promo_inspection")
	if page == 1 {
		query, _, _ := t.Select(g.COUNT("id")).Where(ex).ToSQL()
		fmt.Println("总代佣金:sql:", query)
		err := meta.MerchantDB.Get(&data.T, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}

		if data.T == 0 {
			return data, nil
		}
	}

	offset := pageSize * (page - 1)
	query, _, _ := t.Select(colsPromoInspection...).Where(ex).
		Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("review_at").Desc()).ToSQL()
	fmt.Println("稽查历史:sql:", query)
	err := meta.MerchantDB.Select(&data.D, query)
	if err != nil && err != sql.ErrNoRows {
		return data, pushLog(err, helper.DBErr)
	}
	return data, nil
}

func promoRecrodList(username string) ([]PromoRecord, error) {

	ex := g.Ex{
		"username":      username,
		"state":         2,
		"inspect_state": 1,
	}
	var data []PromoRecord
	t := dialect.From("tbl_promo_record")

	query, _, _ := t.Select(colsPromoRecord...).Where(ex).Order(g.C("created_at").Desc()).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Select(&data, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil

}

func promoDataList(pids []string) ([]PromoData, error) {

	var data []PromoData
	if len(pids) == 0 {
		return data, nil

	}
	ex := g.Ex{
		"id":    pids,
		"state": "2",
	}
	t := dialect.From("tbl_promo")

	query, _, _ := t.Select(colsPromoData...).Where(ex).Order(g.C("created_at").Desc()).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Select(&data, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}
	return data, nil
}

func getWithdrawLast(username string) (Withdraw, error) {

	data := Withdraw{}

	query := elastic.NewBoolQuery()
	query.Filter(elastic.NewTermQuery("state", WithdrawSuccess))
	query.Filter(elastic.NewTermQuery("username", username))
	query.Filter(elastic.NewTermQuery("prefix", meta.Prefix))
	_, esResult, _, err := EsQuerySearch(
		esPrefixIndex("tbl_withdraw"), "created_at", 1, 1, withdrawFields, query, nil)
	if err != nil {
		return data, pushLog(err, helper.ESErr)
	}
	for _, v := range esResult {

		record := Withdraw{}
		_ = helper.JsonUnmarshal(v.Source, &record)
		data = record
		return data, nil
	}
	return data, sql.ErrNoRows
}

func getInspectionLast(username string) (PromoInspection, error) {

	ex := g.Ex{
		"username": username,
		"state":    []int{2, 3},
	}
	w := PromoInspection{}

	query, _, _ := dialect.From("tbl_promo_inspection").Select(colsPromoInspection...).Where(ex).Order(g.C("inspect_at").Desc()).Limit(1).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Get(&w, query)
	if err != nil && err != sql.ErrNoRows {
		return w, pushLog(err, helper.DBErr)
	}
	if err == sql.ErrNoRows {
		return w, nil
	}
	return w, nil
}

// EsPlatValidBet 获取指定会员指定场馆的有效投注
func EsPlatValidBet(username string, pid string, startAt, endAt int64) (decimal.Decimal, error) {

	waterFlow := decimal.NewFromFloat(0.0000)
	if startAt == 0 && endAt == 0 {
		return waterFlow, errors.New(helper.QueryTimeRangeErr)
	}

	boolQuery := elastic.NewBoolQuery()

	filters := make([]elastic.Query, 0)
	rg := elastic.NewRangeQuery("settle_time").Gte(startAt * 1000)
	if startAt == 0 {
		rg.IncludeLower(false)
	}
	if endAt == 0 {
		rg.IncludeUpper(false)
	}

	if endAt > 0 {
		rg.Lt(endAt)
	}

	filters = append(filters, rg)
	boolQuery.Filter(filters...)

	terms := make([]elastic.Query, 0)
	terms = append(terms, elastic.NewTermQuery("name", username))
	shouldQuery := elastic.NewBoolQuery()
	if len(pid) > 0 {
		pids := strings.Split(pid, ",")
		for _, v := range pids {
			if len(v) <= 20 {
				//查询域名,采用模糊匹配
				shouldQuery.Should(elastic.NewTermQuery("api_type", v))
			}
		}

		boolQuery.Must(shouldQuery)
	}
	terms = append(terms, elastic.NewTermQuery("flag", 1))

	boolQuery.Must(terms...)

	fsc := elastic.NewFetchSourceContext(true)
	//打印es查询json
	esService := meta.ES.Search().FetchSourceContext(fsc).Query(boolQuery).Size(0)
	resOrder, err := esService.Index(pullPrefixIndex("tbl_game_record")).
		Aggregation("valid_bet_amount_agg", elastic.NewSumAggregation().Field("valid_bet_amount")).Do(ctx)
	if err != nil {
		fmt.Println(err)
		return waterFlow, err
	}

	validBet, ok := resOrder.Aggregations.Sum("valid_bet_amount_agg")
	if validBet == nil || !ok {
		return waterFlow, errors.New("agg error")
	}

	return decimal.NewFromFloat(*validBet.Value), nil
}

func EsDepost(username string, startAt, endAt int64) (decimal.Decimal, error) {

	waterFlow := decimal.NewFromFloat(0.0000)
	if startAt == 0 && endAt == 0 {
		return waterFlow, errors.New(helper.QueryTimeRangeErr)
	}

	boolQuery := elastic.NewBoolQuery()

	filters := make([]elastic.Query, 0)
	rg := elastic.NewRangeQuery("created_at").Gte(startAt)
	if startAt == 0 {
		rg.IncludeLower(false)
	}
	if endAt == 0 {
		rg.IncludeUpper(false)
	}

	if endAt > 0 {
		rg.Lt(endAt)
	}

	filters = append(filters, rg)
	boolQuery.Filter(filters...)

	terms := make([]elastic.Query, 0)
	terms = append(terms, elastic.NewTermQuery("name", username))
	terms = append(terms, elastic.NewTermQuery("state", DepositSuccess))

	boolQuery.Must(terms...)

	fsc := elastic.NewFetchSourceContext(true)
	//打印es查询json
	esService := meta.ES.Search().FetchSourceContext(fsc).Query(boolQuery).Size(0)
	resOrder, err := esService.Index(esPrefixIndex("tbl_deposit")).
		Aggregation("amount_agg", elastic.NewSumAggregation().Field("amount")).Do(ctx)
	if err != nil {
		fmt.Println(err)
		return waterFlow, err
	}

	depositAmount, ok := resOrder.Aggregations.Sum("amount_agg")
	if depositAmount == nil || !ok {
		return waterFlow, errors.New("agg error")
	}

	return decimal.NewFromFloat(*depositAmount.Value), nil
}

func EsDividend(username string, startAt, endAt int64, ty []int) (decimal.Decimal, error) {

	waterFlow := decimal.NewFromFloat(0.0000)
	if startAt == 0 && endAt == 0 {
		return waterFlow, errors.New(helper.QueryTimeRangeErr)
	}

	boolQuery := elastic.NewBoolQuery()

	filters := make([]elastic.Query, 0)
	rg := elastic.NewRangeQuery("review_at").Gte(startAt)
	if startAt == 0 {
		rg.IncludeLower(false)
	}
	if endAt == 0 {
		rg.IncludeUpper(false)
	}

	if endAt > 0 {
		rg.Lt(endAt)
	}

	filters = append(filters, rg)
	boolQuery.Filter(filters...)

	terms := make([]elastic.Query, 0)
	terms = append(terms, elastic.NewTermQuery("username", username))
	terms = append(terms, elastic.NewRangeQuery("ty").Gte(DividendUpgrade).Lte(DividendRedPacket))
	terms = append(terms, elastic.NewTermQuery("state", DividendReviewPass))

	boolQuery.Must(terms...)

	fsc := elastic.NewFetchSourceContext(true)
	//打印es查询json
	esService := meta.ES.Search().FetchSourceContext(fsc).Query(boolQuery).Size(0)
	resOrder, err := esService.Index(esPrefixIndex("tbl_member_dividend")).
		Aggregation("amount_agg", elastic.NewSumAggregation().Field("amount")).Do(ctx)
	if err != nil {
		fmt.Println(err)
		return waterFlow, err
	}

	handOutAmount, ok := resOrder.Aggregations.Sum("amount_agg")
	if handOutAmount == nil || !ok {
		return waterFlow, errors.New("agg error")
	}

	return decimal.NewFromFloat(*handOutAmount.Value), nil
}

func EsAdjust(username string, startAt, endAt int64) (decimal.Decimal, error) {

	waterFlow := decimal.NewFromFloat(0.0000)
	if startAt == 0 && endAt == 0 {
		return waterFlow, errors.New(helper.QueryTimeRangeErr)
	}

	boolQuery := elastic.NewBoolQuery()

	filters := make([]elastic.Query, 0)
	rg := elastic.NewRangeQuery("review_at").Gte(startAt)
	if startAt == 0 {
		rg.IncludeLower(false)
	}
	if endAt == 0 {
		rg.IncludeUpper(false)
	}

	if endAt > 0 {
		rg.Lt(endAt)
	}

	filters = append(filters, rg)
	boolQuery.Filter(filters...)

	terms := make([]elastic.Query, 0)
	terms = append(terms, elastic.NewTermQuery("username", username))
	terms = append(terms, elastic.NewTermQuery("is_turnover", "1"))
	terms = append(terms, elastic.NewTermQuery("hand_out_state", AdjustSuccess))

	boolQuery.Must(terms...)

	fsc := elastic.NewFetchSourceContext(true)
	//打印es查询json
	esService := meta.ES.Search().FetchSourceContext(fsc).Query(boolQuery).Size(0)
	resOrder, err := esService.Index(esPrefixIndex("tbl_member_adjust")).
		Aggregation("amount_agg", elastic.NewSumAggregation().Field("amount")).Do(ctx)
	if err != nil {
		fmt.Println(err)
		return waterFlow, err
	}

	handOutAmount, ok := resOrder.Aggregations.Sum("amount_agg")
	if handOutAmount == nil || !ok {
		return waterFlow, errors.New("agg error")
	}

	return decimal.NewFromFloat(*handOutAmount.Value), nil
}
