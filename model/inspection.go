package model

import (
	"database/sql"
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/olivere/elastic/v7"
	"github.com/shopspring/decimal"
	"merchant/contrib/helper"
	"strings"
	"time"
)

type Inspection struct {
	T   int64            `json:"t"`
	D   []InspectionData `json:"d"`
	Agg AggInspection    `json:"agg"`
}

type AggInspection struct {
	FlowAmount       string `json:"flow_amount"`
	VaildAmount      string `json:"vaild_amount"`
	UnFishFlowAmount string `json:"un_fish_flow_amount"`
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

	cutTime = int64(mb.LastWithdrawAt)

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
	dividendData, err := EsDividend(username, cutTime, now)
	if err != nil {
		return data, mb, errors.New(helper.DBErr)
	}
	var needFlowAmount decimal.Decimal

	if dividendData.T > 0 {
		for _, v := range dividendData.D {
			dividendAmount := decimal.NewFromFloat(v.Amount)
			flow := decimal.NewFromFloat(v.WaterFlow)
			//组装红利的流水稽查
			uf := flow.Sub(totalVaild)
			if uf.Cmp(decimal.Zero) < 0 {
				uf = decimal.Zero
			}
			data.D = append(data.D, InspectionData{
				No:               fmt.Sprintf(`%d`, i),
				Username:         username,
				Level:            fmt.Sprintf(`%d`, mb.Level),
				TopName:          mb.TopName,
				Title:            "红利/礼金",
				Amount:           "0.0000",
				RewardAmount:     dividendAmount.StringFixed(4),
				ReviewName:       v.ReviewName,
				FlowMultiple:     fmt.Sprintf(`%d`, v.WaterMultiple),
				FlowAmount:       flow.StringFixed(4),
				FinishedAmount:   totalVaild.StringFixed(4),
				UnfinishedAmount: uf.StringFixed(4),
				CreatedAt:        int64(v.ReviewAt),
				Ty:               "2",
				Pid:              "0",
				RecordId:         v.ID,
			})
			needFlowAmount = needFlowAmount.Add(flow)
			i++
		}
	}

	//查调整
	adjustData, err := EsAdjust(username, cutTime, now)
	if err != nil {
		return data, mb, errors.New(helper.DBErr)
	}
	if adjustData.T > 0 {
		for _, v := range adjustData.D {
			adjustAmount := decimal.NewFromFloat(v.Amount)
			multi := decimal.NewFromInt(int64(v.TurnoverMulti))
			//组装vip礼金的流水稽查
			uf := adjustAmount.Mul(multi).Sub(totalVaild)
			if uf.Cmp(decimal.Zero) < 0 {
				uf = decimal.Zero
			}
			data.D = append(data.D, InspectionData{
				No:               fmt.Sprintf(`%d`, i),
				Username:         username,
				Level:            fmt.Sprintf(`%d`, mb.Level),
				TopName:          mb.TopName,
				Title:            "调整（分数调整和输赢调整）",
				Amount:           "0.0000",
				RewardAmount:     adjustAmount.StringFixed(4),
				ReviewName:       v.ReviewName,
				FlowMultiple:     fmt.Sprintf(`%d`, v.TurnoverMulti),
				FlowAmount:       adjustAmount.Mul(multi).StringFixed(4),
				FinishedAmount:   totalVaild.StringFixed(4),
				UnfinishedAmount: uf.StringFixed(4),
				CreatedAt:        v.ReviewAt,
				Ty:               "4",
				Pid:              "0",
				RecordId:         v.ID,
			})
			needFlowAmount = needFlowAmount.Add(adjustAmount.Mul(multi))

			i++
		}
	}

	//查存款
	depostList, err := EsDepost(username, cutTime, now)
	if err != nil {
		return data, mb, errors.New(helper.DBErr)
	}

	if depostList.T > 0 {
		//组装存款的流水稽查
		for _, v := range depostList.D {
			depostAmount := decimal.NewFromFloat(v.Amount)
			uf := depostAmount.Sub(totalVaild)
			if uf.Cmp(decimal.Zero) < 0 {
				uf = decimal.Zero
			}
			data.D = append(data.D, InspectionData{
				No:               fmt.Sprintf(`%d`, i),
				Username:         username,
				Level:            fmt.Sprintf(`%d`, mb.Level),
				TopName:          mb.TopName,
				Title:            "存款",
				Amount:           depostAmount.StringFixed(4),
				RewardAmount:     "0.0000",
				ReviewName:       "无",
				FlowMultiple:     "1",
				FlowAmount:       depostAmount.StringFixed(4),
				FinishedAmount:   totalVaild.StringFixed(4),
				UnfinishedAmount: uf.StringFixed(4),
				CreatedAt:        0,
				Ty:               "1",
				Pid:              "0",
				RecordId:         v.ID,
			})
			needFlowAmount = needFlowAmount.Add(depostAmount)

			i++
		}
	}

	//查活动对应场馆的流水总和
	for _, v := range recordList {
		apitype := ""
		if promoMap[v.Pid].Flag != "rescue" {
			apitype = promoMap[v.Pid].Platforms
		}
		validBetAmount, err := EsPlatValidBet(username, apitype, promoMap[v.Pid].StartAt, now)
		if err != nil {
			return data, mb, errors.New(helper.ESErr)
		}
		uf := decimal.NewFromFloat(v.Flow).Sub(validBetAmount)
		if uf.Cmp(decimal.Zero) == -1 {
			uf = decimal.Zero
		}
		rvName := v.ReviewName
		if len(rvName) == 0 {
			rvName = "系统"
		}
		//组装活动的流水稽查
		data.D = append(data.D, InspectionData{
			No:               fmt.Sprintf(`%d`, i),
			Username:         username,
			Level:            fmt.Sprintf(`%d`, mb.Level),
			TopName:          mb.TopName,
			Title:            v.Title,
			Amount:           fmt.Sprintf(`%.4f`, v.Amount),
			RewardAmount:     fmt.Sprintf(`%.4f`, v.Bonus),
			ReviewName:       rvName,
			FlowMultiple:     fmt.Sprintf(`%d`, v.Multiple),
			FlowAmount:       fmt.Sprintf(`%.4f`, v.Flow),
			FinishedAmount:   validBetAmount.StringFixed(4),
			UnfinishedAmount: uf.StringFixed(4),
			CreatedAt:        v.CreatedAt,
			Ty:               "3",
			Pid:              v.Pid,
			Platforms:        promoMap[v.Pid].Platforms,
			RecordId:         v.Id,
		})

		needFlowAmount = needFlowAmount.Add(decimal.NewFromFloat(v.Flow))
		i++
	}
	data.T = int64(i) - 1
	uf := needFlowAmount.Sub(totalVaild)
	if uf.Cmp(decimal.Zero) < 0 {
		uf = decimal.Zero
	}
	agg := AggInspection{
		FlowAmount:       needFlowAmount.StringFixed(4),
		VaildAmount:      totalVaild.StringFixed(4),
		UnFishFlowAmount: uf.StringFixed(4),
	}
	data.Agg = agg
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
				"inspect_at":    time.Now().Unix(),
				"inspect_uid":   admin["id"],
				"inspect_name":  admin["name"],
				"inspect_state": inspectState,
			}
			query, _, _ := dialect.Update("tbl_promo_record").Set(record).Where(ex).ToSQL()
			fmt.Println(query)
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

	query, _, _ := t.Select(g.COUNT("*")).Where(ex).ToSQL()
	fmt.Println("稽查历史:sql:", query)
	err := meta.MerchantDB.Get(&data.T, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	if data.T == 0 {
		return data, nil
	}

	offset := pageSize * (page - 1)
	query, _, _ = t.Select(colsPromoInspection...).Where(ex).
		Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("review_at").Desc()).ToSQL()
	fmt.Println("稽查历史:sql:", query)
	err = meta.MerchantDB.Select(&data.D, query)
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
		"id": pids,
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
		rg.Lt(endAt * 1000)
	}

	filters = append(filters, rg)
	boolQuery.Filter(filters...)

	terms := make([]elastic.Query, 0)
	terms = append(terms, elastic.NewTermQuery("name", username))
	fmt.Println("pid:", pid)
	if strings.Contains(pid, ",") {
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
	} else if len(pid) > 0 {
		terms = append(terms, elastic.NewTermQuery("api_type", pid))
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

func EsDepost(username string, startAt, endAt int64) (FDepositData, error) {

	data := FDepositData{}

	query := elastic.NewBoolQuery()

	query.Filter(elastic.NewTermsQuery("state", DepositSuccess))

	query.Filter(elastic.NewTermQuery("username", username))

	query.Filter(elastic.NewRangeQuery("created_at").Gte(startAt).Lte(endAt))

	query.Filter(elastic.NewTermQuery("prefix", meta.Prefix))
	t, esResult, _, err := EsQuerySearch(
		esPrefixIndex("tbl_deposit"), "created_at", 1, 100, depositFields, query, nil)
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

func EsDividend(username string, startAt, endAt int64) (DividendEsData, error) {

	data := DividendEsData{}
	query := elastic.NewBoolQuery()
	query.Filter(elastic.NewTermQuery("username", username))
	query.MustNot(elastic.NewTermsQuery("ty", DividendPromo))
	query.Filter(elastic.NewTermQuery("state", DividendReviewPass))
	query.Filter(elastic.NewTermQuery("water_limit", 2))

	if startAt != 0 && endAt != 0 {

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		query.Filter(elastic.NewRangeQuery("review_at").Gte(startAt).Lte(endAt))
	}

	query.Filter(elastic.NewTermQuery("prefix", meta.Prefix))
	fmt.Println("query:", query)
	t, esResult, _, err := EsQuerySearch(
		esPrefixIndex("tbl_member_dividend"), "review_at", 1, 100, dividendFields, query, nil)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	var names []string
	data.T = t
	for _, v := range esResult {

		record := Dividend{}
		fmt.Println(string(v.Source))
		_ = helper.JsonUnmarshal(v.Source, &record)
		record.ID = v.Id
		fmt.Println(record)
		data.D = append(data.D, record)
		names = append(names, record.ParentName)
	}

	return data, nil
}

func EsAdjust(username string, startTime, endTime int64) (AdjustData, error) {

	data := AdjustData{}
	query := elastic.NewBoolQuery()
	if startTime != 0 && endTime != 0 {

		query.Filter(elastic.NewRangeQuery("review_at").Gte(startTime).Lte(endTime))
	}
	query.Filter(elastic.NewTermQuery("is_turnover", "1"))
	query.Filter(elastic.NewTermQuery("state", AdjustReviewPass))
	query.Filter(elastic.NewTermQuery("username", username))

	query.Filter(elastic.NewTermQuery("prefix", meta.Prefix))
	t, esResult, _, err := EsQuerySearch(
		esPrefixIndex("tbl_member_adjust"), "apply_at", 1, 100, adjustFields, query, nil)
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
