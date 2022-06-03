package controller

import (
	"merchant/contrib/helper"
	"merchant/model"

	"github.com/valyala/fasthttp"
)

type SmsRecordController struct{}

// List 验证码列表
func (that *SmsRecordController) List(ctx *fasthttp.RequestCtx) {

	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")
	username := string(ctx.QueryArgs().Peek("username"))
	phone := string(ctx.QueryArgs().Peek("phone"))
	state := string(ctx.QueryArgs().Peek("state"))
	ty := string(ctx.QueryArgs().Peek("ty"))
	startTime := string(ctx.QueryArgs().Peek("start_time"))
	endTime := string(ctx.QueryArgs().Peek("end_time"))

	if page < 1 {
		page = 1
	}
	if pageSize > 50 {
		pageSize = 50
	}
	if pageSize < 20 {
		pageSize = 20
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

	data, err := model.SmsList(uint(page), uint(pageSize), startTime, endTime, username, phone, state, ty)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)

}
