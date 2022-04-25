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
	isTop := ctx.PostArgs().GetUintOrZero("is_top")      //0不置顶 1置顶
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

	if isVip == 1 {
		if level == "" {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}

		lv := map[string]bool{
			"0":  true,
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
		}
		for _, v := range strings.Split(level, ",") {
			if _, ok := lv[v]; !ok {
				helper.Print(ctx, false, helper.ParamErr)
				return
			}
		}
	} else {
		if names != "" {
			usernames := strings.Split(names, ",")
			for _, v := range usernames {
				if !validator.CheckUName(v, 5, 14) {
					helper.Print(ctx, false, helper.UsernameErr)
					return
				}
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
		"is_vip":      isVip,             //是否是vip
		"level":       level,             //会员等级
		"usernames":   names,             //会员名
		"state":       1,                 //1审核中 2审核通过 3审核拒绝 4已删除
		"send_state":  1,                 //1未发送 2已发送
		"send_name":   sendName,          //发送人名
		"apply_at":    ctx.Time().Unix(), //创建时间
		"apply_uid":   admin["id"],       //创建人uid
		"apply_name":  admin["name"],     //创建人名
		"review_at":   0,                 //创建时间
		"review_uid":  0,                 //创建人uid
		"review_name": "",                //创建人名
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

	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")
	title := string(ctx.PostArgs().Peek("title"))                        //标题
	sendName := string(ctx.PostArgs().Peek("send_name"))                 //发送人名
	isVip := string(ctx.PostArgs().Peek("is_vip"))                       //是否vip站内信 1 vip站内信
	sendStartTime := string(ctx.PostArgs().Peek("send_start_time"))      //发送开始时间
	sendEndTime := string(ctx.PostArgs().Peek("send_end_time"))          //发送结束时间
	startTime := string(ctx.QueryArgs().Peek("start_time"))              //申请开始时间
	endTime := string(ctx.QueryArgs().Peek("end_time"))                  //申请结束时间
	reviewStartTime := string(ctx.QueryArgs().Peek("review_start_time")) //审核开始时间
	reviewEndTime := string(ctx.QueryArgs().Peek("review_end_time"))     //审核结束时间

	ex := g.Ex{}
	if page == 0 {
		page = 1
	}
	if pageSize < 10 {
		page = 10
	}

	if title != "" {
		ex["title"] = title
	}

	if sendName != "" {
		ex["send_name"] = sendName
	}

	if isVip != "" {
		if isVip != "0" && isVip != "1" {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}

		ex["is_vip"] = isVip
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

	err := model.MessageUpdate()
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 站内信编辑
func (that *MessageController) Review(ctx *fasthttp.RequestCtx) {

	err := model.MessageReview()
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 站内信删除
func (that *MessageController) Delete(ctx *fasthttp.RequestCtx) {

	err := model.MessageDelete()
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
