package controller

import (
	"errors"
	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"merchant2/model"
	"strconv"
	"strings"
)

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

	if ty == 1 && !validator.CheckUName(username, 4, 9) {
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
			if !validator.CheckUName(confirmName, 4, 9) {
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
