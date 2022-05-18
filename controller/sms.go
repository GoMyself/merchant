package controller

import (
	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"merchant2/model"
)

// SMSChannelController 会员端接口
type SMSChannelController struct{}

// List 短信通道列表及按 渠道名称，创建人 筛选
func (*SMSChannelController) List(ctx *fasthttp.RequestCtx) {

	channelName := string(ctx.PostArgs().Peek("name"))
	createdName := string(ctx.PostArgs().Peek("created_name"))

	ex := g.Ex{}

	if len(channelName) < 5 || len(channelName) > 25 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}
	ex["name"] = channelName

	if createdName != "" && !validator.CheckAName(createdName, 5, 20) {
		helper.Print(ctx, false, helper.AdminNameErr)
		return
	}
	ex["created_name"] = createdName

	list, err := model.SMSChannelList(ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, list)
}

func (*SMSChannelController) UpdateState(ctx *fasthttp.RequestCtx) {

	id := string(ctx.PostArgs().Peek("id"))        // 短信通道ID
	state := ctx.PostArgs().GetUintOrZero("state") // 短信通道状态

	if !validator.CtypeDigit(id) {
		helper.Print(ctx, false, helper.DBErr)
		return
	}

	if state < 0 || state > 4 {
		helper.Print(ctx, false, helper.StateParamErr)
		return
	}

	err := model.SMSChannelUpdateState(id, state)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, "success")
}
