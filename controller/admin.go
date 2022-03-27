package controller

import (
	"encoding/base32"
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"merchant2/model"

	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"

	"strconv"
	"strings"
)

type AdminController struct{}

func (that *AdminController) Insert(ctx *fasthttp.RequestCtx) {

	data := model.Admin{}
	err := validator.Bind(ctx, &data)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	seamo := strings.TrimSpace(string(ctx.PostArgs().Peek("seamo")))
	if _, err := base32.StdEncoding.DecodeString(seamo); err != nil {
		helper.Print(ctx, false, helper.SeamoErr)
		return
	}

	// 判断用户名是否已经存在
	if model.AdminExist(g.Ex{"name": data.Name}) {
		helper.Print(ctx, false, helper.UsernameExist)
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	now := uint32(ctx.Time().Unix())
	data.CreateAt = now
	data.CreatedUid = admin["id"]
	data.CreatedName = admin["name"]
	data.UpdatedAt = now
	data.UpdatedUid = admin["id"]
	data.UpdatedName = admin["name"]
	err = model.AdminInsert(data)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, "succeed")
}

func (that *AdminController) Update(ctx *fasthttp.RequestCtx) {

	id := string(ctx.PostArgs().Peek("id"))
	gid := string(ctx.PostArgs().Peek("group_id"))
	seamo := string(ctx.PostArgs().Peek("seamo"))
	pwd := string(ctx.PostArgs().Peek("password"))
	state := string(ctx.PostArgs().Peek("state"))

	s := map[string]bool{
		"0": true,
		"1": true,
	}
	if _, ok := s[state]; !ok {
		helper.Print(ctx, false, helper.StateParamErr)
		return
	}

	if !validator.CtypeDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	if !validator.CtypeDigit(gid) {
		helper.Print(ctx, false, helper.GroupIDErr)
		return
	}

	if len(pwd) > 0 {
		if !validator.CheckAPassword(pwd, 5, 20) {
			helper.Print(ctx, false, helper.PasswordFMTErr)
			return
		}
	}

	record := g.Record{}
	if seamo != "" {
		_, err := base32.StdEncoding.DecodeString(seamo)
		if err != nil {
			helper.Print(ctx, false, helper.SeamoErr)
			return
		}

		record["seamo"] = seamo
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	record["state"] = state
	record["group_id"] = gid
	record["updated_at"] = ctx.Time().Unix()
	record["updated_uid"] = admin["id"]
	record["updated_name"] = admin["name"]
	err = model.AdminUpdate(id, pwd, record)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, "succeed")
}

func (that *AdminController) UpdateState(ctx *fasthttp.RequestCtx) {

	id := string(ctx.QueryArgs().Peek("id"))
	state := string(ctx.QueryArgs().Peek("state"))

	s := map[string]bool{
		"0": true,
		"1": true,
	}
	if _, ok := s[state]; !ok {
		helper.Print(ctx, false, helper.StateParamErr)
		return
	}

	if !validator.CtypeDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	err := model.AdminUpdateState(ctx, id, state)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, "succeed")
}

func (that *AdminController) List(ctx *fasthttp.RequestCtx) {

	var (
		exs  []g.Expression
		size uint = 10
	)

	ex := g.Ex{}
	page := string(ctx.QueryArgs().Peek("page"))
	name := string(ctx.QueryArgs().Peek("name"))
	state := string(ctx.QueryArgs().Peek("state"))
	groupid := string(ctx.QueryArgs().Peek("groupid"))

	if len(name) > 0 {
		if !validator.CheckAName(name, 5, 20) {
			helper.Print(ctx, false, helper.AdminNameErr)
			return
		}
	}

	cpage, err := strconv.ParseUint(page, 10, 64)
	if err != nil {
		cpage = 1
	}
	if cpage < 1 {
		cpage = 1
	}
	if state == "0" || state == "1" {
		ex["state"] = state
	}

	if validator.CtypeAlnum(name) {
		ex["name"] = name
	}
	if validator.CtypeDigit(groupid) {
		ex["group_id"] = groupid
	}

	if len(ex) > 0 {
		exs = append(exs, ex)
	}

	data, err := model.AdminList(uint(cpage), size, exs...)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

/**
 * @Description: Login 用户登录
 */

func (that *AdminController) Login(ctx *fasthttp.RequestCtx) {

	deviceNo := string(ctx.Request.Header.Peek("no"))
	username := string(ctx.PostArgs().Peek("username"))
	password := string(ctx.PostArgs().Peek("pwd"))
	seamo := string(ctx.PostArgs().Peek("seamo"))

	if !validator.CheckAName(username, 5, 20) {
		helper.Print(ctx, false, helper.AdminNameErr)
		return
	}

	if !validator.CheckAPassword(password, 5, 20) {
		helper.Print(ctx, false, helper.UsernameOrPasswordErr)
		return
	}

	ip := helper.FromRequest(ctx)
	resp, err := model.AdminLogin(deviceNo, username, password, seamo, ip, uint32(ctx.Time().Unix()))
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, resp)
}
