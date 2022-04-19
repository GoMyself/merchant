package controller

import (
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"merchant2/message"
	"merchant2/model"
	"strconv"
	"strings"
)

type MessageController struct{}

type templateInsertParam struct {
	Module  uint8  `rule:"digit" name:"module" min:"1" max:"4" msg:"module error[1-4]"` // 应用模块 1系统公告、2站内消息、3代理公告、4代理消息
	Common  uint8  `name:"common" rule:"digit" min:"1" max:"2" msg:"common error"`      //是否常用： 1是 2否
	Ty      uint8  `name:"ty" rule:"digit" min:"0" max:"1" msg:"ty error[0-1]"`         // 0全部, 1其他
	Title   string `name:"title" rule:"filter" min:"1" max:"50" msg:"title error"`
	Content string `name:"content" rule:"filter" min:"1" max:"1000" msg:"content error[0-1000]"`
	SortId  int    `name:"sort_id" rule:"digit" min:"0" msg:"sort_id error"`
}

type templateUpdateParam struct {
	ID      string `name:"id" rule:"none"`
	Title   string `name:"title" rule:"filter" min:"1" max:"50" msg:"title error"`
	Content string `name:"content" rule:"filter" min:"1" max:"1000" msg:"content error[0-1000]"`
}

type templateListParam struct {
	Title       string `name:"title" rule:"none"`
	Ty          int8   `name:"ty" rule:"digit" default:"-1" min:"-1" max:"1" msg:"ty error[0-1]"`
	Module      uint8  `name:"module" rule:"digit" default:"0" min:"0" max:"4" msg:"module error[1-4]"`
	Common      uint8  `name:"common" rule:"digit" default:"0" min:"0" max:"2" msg:"common error"`
	CreatedName string `name:"created_name" rule:"none"`
	StartTime   string `name:"start_time" rule:"none" default:""` // apply_at 申请时间 开始时间
	EndTime     string `name:"end_time" rule:"none" default:""`
	Page        uint   `name:"page" rule:"digit" default:"1" msg:"page error"`
	PageSize    uint   `name:"page_size" rule:"digit" default:"15" msg:"page_size error"`
}

type postsInsertParam struct {
	Title         string `name:"title" rule:"filter" min:"1" max:"255" msg:"title error[2-50]"`
	Content       string `name:"content" rule:"filter" min:"1" max:"1000" msg:"content error[1-1000]"`
	Redirect      int64  `name:"redirect" rule:"digit" min:"1" max:"2" msg:"redirect error"`                   // 是否跳转 1 是 2否
	RedirectUrl   string `name:"redirect_url" rule:"none"`                                                     // 跳转链接
	Icon          string `name:"icon" rule:"none" msg:"icon error"`                                            // icon url
	Ty            uint8  `name:"ty" rule:"digit" min:"1" max:"3" msg:"ty error"`                               // 消息类型：1普通 2 特殊 3财务
	TopStartTime  string `name:"top_start_time" rule:"none"`                                                   // 置顶时间
	TopEndTime    string `name:"top_end_time" rule:"none"`                                                     // 置顶时间
	ShowStartTime string `name:"show_start_time" rule:"none"`                                                  // 启用开始时间
	ShowEndTime   string `name:"show_end_time" rule:"none"`                                                    // 启用结束时间
	DeviceTy      uint8  `name:"device_ty" rule:"digit" default:"1" min:"1" max:"2" msg:"device_ty error"`     // 设备类型： 1 设备 2 指定域名
	Device        string `name:"device" rule:"sDigit" msg:"device error"`                                      // 设备: 1-全站APP,2-体育APP,3-web,4-h5,5-棋牌app 多端接收用英文逗号分隔  指定域名: url
	Push          uint8  `name:"push" rule:"digit" default:"2" min:"1" max:"2" msg:"push error"`               // 是否推送：1推送 2不推送
	PushDevice    uint8  `name:"push_device" rule:"digit" default:"0" min:"0" max:"2" msg:"push_device error"` // 0全部、1Android、2IOS
}

type postsUpdateParam struct {
	ID      string `rule:"none" name:"id"`
	Title   string `name:"title" rule:"filter" min:"1" max:"255" msg:"title error[2-50]"`
	Content string `name:"content" rule:"filter" min:"1" max:"1000" msg:"content error[1-1000]"`
}

type postsListParam struct {
	Title       string `name:"title" rule:"none" default:""`
	StartTime   string `name:"start_time" rule:"none" default:""`
	EndTime     string `name:"end_time" rule:"none" default:""`
	State       uint8  `name:"state" rule:"none" default:"0"`
	CreatedName string `name:"created_name" rule:"none" default:""`
	Ty          uint8  `name:"ty" rule:"digit" default:"0" min:"0" max:"3" msg:"ty error"`
	Page        uint   `name:"page" rule:"digit" default:"1" msg:"page error"`
	PageSize    uint   `name:"page_size" rule:"digit" default:"15" msg:"page_size error"`
}

type letterListParam struct {
	Title       string `name:"title" rule:"none" default:""`
	StartTime   string `name:"start_time" rule:"none" default:""`
	EndTime     string `name:"end_time" rule:"none" default:""`
	State       uint8  `name:"state" rule:"none" default:"0"`
	CreatedName string `name:"created_name" rule:"none" default:""`
	Page        uint   `name:"page" rule:"digit" default:"1" msg:"page error"`
	PageSize    uint   `name:"page_size" rule:"digit" default:"15" msg:"page_size error"`
}

type letterInsertParam struct {
	Title      string `name:"title" rule:"filter" min:"1" max:"255" msg:"title error[2-50]"`
	Content    string `name:"content" rule:"filter" min:"1" max:"1000" msg:"content error[1-1000]"`
	Icon       string `name:"icon" rule:"none" msg:"icon error"`                                            // icon url
	Ty         uint8  `name:"ty" rule:"digit" min:"1" max:"3" msg:"ty error"`                               // 消息类型：1普通 2 特殊 3财务
	Device     string `name:"device" rule:"sDigit" msg:"device error"`                                      // 设备: 1-全站APP,2-体育APP,3-web,4-h5,5-棋牌app 多端接收用英文逗号分隔  指定域名: url
	Push       uint8  `name:"push" rule:"digit" default:"2" min:"1" max:"2" msg:"push error"`               // 是否推送：1推送 2不推送
	PushDevice uint8  `name:"push_device" rule:"digit" default:"0" min:"0" max:"2" msg:"push_device error"` // 0全部、1Android、2IOS
	IsAll      uint8  `name:"is_all" rule:"digit" min:"1" max:"2" msg:"is_all error"`
	Usernames  string `name:"usernames" rule:"none"`
	Level      string `name:"level" rule:"none" default:"0"`
}

type letterUpdateParam struct {
	ID      string `rule:"none" name:"id"`
	Title   string `name:"title" rule:"filter" min:"1" max:"255" msg:"title error[2-50]"`
	Content string `name:"content" rule:"filter" min:"1" max:"1000" msg:"content error[1-1000]"`
}

type tplListParam struct {
	Ty       uint8  `name:"ty" rule:"digit" default:"1" min:"1" max:"2" msg:"ty error[1-2]"`
	Keyword  string `name:"keyword" rule:"none" `
	Title    string `name:"title" rule:"none"`
	Scene    string `name:"scene" rule:"none"`
	Module   int    `name:"module" rule:"digit" default:"0" min:"0" max:"354" msg:"module error"`
	State    int8   `name:"state" rule:"digit" default:"0" min:"0" max:"2" msg:"state error[0-2]"`
	Page     int    `name:"page" rule:"digit" default:"1" min:"1" msg:"page error"`
	PageSize int    `name:"page_size" rule:"digit" default:"10" min:"10" max:"200" msg:"page_size error"`
}

type tplNewParam struct {
	Scene   int    `name:"scene" rule:"digit" min:"311" max:"337" msg:"scene error[311-337]"`
	Module  int    `name:"module" rule:"digit" min:"351" max:"354" msg:"module error[351-354]"`
	Title   string `name:"title" rule:"filter" min:"1" max:"255" msg:"title error[1-255]"`
	Content string `name:"content" rule:"filter" min:"1" max:"1000" msg:"content error[1-1000]"`
	Icon    string `name:"icon" rule:"url" msg:"icon error"`
}

type tplEditParam struct {
	ID      string `name:"id" rule:"digit" msg:"id error"`
	Title   string `name:"title" rule:"filter" min:"1" max:"255" msg:"title error[1-255]"`
	Content string `name:"content" rule:"filter" min:"1" max:"1000" msg:"content error[1-1000]"`
	Icon    string `name:"icon" rule:"url" msg:"icon error"`
}

type tplStateParam struct {
	ID    string `name:"id" rule:"digit" msg:"id error"`
	State uint8  `name:"state" rule:"digit" min:"1" max:"2" msg:"state error[1-2]"`
}

func (that *MessageController) PostsInsert(ctx *fasthttp.RequestCtx) {

	param := postsInsertParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	posts := message.Posts{
		ID:          helper.GenId(),
		Title:       param.Title,
		RedirectUrl: param.RedirectUrl,
		Redirect:    param.Redirect,
		Content:     param.Content,
		Icon:        param.Icon,
		Ty:          param.Ty,
		State:       message.PostsStateReviewing,
		DeviceTy:    param.DeviceTy,
		Device:      param.Device,
		Push:        param.Push, // 是否推送：1推送 2不推送
		CreatedAt:   ctx.Time().Unix(),
		CreatedUid:  admin["id"],
		CreatedName: admin["name"],
		IsShow:      message.No,
		Top:         message.No,
	}

	if posts.DeviceTy == message.PostsDeviceTyURL {

		urls := strings.Split(param.Device, ",")
		for _, v := range urls {
			if !validator.CheckUrl(v) {
				helper.Print(ctx, false, helper.ImagesURLErr)
				return
			}
		}

		posts.Device = param.Device
	}

	if posts.Redirect == message.Yes {

		if !validator.CheckUrl(param.RedirectUrl) {
			helper.Print(ctx, false, helper.RedirectURLErr)
			return
		}

		posts.RedirectUrl = param.RedirectUrl
	}

	if posts.Push == message.Yes {

		if param.PushDevice < 0 || param.PushDevice > 2 {
			helper.Print(ctx, false, helper.PushDeviceErr)
			return
		}

		posts.PushDevice = param.PushDevice
	}

	if param.TopStartTime != "" && param.TopEndTime != "" {

		posts.Top = message.Yes
	}

	if param.ShowStartTime != "" && param.ShowEndTime != "" {

		posts.IsShow = message.Yes
	}

	err = model.MessagePostsInsert(posts, param.TopStartTime, param.TopEndTime, param.ShowStartTime, param.ShowEndTime)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 系统公告列表
func (that *MessageController) PostsList(ctx *fasthttp.RequestCtx) {

	param := postsListParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	ex := g.Ex{}
	if param.Title != "" {
		ex["title"] = param.Title
	}

	if param.State != 0 {

		if param.State < 1 && param.State > 4 {
			helper.Print(ctx, false, helper.StateParamErr)
			return
		}

		ex["state"] = param.State
	}

	if param.CreatedName != "" {
		ex["created_name"] = param.CreatedName
	}

	if param.Ty != 0 {
		ex["ty"] = param.Ty
	}
	data, err := model.MessagePostsList(param.Page, param.PageSize, param.StartTime, param.EndTime, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 系统公告编辑
func (that *MessageController) PostsUpdate(ctx *fasthttp.RequestCtx) {

	param := postsUpdateParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	err = model.MessagePostsUpdate(param.ID, param.Title, param.Content)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 系统停用 启用 系统审核
func (that *MessageController) PostsState(ctx *fasthttp.RequestCtx) {

	id := string(ctx.PostArgs().Peek("id"))
	if id == "" || !validator.CheckStringDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	state := string(ctx.PostArgs().Peek("state"))
	iState, err := strconv.Atoi(state)
	if err != nil {
		helper.Print(ctx, false, helper.StateParamErr)
		return
	}

	s := map[string]bool{
		"2": true,
		"3": true,
		"4": true,
	}
	if _, ok := s[state]; !ok {
		helper.Print(ctx, false, helper.StateParamErr)
		return
	}

	remark := string(ctx.PostArgs().Peek("remark"))
	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	err = model.MessagePostsState(id, admin["name"], admin["id"], remark, ctx.Time().Unix(), iState)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 系统公告删除
func (that *MessageController) PostsDel(ctx *fasthttp.RequestCtx) {

	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" || !validator.CheckStringDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	err := model.MessagePostsDelete([]string{id})
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

//系统审核详情
func (that *MessageController) PostsDetail(ctx *fasthttp.RequestCtx) {

	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" || !validator.CheckStringDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	data, err := model.MessagePostsReviewDetail(id)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 站内信列表
func (that *MessageController) LetterList(ctx *fasthttp.RequestCtx) {

	param := letterListParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	ex := g.Ex{}

	if param.Title != "" {
		ex["title"] = param.Title
	}

	if param.State != 0 {

		if param.State < 1 && param.State > 4 {
			helper.Print(ctx, false, helper.StateParamErr)
			return
		}

		ex["state"] = param.State
	}

	if param.CreatedName != "" {
		ex["created_name"] = param.CreatedName
	}

	data, err := model.MessageLetterList(param.Page, param.PageSize, param.StartTime, param.EndTime, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 站内消息添加
func (that *MessageController) LetterInsert(ctx *fasthttp.RequestCtx) {

	param := letterInsertParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	if param.IsAll == message.Yes && param.Level == "0" {

		helper.Print(ctx, false, helper.MemberLevelErr)
		return
	}

	if param.IsAll == message.No && param.Level != "0" {

		levels := strings.Split(param.Level, ",")
		for _, v := range levels {

			if !validator.CheckIntScope(v, 1, 11) {

				helper.Print(ctx, false, helper.MemberLevelErr)
				return
			}
		}
	}

	letter := message.Letter{
		ID:          helper.GenId(),
		Title:       param.Title,
		Content:     param.Content,
		Icon:        param.Icon,
		IsAll:       param.IsAll,
		Ty:          param.Ty,
		State:       message.PostsStateReviewing,
		Device:      param.Device,
		Level:       param.Level,
		Push:        param.Push, // 是否推送：1推送 2不推送
		CreatedAt:   ctx.Time().Unix(),
		CreatedUid:  admin["id"],
		CreatedName: admin["name"],
	}

	if letter.Push == message.Yes {

		if param.PushDevice < 0 || param.PushDevice > 2 {
			helper.Print(ctx, false, helper.PushDeviceErr)
			return
		}

		letter.PushDevice = param.PushDevice
	}

	err = model.MessageLetterInsert(letter, param.Usernames)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 站内消息编辑
func (that *MessageController) LetterUpdate(ctx *fasthttp.RequestCtx) {

	param := letterUpdateParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	err = model.MessageLetterUpdate(param.ID, param.Title, param.Content)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 站内消息停用 启用 审核
func (that *MessageController) LetterState(ctx *fasthttp.RequestCtx) {

	id := string(ctx.PostArgs().Peek("id"))
	if id == "" || !validator.CheckStringDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	argState := string(ctx.PostArgs().Peek("state"))
	state, err := strconv.Atoi(argState)
	if err != nil {
		helper.Print(ctx, false, helper.StateParamErr)
		return
	}

	if state < 2 || state > 4 {
		helper.Print(ctx, false, helper.StateParamErr)
		return
	}

	remark := string(ctx.PostArgs().Peek("remark"))
	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	err = model.MessageLetterState(id, admin["name"], admin["id"], remark, ctx.Time().Unix(), state)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 站内消息删除
func (that *MessageController) LetterDel(ctx *fasthttp.RequestCtx) {

	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" || !validator.CheckStringDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	err := model.MessageLetterDelete([]string{id})
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

//站内消息审核详情
func (that *MessageController) LetterDetail(ctx *fasthttp.RequestCtx) {

	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" || !validator.CheckStringDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	data, err := model.MessageLetterReviewDetail(id)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 内容模板 列表
func (that *MessageController) TemplateList(ctx *fasthttp.RequestCtx) {

	param := templateListParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	ex := g.Ex{}
	if param.Title != "" {

		if !validator.CheckStringLength(param.Title, 1, 50) {
			helper.Print(ctx, false, helper.ContentLengthErr)
			return
		}

		ex["title"] = param.Title
	}

	if param.Ty != -1 {
		ex["ty"] = param.Ty
	}

	if param.Module != 0 {

		if param.Module < 1 || param.Module > 4 {
			helper.Print(ctx, false, helper.ModuleErr)
			return
		}

		ex["module"] = param.Module
	}

	if param.Common != 0 {

		if param.Common < 1 || param.Common > 2 {
			helper.Print(ctx, false, helper.CommonFlagErr)
			return
		}

		ex["common"] = param.Common
	}

	if param.CreatedName != "" {
		ex["created_name"] = param.CreatedName
	}

	data, err := model.MessageTemplateList(param.Page, param.PageSize, param.StartTime, param.EndTime, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)

}

// 内容模板 编辑
func (that *MessageController) TemplateUpdate(ctx *fasthttp.RequestCtx) {

	param := templateUpdateParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if param.ID == "" || !validator.CheckStringDigit(param.ID) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	ex := g.Ex{
		"id": param.ID,
	}
	record := g.Record{
		"title":   param.Title,
		"content": param.Content,
	}
	err = model.MessageTemplateUpdate(ex, record)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 内容模板 删除
func (that *MessageController) TemplateDel(ctx *fasthttp.RequestCtx) {

	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" || !validator.CheckStringDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	err := model.MessageTemplateDelete([]string{id})
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 内容模板 添加
func (that *MessageController) TemplateInsert(ctx *fasthttp.RequestCtx) {

	param := templateInsertParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	msg := message.Template{
		ID:          helper.GenId(),
		Module:      param.Module,
		Title:       param.Title,
		Common:      param.Common,
		Content:     param.Content,
		Ty:          param.Ty,
		SortId:      param.SortId,
		CreatedAt:   ctx.Time().Unix(),
		CreatedUid:  admin["id"],
		CreatedName: admin["name"],
	}
	err = model.MessageTemplateInsert(msg)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *MessageController) SystemTemplateList(ctx *fasthttp.RequestCtx) {

	param := tplListParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if param.Ty == message.TemplateTySysDetail {

		if !validator.CheckIntScope(param.Scene, message.SceneFinancialSettlement, message.SceneWithdrawRiskReject) {
			helper.Print(ctx, false, helper.ScenesErr)
			return
		}
	}

	if param.Module > 0 && param.Module < message.ModuleFinancial {
		helper.Print(ctx, false, helper.ModuleErr)
		return
	}

	data, err := model.MessageSystemTplList(param.Keyword, param.Title, param.Scene, param.State, param.Module, param.Page, param.PageSize)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *MessageController) SystemTemplateInsert(ctx *fasthttp.RequestCtx) {

	param := tplNewParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	msg := message.LetterTpl{
		ID:          helper.GenId(),
		Module:      param.Module,
		Ty:          2,
		Scene:       param.Scene,
		Title:       param.Title,
		Content:     param.Content,
		Icon:        param.Icon,
		State:       2,
		CreatedAt:   uint64(ctx.Time().Unix()),
		UpdatedAt:   uint64(ctx.Time().Unix()),
		UpdatedUid:  admin["id"],
		UpdatedName: admin["name"],
	}

	err = model.MessageSystemTplInsert(msg)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *MessageController) SystemTemplateUpdate(ctx *fasthttp.RequestCtx) {

	param := tplEditParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	msg := message.LetterTpl{
		Title:       param.Title,
		Content:     param.Content,
		Icon:        param.Icon,
		UpdatedAt:   uint64(ctx.Time().Unix()),
		UpdatedUid:  admin["id"],
		UpdatedName: admin["name"],
	}
	err = model.MessageSystemTplUpdate(param.ID, message.MessageSystemTplInfo, msg)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *MessageController) SystemTemplateDelete(ctx *fasthttp.RequestCtx) {

	idLs := string(ctx.QueryArgs().Peek("ids"))
	fmt.Println(idLs)
	if len(idLs) == 0 || !validator.CheckStringCommaDigit(idLs) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	var ids []string
	if strings.Contains(idLs, ",") {
		ids = strings.Split(idLs, ",")
	} else {
		ids = append(ids, idLs)
	}

	err := model.MessageSystemTplDelete(ids)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *MessageController) SystemTemplateState(ctx *fasthttp.RequestCtx) {

	param := tplStateParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	msg := message.LetterTpl{
		State:       param.State,
		ID:          param.ID,
		UpdatedAt:   uint64(ctx.Time().Unix()),
		UpdatedUid:  admin["id"],
		UpdatedName: admin["name"],
	}
	err = model.MessageSystemTplUpdate(param.ID, message.MessageSystemTplState, msg)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
