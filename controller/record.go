package controller

import (
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/olivere/elastic/v7"
	"github.com/shopspring/decimal"
	"github.com/valyala/fasthttp"
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"merchant2/model"
	"strconv"
	"strings"
)

type gameAdminParam struct {
	Username   string `rule:"none" msg:"username error" name:"username"`                                    // 下级账号
	ParentName string `rule:"none" name:"parent_name"`                                                      //
	Pid        string `rule:"digit" default:"0" msg:"pid error" name:"pid"`                                 // 场馆
	Flag       string `rule:"digit" default:"1" min:"1" max:"3" name:"flag"`                                //
	StartTime  string `rule:"none" name:"start_time"`                                                       // 查询开始时间
	EndTime    string `rule:"none" name:"end_time"`                                                         // 查询结束时间
	Page       int    `rule:"digit" default:"1" min:"1" msg:"page error" name:"page"`                       // 页码
	PageSize   int    `rule:"digit" default:"10" min:"10" max:"200" msg:"page_size error" name:"page_size"` // 页大小
}

type loginLogParam struct {
	Username  string `rule:"none" name:"username"`                                                         // 下级账号
	Ip        string `rule:"none" name:"ip"`                                                               //
	StartTime string `rule:"none" name:"start_time"`                                                       // 查询开始时间
	EndTime   string `rule:"none" name:"end_time"`                                                         // 查询结束时间
	Page      int    `rule:"digit" default:"1" min:"1" msg:"page error" name:"page"`                       // 页码
	PageSize  int    `rule:"digit" default:"10" min:"10" max:"200" msg:"page_size error" name:"page_size"` // 页大小
}

type depositParam struct {
	Username   string `rule:"none" name:"username"`                                                         // 下级账号
	ParentName string `rule:"none" name:"parent_name"`                                                      //
	State      uint16 `rule:"digit" default:"0" min:"0" max:"363" name:"state"`                             // 361:待确认 362:存款成功 363:已取消
	ChannelId  uint64 `rule:"digit" default:"0" min:"0" name:"channel_id"`                                  // 通道ID
	StartTime  string `rule:"none" name:"start_time"`                                                       // 查询开始时间
	EndTime    string `rule:"none" name:"end_time"`                                                         // 查询结束时间
	Page       int    `rule:"digit" default:"1" min:"1" msg:"page error" name:"page"`                       // 页码
	PageSize   int    `rule:"digit" default:"10" min:"10" max:"200" msg:"page_size error" name:"page_size"` // 页大小
}

type dividendParam struct {
	Username   string `rule:"none" name:"username"`                                                         // 下级账号
	ParentName string `rule:"none" name:"parent_name"`                                                      //
	Ty         uint16 `rule:"digit" default:"0" min:"0" max:"222" name:"ty"`                                //211 平台红利、212 升级红利、213 生日红利、214 每月红利、215 红包红利、216 维护补偿、217 存款优惠、218 活动红利、219 推荐红利、220 红利调整、221 负数清零
	StartTime  string `rule:"none" name:"start_time"`                                                       // 查询开始时间
	EndTime    string `rule:"none" name:"end_time"`                                                         // 查询结束时间
	Page       int    `rule:"digit" default:"1" min:"1" msg:"page error" name:"page"`                       // 页码
	PageSize   int    `rule:"digit" default:"10" min:"10" max:"200" msg:"page_size error" name:"page_size"` // 页大小
}

type rebateParam struct {
	Username   string `rule:"none" name:"username"`                                                         // 下级账号
	ParentName string `rule:"none" name:"parent_name"`                                                      //
	StartTime  string `rule:"none" name:"start_time"`                                                       // 查询开始时间
	EndTime    string `rule:"none" name:"end_time"`                                                         // 查询结束时间
	Page       int    `rule:"digit" default:"1" min:"1" msg:"page error" name:"page"`                       // 页码
	PageSize   int    `rule:"digit" default:"10" min:"10" max:"200" msg:"page_size error" name:"page_size"` // 页大小
}

type adjustRecordParam struct {
	Username   string `rule:"none" name:"username"`                                                         // 下级账号
	ParentName string `rule:"none" name:"parent_name"`                                                      //
	AdjustType string `rule:"digit" default:"0" min:"0" max:"3" name:"adjust_type" msg:"adjust type error"` //
	State      int    `rule:"digit" default:"0" min:"0" max:"263" name:"state" msg:"state error"`           //
	StartTime  string `rule:"none" name:"start_time"`                                                       // 查询开始时间
	EndTime    string `rule:"none" name:"end_time"`                                                         // 查询结束时间
	Page       int    `rule:"digit" default:"1" min:"1" msg:"page error" name:"page"`                       // 页码
	PageSize   int    `rule:"digit" default:"10" min:"10" max:"200" msg:"page_size error" name:"page_size"` // 页大小
}

type withdrawParam struct {
	Username       string `rule:"none" name:"username"`                                                         // 下级账号
	ParentName     string `rule:"none" name:"parent_name"`                                                      //
	State          uint16 `rule:"digit" default:"0" min:"0" max:"379" name:"state"`                             //371:审核中 372:审核拒绝 373:出款中 374:提款成功 375:出款失败 376:异常订单 377:代付失败
	MinAmount      string `rule:"none" default:"0" min:"0" name:"min_amount"`                                   //
	MaxAmount      string `rule:"none" default:"0" min:"0" name:"max_amount"`                                   //
	StartTime      string `rule:"none"  name:"start_time"`                                                      // 查询开始时间
	EndTime        string `rule:"none"  name:"end_time"`                                                        // 查询结束时间
	ApplyStartTime string `rule:"none"  name:"apply_start_time"`                                                // 查询开始时间
	ApplyEndTime   string `rule:"none"  name:"apply_end_time"`                                                  // 查询结束时间
	Page           int    `rule:"digit" default:"1" min:"1" msg:"page error" name:"page"`                       // 页码
	PageSize       int    `rule:"digit" default:"10" min:"10" max:"200" msg:"page_size error" name:"page_size"` // 页大小
}

type groupParam struct {
	Username   string `rule:"none" name:"username"`                                                         // 下级账号
	Uid        string `rule:"none" name:"uid"`                                                              //
	ParentName string `rule:"none" name:"parent_name"`                                                      //
	StartTime  string `rule:"none" name:"start_time"`                                                       // 查询开始时间
	EndTime    string `rule:"none" name:"end_time"`                                                         // 查询结束时间
	Page       int    `rule:"digit" default:"1" min:"1" msg:"page error" name:"page"`                       // 页码
	PageSize   int    `rule:"digit" default:"10" min:"10" max:"200" msg:"page_size error" name:"page_size"` // 页大小
}

type RecordController struct{}

// Transaction 账变记录列表
func (that *RecordController) Transaction(ctx *fasthttp.RequestCtx) {

	ty := ctx.QueryArgs().GetUintOrZero("ty")               // 1中心钱包 2佣金钱包
	uid := string(ctx.QueryArgs().Peek("uid"))              //
	types := string(ctx.QueryArgs().Peek("types"))          // 账变类型
	startTime := string(ctx.QueryArgs().Peek("start_time")) // 查询开始时间
	endTime := string(ctx.QueryArgs().Peek("end_time"))     // 查询结束时间
	page := ctx.QueryArgs().GetUintOrZero("page")           // 页码
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")  // 页大小

	if !validator.CheckStringDigit(uid) {
		helper.Print(ctx, false, helper.UIDErr)
		return
	}

	if ty < 1 || ty > 2 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if page == 0 {
		page = 1
	}

	if pageSize < 10 || pageSize > 200 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	ex := g.Ex{"uid": uid}
	tableName := "tbl_balance_transaction"

	// 中心钱包余额账变
	if ty == 1 {
		// 账变类型筛选
		if types != "" {
			cashTypes := strings.Split(types, ",")
			for _, v := range cashTypes {
				ct, err := strconv.Atoi(v)
				if err != nil || !(ct >= model.TransactionIn && ct <= model.TransactionPromoPayout) &&
					!(ct >= model.TransactionEBetTCPrize && ct <= model.TransactionOfflineDeposit) {
					helper.Print(ctx, false, helper.CashTypeErr)
					return
				}
			}

			if len(cashTypes) > 0 {
				ex["cash_type"] = types
			}
		}
	} else {

		tableName = "tbl_commission_transaction"

		// 佣金钱包账变
		if types != "" {
			cashTypes := strings.Split(types, ",")
			for _, v := range cashTypes {
				v, err := strconv.Atoi(v)
				if err != nil || v < model.COTransactionReceive || v > model.COTransactionRation {
					helper.Print(ctx, false, helper.CashTypeErr)
					return
				}
			}

			if len(cashTypes) > 0 {
				ex["cash_type"] = types
			}
		}

	}

	data, err := model.RecordTransaction(page, pageSize, startTime, endTime, tableName, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *RecordController) Transfer(ctx *fasthttp.RequestCtx) {

	ty := ctx.PostArgs().GetUintOrZero("ty")
	username := string(ctx.PostArgs().Peek("username"))
	billNo := string(ctx.PostArgs().Peek("bill_no"))
	pidIn := ctx.PostArgs().GetUintOrZero("pid_in")
	pidOut := ctx.PostArgs().GetUintOrZero("pid_out")
	transferType := ctx.PostArgs().GetUintOrZero("transfer_type")
	state := ctx.PostArgs().GetUintOrZero("state")
	startTime := string(ctx.PostArgs().Peek("start_time"))
	endTime := string(ctx.PostArgs().Peek("end_time"))
	confirmName := string(ctx.PostArgs().Peek("confirm_name"))
	page := ctx.PostArgs().GetUintOrZero("page")
	pageSize := ctx.PostArgs().GetUintOrZero("page_size")

	if page == 0 {
		page = 1
	}

	if pageSize < 10 || pageSize > 200 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	t := map[int]bool{
		1: true,
		2: true,
	}
	if _, ok := t[ty]; !ok {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if ty == 1 && !validator.CheckUName(username, 5, 14) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	//查询条件
	ex := g.Ex{}
	if billNo != "" {
		ex["bill_no"] = billNo
	} else {
		if username != "" {
			ex["username"] = username
		}

		if transferType > 0 {
			if transferType < model.TransferIn || transferType > model.TransferDividend {
				helper.Print(ctx, false, errors.New(helper.TransferTypeErr))
				return
			}

			ex["transfer_type"] = transferType
		}

		if pidIn > 0 && pidOut == 0 {
			ex["platform_id"] = pidIn
		}

		if pidIn == 0 && pidOut > 0 {
			ex["platform_id"] = pidOut
		}

		if pidIn > 0 && pidOut > 0 {
			ex["platform_id"] = []int{pidIn, pidOut}
		}

		if state > 0 {
			if state < model.TransferStateFailed || state > model.TransferStateManualConfirm {
				helper.Print(ctx, false, errors.New(helper.TransferTypeErr))
				return
			}

			ex["state"] = state
		}

		if confirmName != "" {
			if !validator.CheckUName(confirmName, 5, 14) {
				helper.Print(ctx, false, errors.New(helper.UsernameErr))
			}

			ex["confirm_name"] = confirmName
		}
	}

	data, err := model.RecordTransfer(page, pageSize, startTime, endTime, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 有效投注查询/会员游戏记录详情列表/投注管理列表
func (that *RecordController) RecordGame(ctx *fasthttp.RequestCtx) {

	ty := ctx.PostArgs().GetUintOrZero("ty")
	uid := string(ctx.PostArgs().Peek("uid"))
	pid := string(ctx.PostArgs().Peek("pid"))
	platType := string(ctx.PostArgs().Peek("plat_type"))
	gameName := string(ctx.PostArgs().Peek("game_name"))
	username := string(ctx.PostArgs().Peek("username"))
	parentName := string(ctx.PostArgs().Peek("parent_name"))
	topName := string(ctx.PostArgs().Peek("top_name"))
	billNo := string(ctx.PostArgs().Peek("bill_no"))
	flag := string(ctx.PostArgs().Peek("flag"))
	gameNo := string(ctx.PostArgs().Peek("game_no"))
	presettle := string(ctx.PostArgs().Peek("presettle"))
	resettle := string(ctx.PostArgs().Peek("resettle"))
	betMin := string(ctx.PostArgs().Peek("bet_min"))
	betMax := string(ctx.PostArgs().Peek("bet_max"))
	timeFlag := string(ctx.PostArgs().Peek("time_flag"))
	startTime := string(ctx.PostArgs().Peek("start_time"))
	endTime := string(ctx.PostArgs().Peek("end_time"))
	page := ctx.PostArgs().GetUintOrZero("page")
	pageSize := ctx.PostArgs().GetUintOrZero("page_size")

	if page == 0 {
		page = 1
	}

	if pageSize < 10 || pageSize > 200 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	tf := map[string]bool{
		"1": true,
		"2": true,
		"3": true,
	}
	if _, ok := tf[timeFlag]; !ok {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if ty < 1 || ty > 6 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if (ty == model.GameTyRecordDetail || ty == model.GameTyValid || ty == model.GameMemberWinOrLose) &&
		username == "" {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	if betMin != "" && betMax != "" {
		if !validator.CheckMoney(betMin) {
			helper.Print(ctx, false, helper.AmountErr)
			return
		}

		if !validator.CheckMoney(betMax) {
			helper.Print(ctx, false, helper.AmountErr)
			return
		}
	}

	if presettle != "" {
		if !validator.CtypeDigit(presettle) {
			helper.Print(ctx, false, helper.PresettleFlagErr)
			return
		}
	}

	if resettle != "" {
		if !validator.CtypeDigit(resettle) {
			helper.Print(ctx, false, helper.PresettleFlagErr)
			return
		}
	}

	param := map[string]string{
		"uid":         uid,
		"pid":         pid,
		"plat_type":   platType,
		"game_name":   gameName,
		"username":    username,
		"top_name":    topName,
		"parent_name": parentName,
		"bill_no":     billNo,
		"flag":        flag,
		"time_flag":   timeFlag,
		"start_time":  startTime,
		"end_time":    endTime,
		"game_no":     gameNo,
		"pre_settle":  presettle,
		"resettle":    resettle,
		"bet_min":     betMin,
		"bet_max":     betMax,
	}

	if ty < model.GameMemberTransferGroup {
		data, err := model.Game(ty, pageSize, page, param)
		if err != nil {
			helper.Print(ctx, false, err.Error())
			return
		}

		helper.Print(ctx, true, data)
		return
	}

	data, err := model.GameGroup(ty, pageSize, page, param)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)

}

func (that *RecordController) Game(ctx *fasthttp.RequestCtx) {

	param := gameAdminParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	query := elastic.NewBoolQuery()
	if param.ParentName != "" {

		if !validator.CheckUName(param.ParentName, 5, 14) {
			helper.Print(ctx, false, helper.AgentNameErr)
			return
		}

		query = query.Filter(elastic.NewTermQuery("parent_name", param.ParentName))
	}

	if param.ParentName == "" {
		query.MustNot(elastic.NewTermsQuery("parent_name", "root", ""))
	}

	// 校验username
	// 如果username为空则取改代理下所有的会员
	if param.Username != "" {

		if !validator.CheckUName(param.Username, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}

		query.Filter(elastic.NewTermQuery("name", param.Username))
	}

	if param.Pid != "0" {
		query.Filter(elastic.NewTermQuery("api_type", param.Pid))
	}

	data, err := model.RecordAdminGame(
		param.Flag, param.StartTime, param.EndTime, param.Page, param.PageSize, query)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *RecordController) LoginLog(ctx *fasthttp.RequestCtx) {

	param := loginLogParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	query := elastic.NewBoolQuery().MustNot(elastic.NewTermsQuery("parents.keyword", "root", ""))
	if param.Username != "" {

		if !validator.CheckUName(param.Username, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}

		query.Filter(elastic.NewTermQuery("username", param.Username))
	}

	if param.Ip != "" {

		_, err := helper.Ip2long(param.Ip)
		if err != nil {
			helper.Print(ctx, false, helper.IPErr)
			return
		}

		query.Filter(elastic.NewTermQuery("ips", param.Ip))
	}

	data, err := model.RecordLoginLog(param.Page, param.PageSize, param.StartTime, param.EndTime, query)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *RecordController) Deposit(ctx *fasthttp.RequestCtx) {

	param := depositParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	query := elastic.NewBoolQuery()
	if param.ParentName != "" {

		if !validator.CheckUName(param.ParentName, 5, 14) {
			helper.Print(ctx, false, helper.AgentNameErr)
			return
		}

		query.Filter(elastic.NewTermQuery("parent_name", param.ParentName))
	}

	if param.ParentName == "" {
		query.MustNot(elastic.NewTermsQuery("parent_name", "root", ""))
	}

	if param.State > 0 {

		if !validator.CheckIntScope(fmt.Sprintf("%d", param.State), model.DepositConfirming, model.DepositCancelled) {
			helper.Print(ctx, false, helper.StateParamErr)
			return
		}

		query.Filter(elastic.NewTermQuery("state", param.State))
	}

	if param.State == 0 {
		query.Filter(elastic.NewTermsQuery("state", model.DepositSuccess, model.DepositCancelled))
	}

	if param.Username != "" {
		if !validator.CheckUName(param.Username, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}

		query.Filter(elastic.NewTermQuery("username", param.Username))
	}

	if param.ChannelId > 0 {

		query.Filter(elastic.NewTermQuery("channel_id", param.ChannelId))
	}

	data, err := model.RecordDeposit(param.Page, param.PageSize, param.StartTime, param.EndTime, query)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *RecordController) Dividend(ctx *fasthttp.RequestCtx) {

	param := dividendParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	query := elastic.NewBoolQuery()
	if param.ParentName != "" {

		if !validator.CheckUName(param.ParentName, 5, 14) {
			helper.Print(ctx, false, helper.AgentNameErr)
			return
		}

		query.Filter(elastic.NewTermQuery("parent_name", param.ParentName))
	}

	if param.ParentName == "" {
		query.MustNot(elastic.NewTermsQuery("parent_name", "root", ""))
	}

	if param.Ty > 0 {

		if !validator.CheckIntScope(fmt.Sprintf("%d", param.Ty), model.DividendSite, model.DividendAgency) {
			helper.Print(ctx, false, helper.StateParamErr)
			return
		}

		query.Filter(elastic.NewTermQuery("ty", param.Ty))
	}

	if param.Username != "" {

		if !validator.CheckUName(param.Username, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}

		query.Filter(elastic.NewTermQuery("username", param.Username))
	}

	data, err := model.RecordDividend(param.Page, param.PageSize, param.StartTime, param.EndTime, query)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *RecordController) Rebate(ctx *fasthttp.RequestCtx) {

	param := rebateParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	ex := g.Ex{}

	if param.Username != "" {

		if !validator.CheckStringAlnum(param.Username) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}

		ex["username"] = param.Username
	}

	data, err := model.RecordRebate(param.Page, param.PageSize, param.StartTime, param.EndTime, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *RecordController) Adjust(ctx *fasthttp.RequestCtx) {

	param := adjustRecordParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	query := elastic.NewBoolQuery()
	if param.ParentName != "" {

		if !validator.CheckUName(param.ParentName, 5, 14) {
			helper.Print(ctx, false, helper.AgentNameErr)
			return
		}

		query.Filter(elastic.NewTermQuery("parent_name", param.ParentName))
	}

	if param.ParentName == "" {
		query.MustNot(elastic.NewTermsQuery("parent_name", "root", ""))
	}

	if param.Username != "" {

		if !validator.CheckUName(param.Username, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}

		query.Filter(elastic.NewTermQuery("username", param.Username))
	}

	if param.State > 0 {

		if param.State < model.AdjustFailed || param.State > model.AdjustPlatDealing {
			helper.Print(ctx, false, helper.StateParamErr)
			return
		}

		query.Filter(elastic.NewTermQuery("hand_out_state", param.State))
	}

	if param.AdjustType != "0" {

		if !validator.CheckIntScope(param.AdjustType, 1, 3) {
			helper.Print(ctx, false, helper.AdjustTyErr)
			return
		}

		query.Filter(elastic.NewTermQuery("adjust_type", param.AdjustType))
	}

	data, err := model.RecordAdjust(param.Page, param.PageSize, param.StartTime, param.EndTime, query)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 代理管理-记录管理-提款
func (that *RecordController) Withdraw(ctx *fasthttp.RequestCtx) {

	param := withdrawParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	query := elastic.NewBoolQuery()
	if param.ParentName != "" {

		if !validator.CheckUName(param.ParentName, 5, 14) {
			helper.Print(ctx, false, helper.AgentNameErr)
			return
		}

		query.Filter(elastic.NewTermQuery("parent_name", param.ParentName))
	}

	if param.ParentName == "" {
		query.MustNot(elastic.NewTermsQuery("parent_name", "root", ""))
	}

	if param.State > 0 {

		if !validator.CheckIntScope(fmt.Sprintf("%d", param.State), model.WithdrawReviewing, model.WithdrawDispatched) {
			helper.Print(ctx, false, helper.StateParamErr)
			return
		}

		query.Filter(elastic.NewTermQuery("state", param.State))
	}

	if param.MinAmount != "0" || param.MaxAmount != "0" {

		state, err := validator.CheckAmountRange(param.MinAmount, param.MaxAmount)
		if err != nil || state == -1 {
			helper.Print(ctx, false, helper.AmountErr)
			return
		}

		min, err := decimal.NewFromString(param.MinAmount)
		if err != nil {
			helper.Print(ctx, false, helper.AmountErr)
			return
		}

		max, err := decimal.NewFromString(param.MaxAmount)
		if err != nil {
			helper.Print(ctx, false, helper.AmountErr)
			return
		}

		minVal, _ := min.Float64()
		maxVal, _ := max.Float64()

		query.Filter(elastic.NewRangeQuery("amount").Gte(minVal).Lte(maxVal))
	}

	if param.Username != "" {

		if !validator.CheckUName(param.Username, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}

		query.Filter(elastic.NewTermQuery("username", param.Username))
	}

	data, err := model.RecordWithdraw(param.Page,
		param.PageSize, param.StartTime, param.EndTime, param.ApplyStartTime, param.ApplyEndTime, query)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	rs, err := model.WithdrawDealListData(data)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, rs)
}

func (that *RecordController) Group(ctx *fasthttp.RequestCtx) {

	param := groupParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	ex := g.Ex{}
	if param.Username != "" {

		if !validator.CheckUName(param.Username, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}

		ex["username"] = param.Username
	}

	if param.Uid != "" {

		if !validator.CheckStringDigit(param.Uid) {
			helper.Print(ctx, false, helper.IDErr)
			return
		}

		ex["uid"] = param.Uid
	}

	if param.ParentName != "" {

		//orEx := g.Or(
		//	g.Ex{"after_name": param.ParentName},
		//	g.Ex{"before_name": param.ParentName},
		//)
		//
		//g.And(ex, orEx)

		ex["before_name"] = param.ParentName
	}

	data, err := model.RecordGroup(param.Page, param.PageSize, param.StartTime, param.EndTime, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}
