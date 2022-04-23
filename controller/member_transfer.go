package controller

import (
	"errors"
	"github.com/valyala/fasthttp"
	"merchant2/contrib/helper"
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

	err = model.MemberTransferAg(mb, destMb)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// List  团队转代申请列表
func (that *MemberTransferController) List(ctx *fasthttp.RequestCtx) {

}

// Insert  团队转代
func (that *MemberTransferController) Insert(ctx *fasthttp.RequestCtx) {

	username := string(ctx.PostArgs().Peek("username"))
	destName := string(ctx.PostArgs().Peek("dest_name"))
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

	// 没有下线，相当于跳线转代
	if !model.MemberTransferSubCheck(username) {
		err = model.MemberTransferAg(mb, destMb)
	} else {
		err = model.MemberTransferInsert(mb, destMb)
	}
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// Review  团队转代申请审核
func (that *MemberTransferController) Review(ctx *fasthttp.RequestCtx) {

}

// Delete  团队转代申请删除
func (that *MemberTransferController) Delete(ctx *fasthttp.RequestCtx) {

}
