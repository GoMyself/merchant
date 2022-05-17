package controller

import (
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"merchant2/model"
	"strconv"

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
		if !validator.CheckUName(username, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}

		//param["username"] = username
		ex["username"] = username
	}

	agency := string(ctx.QueryArgs().Peek("agency"))
	if len(agency) > 0 {
		if !validator.CheckUName(agency, 5, 14) {
			helper.Print(ctx, false, helper.AgentNameErr)
			return
		}

		//param["parents"] = agency
		ex["parents"] = agency
	}

	deviceNo := string(ctx.QueryArgs().Peek("device_no"))
	if len(deviceNo) > 0 {
		//param["device_no.keyword"] = deviceNo
		ex["device_no.keyword"] = deviceNo
	}

	ip := string(ctx.QueryArgs().Peek("ip"))
	if len(ip) > 0 {
		//param["ips.keyword"] = ip
		ex["ips.keyword"] = ip
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

	page := string(ctx.QueryArgs().Peek("page"))
	pageSize := string(ctx.QueryArgs().Peek("page_size"))
	if !validator.CheckStringDigit(page) || !validator.CheckStringDigit(pageSize) {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

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
		switch ty {
		case model.TyBankcard:
			if !validator.CheckStringLength(value, 6, 20) || !validator.CheckStringDigit(value) {
				helper.Print(ctx, false, helper.ParamErr)
				return
			}

			//cardNoHash := fmt.Sprintf("%d", model.MurmurHash(value, 0))
			ex["value"] = value
			//fmt.Println(cardNoHash)
		default:
			if !validator.CheckStringLength(value, 1, 60) {
				helper.Print(ctx, false, helper.ParamErr)
				return
			}
		}
		if !validator.CheckStringLength(value, 1, 60) {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}
	}

	startTime := string(ctx.QueryArgs().Peek("start_time"))
	endTime := string(ctx.QueryArgs().Peek("end_time"))
	p, _ := strconv.Atoi(page)
	ps, _ := strconv.Atoi(pageSize)

	data, err := model.BlacklistList(uint(p), uint(ps), startTime, endTime, ty, ex)
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

	ex := g.Ex{
		"id": id,
	}
	record := g.Record{
		"remark": validator.FilterInjection(remark),
	}
	err := model.BlacklistUpdate(ex, record)
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

	err := model.BlacklistDelete(id)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
