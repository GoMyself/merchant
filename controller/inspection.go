package controller

import (
	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"
	"merchant2/contrib/helper"
	"merchant2/model"
)

type InspectionController struct{}

//List 稽查列表
func (that *InspectionController) List(ctx *fasthttp.RequestCtx) {

	username := string(ctx.QueryArgs().Peek("username"))

	if len(username) == 0 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	data, _, err := model.InspectionList(username)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

//Review 稽查审核
func (that *InspectionController) Review(ctx *fasthttp.RequestCtx) {

	username := string(ctx.PostArgs().Peek("username"))
	inspectState := string(ctx.PostArgs().Peek("state"))
	billNo := string(ctx.PostArgs().Peek("bill_no"))
	remark := string(ctx.PostArgs().Peek("remark"))

	if len(username) == 0 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if len(billNo) == 0 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	data, err := model.InspectionReview(username, inspectState, billNo, remark, admin)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *InspectionController) History(ctx *fasthttp.RequestCtx) {

	username := string(ctx.PostArgs().Peek("username"))
	inspectState := string(ctx.PostArgs().Peek("state"))
	billNo := string(ctx.PostArgs().Peek("bill_no"))

	reviewName := string(ctx.PostArgs().Peek("review_name"))
	inspectName := string(ctx.PostArgs().Peek("inspect_name"))
	page := string(ctx.PostArgs().GetUintOrZero("page"))
	pageSize := string(ctx.PostArgs().GetUintOrZero("page_size"))
	ex := g.Ex{}
	if len(username) != 0 {
		ex["username"] = username
	}

	if len(billNo) != 0 {
		ex["bill_no"] = billNo
	}

	if len(inspectState) != 0 {
		ex["state"] = inspectState
	}

	if len(reviewName) != 0 {
		ex["review_name"] = reviewName
	}

	if len(inspectName) != 0 {
		ex["inspect_name"] = inspectName
	}

	data, err := model.InspectionHistory(ex, page, pageSize)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}
