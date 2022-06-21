package controller

import (
	"fmt"
	"merchant/contrib/helper"
	"merchant/contrib/validator"
	"merchant/model"
	"net/url"
	"strconv"
	"strings"
	"time"

	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/valyala/fasthttp"
)

type BannerController struct{}

type bannerListParam struct {
	Flags     uint8  `rule:"digit" min:"1" max:"5" msg:"flags error,val [1-5]" name:"flags"`                   //1:APP闪屏页广告、2:轮播图广告、3:WEB体育场馆广告、4:站点广告位。
	Device    string `rule:"none" name:"device"`                                                               //设备号，多个设备用户逗号分隔
	StartTime string `rule:"none" msg:"start_time error" name:"start_time"`                                    //开始时间
	EndTime   string `rule:"none" msg:"end_time error" name:"end_time"`                                        //结束时间
	State     uint8  `rule:"digit" default:"0" min:"0" max:"3" msg:"state val[0-3]" required:"0" name:"state"` //状态 1待发布 2开启 3停用
	PageSize  uint   `rule:"digit" default:"10" min:"10" max:"200" msg:"page_size error" name:"page_size"`     //每页数量
	Page      uint   `rule:"digit" default:"1" min:"1" msg:"page error" name:"page"`                           //页码
}

type bannerUpdateParam struct {
	ID          string `json:"id" db:"id" rule:"digit" msg:"id error" name:"id"`                                                          //
	Title       string `json:"title" db:"title"  rule:"filter" msg:"title error" required:"0" name:"title"`                               //标题
	Device      string `json:"device" db:"device" rule:"sDigit" msg:"device error" required:"0" name:"device"`                            //设备类型(1,2)
	RedirectURL string `json:"redirect_url" db:"redirect_url" rule:"none" msg:"redirect_url error" required:"0" name:"redirect_url"`      //跳转地址
	Images      string `json:"images" db:"images" rule:"none" msg:"images error" required:"0" name:"images"`                              //图片路径
	Seq         string `json:"seq" db:"seq" rule:"digit" min:"1" max:"100" msg:"seq error" required:"0" name:"seq"`                       //排序
	Flags       string `json:"flags" db:"flags" rule:"digit" min:"1" max:"10" msg:"flags error" required:"0" name:"flags"`                //广告类型
	ShowType    string `json:"show_type" db:"show_type" rule:"digit" min:"1" max:"2" msg:"show_type error" required:"0" name:"show_type"` //1 永久有效 2 指定时间
	ShowAt      string `json:"show_at" db:"show_at" rule:"none" msg:"show_at error" required:"0" name:"show_at"`                          //开始展示时间
	HideAt      string `json:"hide_at" db:"hide_at" rule:"none" msg:"hide_at error" required:"0" name:"hide_at"`                          //结束展示时间
	URLType     string `json:"url_type" db:"url_type" rule:"digit" min:"0" max:"3" msg:"url_type error" required:"0" name:"url_type"`     //链接类型 1站内 2站外
}

func (that *BannerController) List(ctx *fasthttp.RequestCtx) {

	params := bannerListParam{}
	err := validator.Bind(ctx, &params)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	exs := exp.NewExpressionList(exp.AndType)
	ex := g.Ex{"flags": params.Flags}
	if params.Device != "" {
		//ex["device"] = params.Device
		exs = exs.Append(g.Or(g.Ex{"device": g.Op{"like": params.Device}}, g.Ex{"device": 0}))
	}

	if params.State > 0 {
		ex["state"] = params.State
	}

	exs = exs.Append(ex)

	data, err := model.BannerList(params.StartTime, params.EndTime, params.Page, params.PageSize, exs)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *BannerController) Insert(ctx *fasthttp.RequestCtx) {

	params := model.Banner{}
	err := validator.Bind(ctx, &params)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	params.Title, _ = url.QueryUnescape(params.Title)
	if params.ShowType == model.BannerShowTypeSpecify {
		_, err := time.Parse("2006-01-02 15:04:05", params.ShowAt)
		if err != nil {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}

		_, err = time.Parse("2006-01-02 15:04:05", params.HideAt)
		if err != nil {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}
	}

	data, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	params.UpdatedAt = fmt.Sprintf("%d", ctx.Time().Unix())
	params.ID = helper.GenId()
	params.UpdatedName = data["name"]
	params.UpdatedUID = data["id"]
	params.State = model.BannerStateWait

	switch params.URLType {
	case "1": //站内链接
		if !strings.HasPrefix(params.RedirectURL, "/") {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}
	case "2": //站外链接
		_, err := url.Parse(params.RedirectURL)
		if err != nil {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}
	}

	err = model.BannerInsert(params)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *BannerController) Update(ctx *fasthttp.RequestCtx) {

	params := bannerUpdateParam{}
	err := validator.Bind(ctx, &params)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	params.Title, _ = url.QueryUnescape(params.Title)
	record := g.Record{}
	if len(params.Title) > 0 {
		record["title"] = params.Title
	}

	if len(params.Device) > 0 {
		record["device"] = params.Device
	}

	if len(params.RedirectURL) > 0 {
		switch params.URLType {
		case "1": //站内链接
			if !strings.HasPrefix(params.RedirectURL, "/") {
				helper.Print(ctx, false, helper.ParamErr)
				return
			}
		case "2": //站外链接
			_, err := url.Parse(params.RedirectURL)
			if err != nil {
				helper.Print(ctx, false, helper.ParamErr)
				return
			}
		}

		record["redirect_url"] = params.RedirectURL
	}

	if len(params.Images) > 0 {
		record["images"] = params.Images
	}

	if len(params.Seq) > 0 {
		record["seq"] = params.Seq
	}

	if len(params.Flags) > 0 {
		record["flags"] = params.Flags
	}

	if len(params.ShowType) > 0 {
		record["show_type"] = params.ShowType
	}

	if len(params.URLType) > 0 {
		record["url_type"] = params.URLType
	}

	if len(record) == 0 {
		helper.Print(ctx, false, helper.NoDataUpdate)
		return
	}

	data, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	record["updated_name"] = data["name"]
	record["updated_uid"] = data["id"]
	record["updated_at"] = fmt.Sprintf("%d", ctx.Time().Unix())
	err = model.BannerUpdate(params.ShowAt, params.HideAt, params.ID, record)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *BannerController) Delete(ctx *fasthttp.RequestCtx) {

	id := string(ctx.QueryArgs().Peek("id"))
	if !validator.CheckStringDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	err := model.BannerDelete(id)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *BannerController) UpdateState(ctx *fasthttp.RequestCtx) {

	id := string(ctx.PostArgs().Peek("id"))
	if !validator.CheckStringDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	stateVal := string(ctx.PostArgs().Peek("state"))
	if !validator.CheckStringDigit(stateVal) &&
		!validator.CheckIntScope(stateVal, model.BannerStateWait, model.BannerStateEnd) {
		helper.Print(ctx, false, helper.StateParamErr)
		return
	}

	state, _ := strconv.Atoi(stateVal)
	err := model.BannerUpdateState(id, uint8(state))
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
