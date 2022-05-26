package controller

import (
	"merchant2/contrib/helper"
	"merchant2/model"

	"github.com/valyala/fasthttp"
)

type SmsRecordController struct{}

// List 验证码列表
func (that *SmsRecordController) List(ctx *fasthttp.RequestCtx) {

	var size uint = 10

	page := ctx.QueryArgs().GetUintOrZero("page")
	username := string(ctx.QueryArgs().Peek("username"))
	phone := string(ctx.QueryArgs().Peek("phone"))
	state := string(ctx.QueryArgs().Peek("state"))
	startTime := string(ctx.QueryArgs().Peek("start_time"))
	endTime := string(ctx.QueryArgs().Peek("end_time"))

	if page < 1 {
		page = 1
	}
	// 会员名校验
	//if username != "" {
	//	if !helper.CtypeAlnum(username) {
	//		helper.Print(ctx, false, helper.UsernameErr)
	//		return
	//	}
	//}

	//// 手机号校验
	//if phone != "" {
	//	if !helper.CtypeDigit(phone) {
	//		helper.Print(ctx, false, helper.PhoneFMTErr)
	//		return
	//	}
	//}
	//
	//if state == "" {
	//	helper.Print(ctx, false, helper.PhoneFMTErr)
	//	return
	//}

	if startTime == "" || endTime == "" {
		helper.Print(ctx, false, helper.DateTimeErr)
		return
	}

	data, err := model.SmsList(uint(page), size, startTime, endTime, username, phone, state)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)

}
