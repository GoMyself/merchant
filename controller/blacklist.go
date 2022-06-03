package controller

import (
	"fmt"
	"log"
	"merchant/contrib/helper"
	"merchant/contrib/validator"
	"merchant/model"
	"os"
	"strconv"

	g "github.com/doug-martin/goqu/v9"
	"github.com/olivere/elastic/v7"
	"github.com/valyala/fasthttp"
)

type BlacklistController struct{}

var (
	logger = log.New(os.Stderr, "WARNING -:", 18)
)

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

	parentName := string(ctx.QueryArgs().Peek("parent_name"))
	if len(parentName) > 4 {
		if !validator.CheckUName(parentName, 5, 14) {
			helper.Print(ctx, false, helper.AgentNameErr)
			return
		}

		fmt.Println(parentName)
		//param["parents"] = agency
		ex["parent_name"] = parentName
	}
	if parentName == "root" {
		ex["parent_name"] = "root"
	}

	deviceNo := string(ctx.QueryArgs().Peek("device_no"))
	if len(deviceNo) > 0 {
		//param["device_no.keyword"] = deviceNo
		ex["device_no"] = deviceNo
	}

	ip := string(ctx.QueryArgs().Peek("ip"))
	if len(ip) > 0 {
		//param["ips.keyword"] = ip
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
	page_size := ctx.QueryArgs().GetUintOrZero("page_size")
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
	data, err := model.BlacklistList(uint(page), uint(page_size), startTime, endTime, ty, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *BlacklistController) Insert(ctx *fasthttp.RequestCtx) {
	ty := ctx.PostArgs().GetUintOrZero("ty")
	if _, ok := model.BlackTy[ty]; !ok {
		logger.Printf("insert blacklist failed: %+v %v  ParamErr:%v \n", &ctx, false, helper.ParamErr)
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	value := string(ctx.PostArgs().Peek("value"))
	switch ty {
	case model.TyBankcard:
		if !validator.CheckStringLength(value, 6, 20) || !validator.CheckStringDigit(value) {
			logger.Printf("case TyBankcard blacklist: %+v %v  ParamErr:%v \n", value, false, helper.ParamErr)
			helper.Print(ctx, false, helper.ParamErr)
			return
		}
	case model.TyRebate, model.TyCGRebate:
		if !validator.CheckUName(value, 5, 14) {
			logger.Printf("case TyCGRebate blacklist: %+v %v  UsernameErr:%v \n", value, false, helper.UsernameErr)
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}
	default:
		if !validator.CheckStringLength(value, 1, 60) {
			logger.Printf("case default blacklist: %+v %v  ParamErr:%v \n", value, false, helper.ParamErr)
			helper.Print(ctx, false, helper.ParamErr)
			return
		}
	}

	remark := string(ctx.PostArgs().Peek("remark"))
	if !validator.CheckStringLength(remark, 1, 1000) {
		logger.Printf("remark check blacklist: %+v %v  ParamErr:%v \n", remark, false, helper.RemarkFMTErr)
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
	logger.Printf("insert result record: %+v %v  record:%v err:%v \n", ty, value, record, err)

	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 只能更新remark
func (that *BlacklistController) Update(ctx *fasthttp.RequestCtx) {

	id := string(ctx.PostArgs().Peek("id"))
	logger.Printf("case update check blacklist ctx: %+v remark:%v, id:%v  RemarkFMTErr:%v \n", id, helper.RemarkFMTErr)

	if !validator.CheckStringDigit(id) {
		logger.Printf("case update check id IDErr: %+v  id:%v  IDErr:%v \n", helper.IDErr, id, helper.IDErr)

		helper.Print(ctx, false, helper.IDErr)
		return
	}

	remark := string(ctx.PostArgs().Peek("remark"))
	if !validator.CheckStringLength(remark, 1, 1000) {
		logger.Printf("case update check blacklist ctx: %+v remark:%v, id:%v  RemarkFMTErr:%v \n", ctx, remark, id, helper.RemarkFMTErr)

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
	logger.Printf("case update check blacklist id: %+v %v  IDErr:%v \n", err.Error(), id, record)

	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *BlacklistController) Delete(ctx *fasthttp.RequestCtx) {

	id := string(ctx.QueryArgs().Peek("id"))

	if !validator.CheckStringDigit(id) {
		logger.Printf("case delete check blacklist id: %+v %v  IDErr:%v \n", ctx, id, helper.IDErr)

		helper.Print(ctx, false, helper.IDErr)
		return
	}

	err := model.BlacklistDelete(id)
	logger.Printf("case delete result err: %+v id:%v  IDErr:%v \n", err, id, helper.IDErr)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
