package controller

import (
	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"merchant2/model"
	"strings"
)

type MessageController struct{}

// 站内信新增
func (that *MessageController) Insert(ctx *fasthttp.RequestCtx) {

	title := string(ctx.PostArgs().Peek("title"))        //标题
	subTitle := string(ctx.PostArgs().Peek("sub_title")) //副标题
	content := string(ctx.PostArgs().Peek("content"))    //内容
	ty := ctx.PostArgs().GetUintOrZero("ty")             //1站内消息 2活动消息
	isTop := ctx.PostArgs().GetUintOrZero("is_top")      //0不置顶 1置顶
	isPush := ctx.PostArgs().GetUintOrZero("is_push")    //0不推送 1推送
	sendName := string(ctx.PostArgs().Peek("send_name")) //发送人名
	sendAt := string(ctx.PostArgs().Peek("send_at"))     //发送时间
	isVip := ctx.PostArgs().GetUintOrZero("is_vip")      //是否vip站内信 1 vip站内信
	level := string(ctx.PostArgs().Peek("level"))        //vip等级 0-10,多个逗号分割
	names := string(ctx.PostArgs().Peek("names"))        //会员名，多个用逗号分割

	if len(title) == 0 ||
		len(subTitle) == 0 ||
		len(content) == 0 ||
		len(sendName) == 0 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if ty != 1 && ty != 2 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if isVip == 1 {
		if level == "" {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}

		lv := map[string]bool{
			"1":  true,
			"2":  true,
			"3":  true,
			"4":  true,
			"5":  true,
			"6":  true,
			"7":  true,
			"8":  true,
			"9":  true,
			"10": true,
			"11": true,
		}
		for _, v := range strings.Split(level, ",") {
			if _, ok := lv[v]; !ok {
				helper.Print(ctx, false, helper.ParamErr)
				return
			}
		}
	} else {
		if names == "" {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}

		usernames := strings.Split(names, ",")
		for _, v := range usernames {
			if !validator.CheckUName(v, 5, 14) {
				helper.Print(ctx, false, helper.UsernameErr)
				return
			}
		}
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	record := g.Record{
		"id":          helper.GenId(),
		"title":       title,             //标题
		"sub_title":   subTitle,          //副标题
		"content":     content,           //内容
		"is_top":      isTop,             //0不置顶 1置顶
		"is_push":     isPush,            //0不推送 1推送
		"is_vip":      isVip,             //是否是vip
		"level":       level,             //会员等级
		"ty":          ty,                //站内信类型
		"usernames":   names,             //会员名
		"state":       1,                 //1审核中 2审核通过 3审核拒绝 4已删除
		"send_state":  1,                 //1未发送 2已发送
		"send_name":   sendName,          //发送人名
		"apply_at":    ctx.Time().Unix(), //创建时间
		"apply_uid":   admin["id"],       //创建人uid
		"apply_name":  admin["name"],     //创建人名
		"review_at":   0,                 //审核时间
		"review_uid":  0,                 //审核人uid
		"review_name": "",                //审核人名
	}
	err = model.MessageInsert(record, sendAt)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 站内信列表
func (that *MessageController) List(ctx *fasthttp.RequestCtx) {

	page := ctx.PostArgs().GetUintOrZero("page")
	pageSize := ctx.PostArgs().GetUintOrZero("page_size")
	flag := ctx.PostArgs().GetUintOrZero("flag")                        //1审核列表 2历史记录
	title := string(ctx.PostArgs().Peek("title"))                       //标题
	sendName := string(ctx.PostArgs().Peek("send_name"))                //发送人名
	isVip := string(ctx.PostArgs().Peek("is_vip"))                      //是否vip站内信 1 vip站内信
	isPush := string(ctx.PostArgs().Peek("is_push"))                    //0不推送 1推送
	ty := ctx.PostArgs().GetUintOrZero("ty")                            //1站内消息 2活动消息
	sendStartTime := string(ctx.PostArgs().Peek("send_start_time"))     //发送开始时间
	sendEndTime := string(ctx.PostArgs().Peek("send_end_time"))         //发送结束时间
	startTime := string(ctx.PostArgs().Peek("start_time"))              //申请开始时间
	endTime := string(ctx.PostArgs().Peek("end_time"))                  //申请结束时间
	reviewStartTime := string(ctx.PostArgs().Peek("review_start_time")) //审核开始时间
	reviewEndTime := string(ctx.PostArgs().Peek("review_end_time"))     //审核结束时间

	ex := g.Ex{}
	if page == 0 {
		page = 1
	}
	if pageSize < 10 {
		pageSize = 10
	}

	if flag != 1 && flag != 2 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if flag == 1 {
		ex["state"] = 1
	} else {
		ex["state"] = []int{2, 3, 4}
	}

	if title != "" {
		ex["title"] = title
	}

	if sendName != "" {
		ex["send_name"] = sendName
	}

	if ty > 0 {
		if ty != 1 && ty != 2 {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}

		ex["ty"] = ty
	}

	if isVip != "" {
		if isVip != "0" && isVip != "1" {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}

		ex["is_vip"] = isVip
	}

	if isPush != "" {
		if isPush != "0" && isPush != "1" {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}

		ex["is_push"] = isPush
	}

	data, err := model.MessageList(page, pageSize, sendStartTime, sendEndTime, startTime, endTime, reviewStartTime, reviewEndTime, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 站内信编辑
func (that *MessageController) Update(ctx *fasthttp.RequestCtx) {

	id := string(ctx.PostArgs().Peek("id"))
	title := string(ctx.PostArgs().Peek("title"))        //标题
	subTitle := string(ctx.PostArgs().Peek("sub_title")) //副标题
	content := string(ctx.PostArgs().Peek("content"))    //内容
	isTop := string(ctx.PostArgs().Peek("is_top"))       //0不置顶 1置顶
	sendName := string(ctx.PostArgs().Peek("send_name")) //发送人名
	sendAt := string(ctx.PostArgs().Peek("send_at"))     //发送时间

	record := g.Record{}
	if !validator.CtypeDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	if title != "" {
		record["title"] = title
	}

	if subTitle != "" {
		record["sub_title"] = subTitle
	}

	if content != "" {
		record["content"] = content
	}

	if isTop != "" {
		if isTop != "0" && isTop != "1" {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}

		record["is_top"] = isTop
	}

	if sendName != "" {
		record["send_name"] = sendName
	}
	err := model.MessageUpdate(id, sendAt, record)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 站内信编辑
func (that *MessageController) Review(ctx *fasthttp.RequestCtx) {

	id := string(ctx.PostArgs().Peek("id"))
	state := ctx.PostArgs().GetUintOrZero("state")
	if !validator.CtypeDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	states := map[int]bool{
		2: true,
		3: true,
	}
	if _, ok := states[state]; !ok {
		helper.Print(ctx, false, helper.StateParamErr)
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	err = model.MessageReview(id, state, admin)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 已发站内信详情
func (that *MessageController) Detail(ctx *fasthttp.RequestCtx) {

	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")
	id := string(ctx.QueryArgs().Peek("id"))

	if page == 0 {
		page = 1
	}
	if pageSize < 10 {
		pageSize = 10
	}
	s, err := model.MessageDetail(id, page, pageSize)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.PrintJson(ctx, true, s)
}

// 站内信删除
func (that *MessageController) Delete(ctx *fasthttp.RequestCtx) {

	id := string(ctx.PostArgs().Peek("id"))
	msgID := string(ctx.PostArgs().Peek("msg_id"))
	if msgID != "" {
		for _, v := range strings.Split(msgID, ",") {
			if !validator.CtypeDigit(v) {
				helper.Print(ctx, false, helper.ParamErr)
				return
			}
		}
	}

	err := model.MessageDelete(id, msgID)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
