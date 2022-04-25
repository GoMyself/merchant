package controller

import (
	"github.com/valyala/fasthttp"
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"merchant2/model"
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
		if !validator.CheckUName(username, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}
	}

	// 手机号校验
	if phone != "" {
		if !validator.IsVietnamesePhone(phone) {
			helper.Print(ctx, false, helper.PhoneFMTErr)
			return
		}
	}

	startAt := ctx.Time().Unix()
	data, err := model.VerifyCodeList(username, phone, startAt)
	if err != nil {
		helper.Print(ctx, false, err)
		return
	}

	helper.PrintJson(ctx, true, data)

}
