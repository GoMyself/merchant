package controller

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"merchant2/contrib/helper"
	"merchant2/model"
)

type RebateController struct{}

func (that *RebateController) Scale(ctx *fasthttp.RequestCtx) {

	vs := model.RebateScale()
	s := fmt.Sprintf(
		`{"ty":"%s","zr":"%s","dj":"%s","qp":"%s","dz":"%s"}`,
		vs.TY.StringFixed(1),
		vs.ZR.StringFixed(1),
		vs.DJ.StringFixed(1),
		vs.QP.StringFixed(1),
		vs.DZ.StringFixed(1),
	)

	helper.PrintJson(ctx, true, s)
}
