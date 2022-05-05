package controller

import (
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"merchant2/model"
	"strconv"
	"strings"

	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/valyala/fasthttp"
)

type CommissionController struct{}

// 红利审核列表参数
type commissionUpdateParam struct {
	IDS          string `rule:"sDigit" name:"ids" msg:"ids error"`                               // 订单号
	ReviewRemark string `rule:"filter" name:"review_remark" max:"300" msg:"review_remark error"` // 审核备注
	State        int    `name:"state" rule:"digit" min:"2" max:"3" msg:"state error"`            // 2 审核通过  3 审核不通过
}

type commissionRationParam struct {
	IDS string `rule:"sDigit" name:"ids" msg:"ids error"` // 订单号
}

func (that *CommissionController) TopList(ctx *fasthttp.RequestCtx) {

	username := string(ctx.QueryArgs().Peek("username"))     //代理账号
	planId := string(ctx.QueryArgs().Peek("plan_id"))        //返佣方案
	day := string(ctx.QueryArgs().Peek("day"))               //佣金月份
	activeMax := ctx.QueryArgs().GetUintOrZero("active_max") //1 审核中 2 审核通过 3 审核不通过
	activeMin := ctx.QueryArgs().GetUintOrZero("active_min") //1 自动 2 手动
	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")
	sortField := string(ctx.QueryArgs().Peek("sort_field"))
	isAsc := ctx.QueryArgs().GetUintOrZero("is_asc")
	if page == 0 {
		page = 1
	}

	if pageSize < 10 || pageSize > 200 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	ex := g.Ex{
		"state": 1,
	}

	if len(username) > 0 {
		ex["username"] = username
	}
	if len(planId) > 0 {
		ex["plan_id"] = planId
	}
	if activeMax > 0 && activeMin > 0 {
		ex["active_num"] = g.Op{"between": exp.NewRangeVal(activeMin, activeMax)}
	}
	if sortField != "" {
		sortFields := map[string]bool{
			"deposit_amount":     true,
			"withdraw_amount":    true,
			"win_amount":         true,
			"platform_amount":    true,
			"rebate_amount":      true,
			"dividend_amount":    true,
			"adjust_amount":      true,
			"net_win":            true,
			"balance_amount":     true,
			"adjust_commission":  true,
			"adjust_win":         true,
			"amount":             true,
			"dividend_ag_amount": true,
			"last_month_amount":  true,
		}

		if _, ok := sortFields[sortField]; !ok {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}

		if !validator.CheckIntScope(strconv.Itoa(isAsc), 0, 1) {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}
	}

	data, err := model.TopCommissionList(sortField, isAsc, page, pageSize, day, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

//发放佣金
func (that *CommissionController) Ration(ctx *fasthttp.RequestCtx) {

	param := commissionRationParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	ids := strings.Split(param.IDS, ",")
	if len(ids) == 0 {
		helper.Print(ctx, false, helper.NoDataUpdate)
		return
	}

	err = model.CommissionRation(ctx.Time().Unix(), admin["id"], admin["name"], ids)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *CommissionController) RecordList(ctx *fasthttp.RequestCtx) {

	username := string(ctx.QueryArgs().Peek("username"))                 //发放账号
	receiveName := string(ctx.QueryArgs().Peek("receive_name"))          //接收账号
	reviewName := string(ctx.QueryArgs().Peek("review_name"))            //审核账号
	transferType := ctx.QueryArgs().GetUintOrZero("transfer_type")       //类型 1佣金发放 2 佣金提取 3佣金下发
	startTime := string(ctx.QueryArgs().Peek("start_time"))              //创建开始时间
	endTime := string(ctx.QueryArgs().Peek("end_time"))                  //创建结束时间
	reviewStartTime := string(ctx.QueryArgs().Peek("review_start_time")) //审核开始时间
	reviewEndTime := string(ctx.QueryArgs().Peek("review_end_time"))     //审核结束时间
	state := ctx.QueryArgs().GetUintOrZero("state")                      //1 审核中 2 审核通过 3 审核不通过
	automatic := ctx.QueryArgs().GetUintOrZero("automatic")              //1 自动 2 手动
	flag := ctx.QueryArgs().GetUintOrZero("flag")                        //0 所有 1 审核列表 2历史列表
	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")

	if page == 0 {
		page = 1
	}

	if pageSize < 10 || pageSize > 200 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	ex := g.Ex{}

	// 接收会员
	if receiveName != "" {
		if !validator.CheckUName(receiveName, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}

		ex["receive_name"] = receiveName
	}

	// 审核管理员
	if reviewName != "" {
		if validator.CheckUName(reviewName, 5, 20) {
			helper.Print(ctx, false, helper.AdminNameErr)
			return
		}

		ex["review_name"] = reviewName
	}

	t := map[int]bool{
		1: true, //佣金发放
		2: true, //佣金提取
		3: true, //佣金下发
	}
	if transferType > 0 {
		if _, ok := t[transferType]; !ok {
			helper.Print(ctx, false, helper.TransferTypeErr)
			return
		}

		ex["transfer_type"] = transferType
	}

	// 可能为会员账号，也可能是后台账号
	if username != "" {
		if validator.CheckUName(username, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}
		ex["username"] = username
	}

	// 红利审核列表
	if flag == 1 {
		ex["state"] = 1 //审核中
	} else { //
		// 默认为红利历史列表
		s := map[int]bool{
			2: true, //审核通过
			3: true, //审核不通过
		}

		// 查询所有
		if state > 0 {
			if _, ok := s[state]; !ok {
				helper.Print(ctx, false, helper.StateParamErr)
				return
			}

			ex["state"] = state
		} else {
			ex["state"] = []int{2, 3}
		}
	}

	a := map[int]bool{
		1: true, //自动
		2: true, //手动
	}
	if automatic > 0 {
		if _, ok := a[automatic]; !ok {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}

		ex["automatic"] = automatic
	}

	data, err := model.CommissionRecordList(page, pageSize, startTime, endTime, reviewStartTime, reviewEndTime, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 佣金记录审核
func (that *CommissionController) RecordReview(ctx *fasthttp.RequestCtx) {

	param := commissionUpdateParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	ids := strings.Split(param.IDS, ",")
	if len(ids) == 0 {
		helper.Print(ctx, false, helper.NoDataUpdate)
		return
	}

	s := map[int]bool{
		2: true,
		3: true,
	}
	if _, ok := s[param.State]; !ok {
		helper.Print(ctx, false, helper.StateParamErr)
		return
	}
	err = model.CommissionRecordReview(param.State, ctx.Time().Unix(), admin["id"], admin["name"], param.ReviewRemark, ids)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *CommissionController) PlanInsert(ctx *fasthttp.RequestCtx) {

	name := string(ctx.PostArgs().Peek("name"))
	commissionMonth := string(ctx.PostArgs().Peek("commission_month"))
	content := ctx.PostArgs().Peek("content")

	if name == "" {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	err := model.CommissionPlanInsert(ctx, name, commissionMonth, content)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *CommissionController) PlanUpdate(ctx *fasthttp.RequestCtx) {

	id := string(ctx.PostArgs().Peek("id"))
	name := string(ctx.PostArgs().Peek("name"))
	commissionMonth := string(ctx.PostArgs().Peek("commission_month"))
	content := ctx.PostArgs().Peek("content")

	if name == "" || id == "" {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	_, err := model.CommissionPlanFind(g.Ex{"id": id})
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	err = model.CommissionPlanUpdate(ctx, id, name, commissionMonth, content)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// PlanList 佣金方案列表
func (that *CommissionController) PlanList(ctx *fasthttp.RequestCtx) {

	username := string(ctx.PostArgs().Peek("username"))
	name := string(ctx.PostArgs().Peek("name"))                         // 方案名称
	startTime := string(ctx.PostArgs().Peek("start_time"))              //
	endTime := string(ctx.PostArgs().Peek("end_time"))                  //
	updateStartTime := string(ctx.PostArgs().Peek("update_start_time")) //
	updateEndTime := string(ctx.PostArgs().Peek("update_end_time"))     //
	page := ctx.PostArgs().GetUintOrZero("page")
	pageSize := ctx.PostArgs().GetUintOrZero("page_size")

	if page == 0 {
		page = 1
	}

	if pageSize < 10 || pageSize > 200 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	ex := g.Ex{}

	if username != "" {
		if !validator.CheckAName(username, 5, 20) {
			helper.Print(ctx, false, helper.AdminNameErr)
			return
		}

		ex["updated_name"] = username

	}

	if name != "" {
		ex["name"] = name
	}

	data, err := model.CommissionPlanList(ex, startTime, endTime, updateStartTime, updateEndTime, page, pageSize)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// PlanDetail 佣金方案详情
func (that *CommissionController) PlanDetail(ctx *fasthttp.RequestCtx) {

	id := string(ctx.QueryArgs().Peek("id"))

	ex := g.Ex{}

	if id == "" {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	ex["plan_id"] = id
	data, err := model.CommissionPlanDetail(ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}
