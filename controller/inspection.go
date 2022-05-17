package controller

import (
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

	data, err := model.InspectionList(username)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}
