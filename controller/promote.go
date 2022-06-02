package controller

import (
	"github.com/valyala/fasthttp"
	"merchant/contrib/helper"
	"merchant/contrib/validator"
	"merchant/model"
	"strconv"
	"strings"
)

type PromoteController struct{}

// 推广域名统计信息
func (that *PromoteController) List(ctx *fasthttp.RequestCtx) {

	page := string(ctx.PostArgs().Peek("page"))
	pageSize := string(ctx.PostArgs().Peek("page_size"))
	startTime := string(ctx.PostArgs().Peek("start_time"))
	endTime := string(ctx.PostArgs().Peek("end_time"))
	url := string(ctx.PostArgs().Peek("urls"))
	ty := string(ctx.PostArgs().Peek("ty"))

	if !validator.CheckIntScope(ty, model.UrlTyOfficial, model.UrlTyTg) {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	var urls []string
	if len(url) > 0 {
		urls = strings.Split(url, ",")
	}

	cpage, err := strconv.Atoi(page)
	if err != nil {
		cpage = 1
	}

	cpageSize, err := strconv.Atoi(pageSize)
	if err != nil {
		cpageSize = 10
	}

	tyNum, _ := strconv.Atoi(ty)
	data, err := model.PromoteInfoList(tyNum, urls, startTime, endTime, cpage, cpageSize)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 推广域名关联的ip信息
func (that *PromoteController) IPList(ctx *fasthttp.RequestCtx) {

	url := string(ctx.PostArgs().Peek("url"))
	page := string(ctx.PostArgs().Peek("page"))
	pageSize := string(ctx.PostArgs().Peek("page_size"))
	startTime := string(ctx.PostArgs().Peek("start_time"))
	endTime := string(ctx.PostArgs().Peek("end_time"))

	if len(url) == 0 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	cpage, err := strconv.Atoi(page)
	if err != nil {
		cpage = 1
	}

	cpageSize, err := strconv.Atoi(pageSize)
	if err != nil {
		cpageSize = 10
	}

	data, err := model.PromoteIPList(url, startTime, endTime, cpage, cpageSize)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 推广域名关联的会员信息
func (that *PromoteController) MemberList(ctx *fasthttp.RequestCtx) {

	url := string(ctx.PostArgs().Peek("url"))
	page := string(ctx.PostArgs().Peek("page"))
	pageSize := string(ctx.PostArgs().Peek("page_size"))
	startTime := string(ctx.PostArgs().Peek("start_time"))
	endTime := string(ctx.PostArgs().Peek("end_time"))

	if len(url) == 0 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	cpage, err := strconv.Atoi(page)
	if err != nil {
		cpage = 1
	}

	cpageSize, err := strconv.Atoi(pageSize)
	if err != nil {
		cpageSize = 10
	}

	data, err := model.PromoteMemberList(url, startTime, endTime, uint(cpage), uint(cpageSize))
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)

}
