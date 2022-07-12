package controller

import (
	"merchant/contrib/helper"
	"merchant/contrib/validator"
	"merchant/model"
	"strconv"
	"strings"

	g "github.com/doug-martin/goqu/v9"
	"github.com/olivere/elastic/v7"
	"github.com/valyala/fasthttp"
)

type BlacklistController struct{}

func (that *BlacklistController) LogList(ctx *fasthttp.RequestCtx) {

	page := string(ctx.QueryArgs().Peek("page"))
	pageSize := string(ctx.QueryArgs().Peek("page_size"))
	if !validator.CheckStringDigit(page) || !validator.CheckStringDigit(pageSize) {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	ex := g.Ex{}
	//param := map[string]interface{}{}
	username := string(ctx.QueryArgs().Peek("username"))
	if len(username) > 0 {
		username = strings.ToLower(username)
		if !validator.CheckUName(username, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}

		ex["username"] = username
	}

	parentName := string(ctx.QueryArgs().Peek("parent_name"))
	if len(parentName) > 4 {
		parentName = strings.ToLower(parentName)
		if !validator.CheckUName(parentName, 5, 14) {
			helper.Print(ctx, false, helper.AgentNameErr)
			return
		}

		ex["parent_name"] = parentName
	}
	if parentName == "root" {
		ex["parent_name"] = "root"
	}

	deviceNo := string(ctx.QueryArgs().Peek("device_no"))
	if len(deviceNo) > 0 {
		ex["device_no"] = deviceNo
	}

	ip := string(ctx.QueryArgs().Peek("ip"))
	if len(ip) > 0 {
		ex["ip"] = ip
	}

	device := string(ctx.QueryArgs().Peek("device"))
	if len(device) > 0 {
		i, err := strconv.Atoi(device)
		if err != nil {
			helper.Print(ctx, false, helper.DeviceTypeErr)
			return
		}

		if _, ok := model.DeviceMap[i]; !ok {
			helper.Print(ctx, false, helper.DeviceTypeErr)
			return
		}

		//param["device"] = device
		ex["device"] = device
	}

	startTime := string(ctx.QueryArgs().Peek("start_time"))
	endTime := string(ctx.QueryArgs().Peek("end_time"))
	p, _ := strconv.Atoi(page)
	ps, _ := strconv.Atoi(pageSize)

	//data, err := model.MemberLoginLogList(startTime, endTime, p, ps, param)
	data, err := model.MemberLoginLogList(startTime, endTime, p, ps, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *BlacklistController) AssociateList(ctx *fasthttp.RequestCtx) {

	page := string(ctx.QueryArgs().Peek("page"))
	pageSize := string(ctx.QueryArgs().Peek("page_size"))
	if !validator.CheckStringDigit(page) || !validator.CheckStringDigit(pageSize) {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	tys := string(ctx.QueryArgs().Peek("ty"))
	if !validator.CheckIntScope(tys, model.TyDevice, model.TyVirtualAccount) {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	value := string(ctx.QueryArgs().Peek("value"))
	if !validator.CheckStringLength(value, 1, 60) {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	ty, _ := strconv.Atoi(tys)
	query := elastic.NewBoolQuery()

	aggField := "device_no.keyword"
	if ty == model.TyDevice {
		query.Filter(elastic.NewTermQuery("device_no.keyword", value))
		aggField = "ips.keyword"
	} else if ty == model.TyIP {
		query.Filter(elastic.NewTermQuery("ips.keyword", value))
	}

	p, _ := strconv.Atoi(page)
	ps, _ := strconv.Atoi(pageSize)

	data, err := model.MemberAssocLoginLogList(p, ps, aggField, query)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.PrintJson(ctx, true, data)
}

func (that *BlacklistController) List(ctx *fasthttp.RequestCtx) {

	startTime := string(ctx.QueryArgs().Peek("start_time"))
	endTime := string(ctx.QueryArgs().Peek("end_time"))
	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")
	ty := ctx.QueryArgs().GetUintOrZero("ty")

	if _, ok := model.BlackTy[ty]; !ok {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	ex := g.Ex{
		"ty": ty,
	}
	value := string(ctx.QueryArgs().Peek("value"))
	if len(value) > 0 {
		ex["value"] = value
	}
	data, err := model.BlacklistList(uint(page), uint(pageSize), startTime, endTime, ty, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *BlacklistController) Insert(ctx *fasthttp.RequestCtx) {

	ty := ctx.PostArgs().GetUintOrZero("ty")
	if _, ok := model.BlackTy[ty]; !ok {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	value := string(ctx.PostArgs().Peek("value"))
	switch ty {
	case model.TyBankcard:
		if !validator.CheckStringLength(value, 6, 20) || !validator.CheckStringDigit(value) {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}
	case model.TyRebate, model.TyCGRebate, model.TyPromoteLink:
		value = strings.ToLower(value)
		if !validator.CheckUName(value, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}

		if !model.MemberExist(value) {
			helper.Print(ctx, false, helper.UserNotExist)
			return
		}

	default:
		if !validator.CheckStringLength(value, 1, 60) {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}
	}

	remark := string(ctx.PostArgs().Peek("remark"))
	if !validator.CheckStringLength(remark, 1, 1000) {
		helper.Print(ctx, false, helper.RemarkFMTErr)
		return
	}

	record := g.Record{
		"id":     helper.GenId(),
		"ty":     ty,
		"value":  value,
		"remark": remark,
	}
	err := model.BlacklistInsert(ctx, ty, value, record)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 只能更新remark
func (that *BlacklistController) Update(ctx *fasthttp.RequestCtx) {

	id := string(ctx.PostArgs().Peek("id"))
	if !validator.CheckStringDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	remark := string(ctx.PostArgs().Peek("remark"))
	if !validator.CheckStringLength(remark, 1, 1000) {
		helper.Print(ctx, false, helper.RemarkFMTErr)
		return
	}

	remark = validator.FilterInjection(remark)
	err := model.BlacklistUpdate(id, remark)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *BlacklistController) Delete(ctx *fasthttp.RequestCtx) {

	id := string(ctx.QueryArgs().Peek("id"))
	if !validator.CheckStringDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	/// 从数据库 和 redis删除黑名单
	err := model.BlacklistDelete(id)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *BlacklistController) ClearPhone(ctx *fasthttp.RequestCtx) {

	phone := string(ctx.PostArgs().Peek("phone"))
	if !validator.IsVietnamesePhone(phone) {
		helper.Print(ctx, false, helper.PhoneFMTErr)
		return
	}

	err := model.BlacklistClearPhone(phone)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, "success")
}
