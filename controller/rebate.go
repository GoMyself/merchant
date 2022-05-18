package controller

import (
	"fmt"
	"merchant2/contrib/helper"
	"merchant2/model"

	"github.com/valyala/fasthttp"
)

type RebateController struct{}

func (that *RebateController) Scale(ctx *fasthttp.RequestCtx) {

	vs := model.RebateScale()
	s := fmt.Sprintf(
		`{"ty":"%s","zr":"%s","dj":"%s","qp":"%s","dz":"%s","cp":"%s","fc":"%s","cg_official_rebate":"%s","cg_high_rebate":"%s"}`,
		vs.TY.StringFixed(1),
		vs.ZR.StringFixed(1),
		vs.DJ.StringFixed(1),
		vs.QP.StringFixed(1),
		vs.DZ.StringFixed(1),
		vs.CP.StringFixed(1),
		vs.FC.StringFixed(1),
		vs.CGOfficialRebate.StringFixed(2),
		vs.CGHighRebate.StringFixed(2),
	)

	helper.PrintJson(ctx, true, s)
}
