package controller

import (
	"fmt"
	"merchant/contrib/helper"
	"merchant/model"

	"github.com/valyala/fasthttp"
)

type RebateController struct{}

func (that *RebateController) Scale(ctx *fasthttp.RequestCtx) {

	vs := model.MemberRebateScale()
	s := fmt.Sprintf(
		`{"ty":"%s","zr":"%s","dj":"%s","qp":"%s","dz":"%s","cp":"%s","fc":"%s","by":"%s","cg_official_rebate":"%s","cg_high_rebate":"%s"}`,
		vs.TY.StringFixed(1),
		vs.ZR.StringFixed(1),
		vs.DJ.StringFixed(1),
		vs.QP.StringFixed(1),
		vs.DZ.StringFixed(1),
		vs.CP.StringFixed(1),
		vs.FC.StringFixed(1),
		vs.BY.StringFixed(1),
		vs.CGOfficialRebate.StringFixed(2),
		vs.CGHighRebate.StringFixed(2),
	)

	helper.PrintJson(ctx, true, s)
}

func (that *RebateController) EnableMod(ctx *fasthttp.RequestCtx) {

	enable := ctx.QueryArgs().GetBool("enable")
	err := model.MemberRebateEnableMod(enable)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
