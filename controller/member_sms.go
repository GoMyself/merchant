package controller

import (
	"merchant2/contrib/helper"
	"merchant2/model"

	"github.com/valyala/fasthttp"
)

type SmsRecordController struct{}

// List 验证码列表
func (that *SmsRecordController) List(ctx *fasthttp.RequestCtx) {

	username := string(ctx.PostArgs().Peek("username"))
	phone := string(ctx.PostArgs().Peek("phone"))
	if username == "" && phone == "" {
		helper.Print(ctx, false, helper.ParamNull)
		return
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

	data, err := model.SmsList(username, phone)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.PrintJson(ctx, true, data)

}
