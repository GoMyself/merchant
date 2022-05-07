package controller

import (
	"errors"
	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"merchant2/model"
)

type MemberTransferController struct{}

func transferRebateRateCheck(mb, destMb model.Member) error {

	src, err := model.MemberRebateFindOne(mb.UID)
	if err != nil {
		return err
	}

	dest, err := model.MemberRebateFindOne(destMb.UID)
	if err != nil {
		return err
	}

	if src.TY.GreaterThan(dest.TY) || //体育返水比例
		src.ZR.GreaterThan(dest.ZR) || //真人返水比例
		src.QP.GreaterThan(dest.QP) || //棋牌返水比例
		src.DJ.GreaterThan(dest.DJ) || //电竞返水比例
		src.DZ.GreaterThan(dest.DZ) || //电子返水比例
		src.CP.GreaterThan(dest.CP) { //彩票返水比例
		return errors.New(helper.RebateOutOfRange)
	}

	return nil
}

// Transfer  跳线转代
func (that *MemberTransferController) Transfer(ctx *fasthttp.RequestCtx) {

	username := string(ctx.PostArgs().Peek("username"))
	destName := string(ctx.PostArgs().Peek("dest_name"))

	if username == destName {
		helper.Print(ctx, false, helper.TransferToAgencyErr)
		return
	}

	// 已有下线，不允许使用跳线转代
	if model.MemberTransferSubCheck(username) {
		helper.Print(ctx, false, helper.MemberHaveSubAlready)
		return
	}

	mb, err := model.MemberFindOne(username)
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	if mb.ParentName == destName {
		helper.Print(ctx, false, helper.IsAgentSubAlready)
		return
	}

	destMb, err := model.MemberFindOne(destName)
	if err != nil {
		helper.Print(ctx, false, helper.AgentNameErr)
		return
	}

	err = transferRebateRateCheck(mb, destMb)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	err = model.MemberTransferAg(mb, destMb, admin)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// List  团队转代申请列表
func (that *MemberTransferController) List(ctx *fasthttp.RequestCtx) {

	//id := string(ctx.QueryArgs().Peek("id"))
	//page := ctx.QueryArgs().GetUintOrZero("page")
	//pageSize := ctx.QueryArgs().GetUintOrZero("page_size")
	//flag := ctx.QueryArgs().GetUintOrZero("flag")                        //1 审核列表 2 历史记录
	//username := string(ctx.QueryArgs().Peek("username"))                 //会员名
	//afterName := string(ctx.QueryArgs().Peek("after_name"))              //转以后代理名
	//applyName := string(ctx.QueryArgs().Peek("apply_name"))              //申请人名
	//reviewName := string(ctx.QueryArgs().Peek("review_name"))            //审核人名
	//startTime := string(ctx.QueryArgs().Peek("start_time"))              //申请开始时间
	//endTime := string(ctx.QueryArgs().Peek("end_time"))                  //申请结束时间
	//reviewStartTime := string(ctx.QueryArgs().Peek("review_start_time")) //审核开始时间
	//reviewEndTime := string(ctx.QueryArgs().Peek("review_end_time"))     //审核结束时间

	id := string(ctx.PostArgs().Peek("id"))
	page := ctx.PostArgs().GetUintOrZero("page")
	pageSize := ctx.PostArgs().GetUintOrZero("page_size")
	flag := ctx.PostArgs().GetUintOrZero("flag")                        //1 审核列表 2 历史记录
	username := string(ctx.PostArgs().Peek("username"))                 //会员名
	afterName := string(ctx.PostArgs().Peek("after_name"))              //转以后代理名
	applyName := string(ctx.PostArgs().Peek("apply_name"))              //申请人名
	reviewName := string(ctx.PostArgs().Peek("review_name"))            //审核人名
	startTime := string(ctx.PostArgs().Peek("start_time"))              //申请开始时间
	endTime := string(ctx.PostArgs().Peek("end_time"))                  //申请结束时间
	reviewStartTime := string(ctx.PostArgs().Peek("review_start_time")) //审核开始时间
	reviewEndTime := string(ctx.PostArgs().Peek("review_end_time"))     //审核结束时间

	ex := g.Ex{}
	if page == 0 {
		page = 1
	}
	if pageSize < 10 {
		page = 10
	}

	if id == "" {
		flags := map[int]bool{
			1: true,
			2: true,
		}
		if _, ok := flags[flag]; !ok {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}

		if flag == 1 {
			ex["status"] = 1
		} else {
			ex["status"] = []int{2, 3, 4}
		}

		if username != "" {
			if !validator.CheckUName(username, 5, 14) {
				helper.Print(ctx, false, helper.UsernameErr)
				return
			}

			ex["username"] = username
		}

		if afterName != "" {
			if !validator.CheckUName(afterName, 5, 14) {
				helper.Print(ctx, false, helper.AgentNameErr)
				return
			}

			ex["after_name"] = afterName
		}

		if applyName != "" {
			if !validator.CheckAName(applyName, 5, 20) {
				helper.Print(ctx, false, helper.AdminNameErr)
				return
			}

			ex["apply_name"] = applyName
		}

		if reviewName != "" {
			if !validator.CheckAName(reviewName, 5, 20) {
				helper.Print(ctx, false, helper.AdminNameErr)
				return
			}

			ex["review_name"] = reviewName
		}
	} else {
		ex = g.Ex{
			"id": id,
		}
	}

	data, err := model.MemberTransferList(page, pageSize, startTime, endTime, reviewStartTime, reviewEndTime, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// Insert  团队转代
func (that *MemberTransferController) Insert(ctx *fasthttp.RequestCtx) {

	username := string(ctx.PostArgs().Peek("username"))
	destName := string(ctx.PostArgs().Peek("dest_name"))
	remark := string(ctx.PostArgs().Peek("remark"))
	if username == destName {
		helper.Print(ctx, false, helper.TransferToAgencyErr)
		return
	}

	mb, err := model.MemberFindOne(username)
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	destMb, err := model.MemberFindOne(destName)
	if err != nil {
		helper.Print(ctx, false, helper.AgentNameErr)
		return
	}

	err = transferRebateRateCheck(mb, destMb)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	// 没有下线，相当于跳线转代
	if !model.MemberTransferSubCheck(username) {
		err = model.MemberTransferAg(mb, destMb, admin)
	} else {
		if model.MemberTransferExist(mb.Username) {
			helper.Print(ctx, false, helper.TransferApplyExist)
			return
		}

		err = model.MemberTransferInsert(mb, destMb, admin, remark)
	}
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// Review  团队转代申请审核
func (that *MemberTransferController) Review(ctx *fasthttp.RequestCtx) {

	id := string(ctx.PostArgs().Peek("id"))
	status := ctx.PostArgs().GetUintOrZero("status")
	reviewRemark := string(ctx.PostArgs().Peek("review_remark"))

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	err = model.MemberTransferReview(id, reviewRemark, status, admin)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// Delete  团队转代申请删除
func (that *MemberTransferController) Delete(ctx *fasthttp.RequestCtx) {

	id := string(ctx.QueryArgs().Peek("id"))

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	err = model.MemberTransferDelete(id, admin)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
