package controller

import (
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"merchant2/model"
)

type SmsRecordController struct{}

// List 验证码列表
func (that *SmsRecordController) List(ctx *fasthttp.RequestCtx) {

	username := string(ctx.PostArgs().Peek("username"))
	phone := string(ctx.PostArgs().Peek("phone")) //手机号
	startTime := string(ctx.PostArgs().Peek("start_time"))
	endTime := string(ctx.PostArgs().Peek("end_time"))
	page := ctx.PostArgs().GetUintOrZero("page")
	pageSize := ctx.PostArgs().GetUintOrZero("page_size")

	if username == "" || !validator.CheckUName(username, 5, 14) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	ex := g.Ex{
		"username": username,
	}

	if phone != "" {
		if !validator.CheckStringDigit(phone) {
			helper.Print(ctx, false, helper.PhoneFMTErr)
			return
		}

		ex["phone_hash"] = fmt.Sprintf("%d", model.MurmurHash(phone, 0))
	}

	if page == 0 {
		page = 1
	}

	if pageSize == 0 {
		pageSize = 15
	}

	data, err := model.SmsRecordList(page, pageSize, startTime, endTime, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}
