package controller

import (
	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"
	"merchant/contrib/helper"
	"merchant/contrib/validator"
	"merchant/model"
)

type LinkController struct{}

func (that *LinkController) List(ctx *fasthttp.RequestCtx) {

	page := ctx.PostArgs().GetUintOrZero("page")
	pageSize := ctx.PostArgs().GetUintOrZero("page_size")
	username := string(ctx.QueryArgs().Peek("username"))
	shortURL := string(ctx.QueryArgs().Peek("short_url"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 15
	}
	ex := g.Ex{}
	if username != "" {
		if !validator.CheckUName(username, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}

		ex["username"] = username
	}

	if shortURL != "" {
		ex["short_url"] = shortURL
	}

	data, err := model.LinkList(uint(page), uint(pageSize), ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *LinkController) Delete(ctx *fasthttp.RequestCtx) {

	username := string(ctx.QueryArgs().Peek("username"))
	id := string(ctx.QueryArgs().Peek("id"))
	if !helper.CtypeDigit(id) {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	mb, err := model.MemberInfo(username)
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	err = model.LinkDelete(mb.UID, id)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
