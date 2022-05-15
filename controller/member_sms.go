package controller

import (
	"merchant2/contrib/helper"
	"merchant2/model"

	"github.com/valyala/fasthttp"
)

type SmsRecordController struct{}

// List 验证码列表
func (that *SmsRecordController) List(ctx *fasthttp.RequestCtx) {

	var size uint = 50

	page := ctx.QueryArgs().GetUintOrZero("page")
	username := string(ctx.QueryArgs().Peek("username"))
	phone := string(ctx.QueryArgs().Peek("phone"))

	if username == "" && phone == "" {
		helper.Print(ctx, false, helper.ParamNull)
		return
	}

	if page < 1 {
		page = 1
	}
	// 会员名校验
	if username != "" {
		if !helper.CtypeAlnum(username) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}
	}

	// 手机号校验
	if phone != "" {
		if !helper.CtypeDigit(phone) {
			helper.Print(ctx, false, helper.PhoneFMTErr)
			return
		}
	}

	data, err := model.SmsList(uint(page), size, username, phone)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)

}
