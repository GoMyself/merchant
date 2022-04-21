package controller

import (
	"fmt"
	"github.com/doug-martin/goqu/v9/exp"
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"merchant2/model"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wI2L/jettison"

	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"
)

type MemberController struct{}

type memberStateParam struct {
	Username string `rule:"none" name:"username"`                                 // 用户username  批量用逗号隔开
	State    int8   `rule:"digit" name:"state" min:"1" max:"2" msg:"state error"` // 状态： 1 正常 2 禁用
	Remark   string `rule:"filter" name:"remark" max:"300" msg:"remark error"`    // 备注
}

// setTagParam 设置/批量设置用户标签，取消用户标签
type setTagParam struct {
	Batch int    `rule:"digit" min:"0" max:"1" default:"0" msg:"batch error" name:"batch"` // 1批量添加 0编辑单个用户标签
	uid   string `rule:"sDigit" msg:"uid error" name:"uid"`
	tags  string `rule:"sDigit" min:"1" msg:"tags error" name:"tags"`
}

// setSVipParam 解除密码限制/解除短信限制 parameters structure
type retryResetParam struct {
	Username string `rule:"uname" min:"4" max:"9" msg:"username error" name:"username"`
	Ty       uint8  `rule:"digit" min:"1" max:"3" msg:"ty error" name:"ty"` // 1解除密码限制 2解除短信限制 3解除场馆钱包限制
	Pid      string `rule:"none" msg:"pid error" name:"pid"`                // 场馆id(解除场馆钱包限制时需要)
}

// 用户备注参数
type remarkLogParams struct {
	Username string `rule:"none" name:"username" msg:"username error"`
	File     string `rule:"none" name:"file" msg:"file error" default:""`
	Msg      string `rule:"none" name:"msg" max:"300"`
}

// GetAccountInfo 会员列表-帐户信息
func (that *MemberController) AccountInfo(ctx *fasthttp.RequestCtx) {

	username := string(ctx.QueryArgs().Peek("username"))
	if !validator.CheckUName(username, 4, 9) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	data, err := model.MemberAccountInfo(username)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// Balance 查询钱包余额
func (that *MemberController) BalanceBatch(ctx *fasthttp.RequestCtx) {

	uids := string(ctx.PostArgs().Peek("uids"))
	if !validator.CheckStringCommaDigit(uids) {
		helper.Print(ctx, false, helper.UIDErr)
		return
	}

	s := strings.Split(uids, ",")
	balance, err := model.MemberBalanceBatch(s)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.PrintJson(ctx, true, balance)
}

func (that *MemberController) TagBatch(ctx *fasthttp.RequestCtx) {

	uids := string(ctx.PostArgs().Peek("uids"))
	if !validator.CheckStringCommaDigit(uids) {
		helper.Print(ctx, false, helper.UIDErr)
		return
	}

	s := strings.Split(uids, ",")
	balance, err := model.MemberBatchTag(s)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.PrintJson(ctx, true, balance)
}

func (that *MemberController) Insert(ctx *fasthttp.RequestCtx) {

	name := string(ctx.PostArgs().Peek("username"))
	password := string(ctx.PostArgs().Peek("password"))
	maintainName := string(ctx.PostArgs().Peek("maintain_name"))
	groupName := string(ctx.PostArgs().Peek("group_name"))
	remark := string(ctx.PostArgs().Peek("remark"))
	sty := string(ctx.PostArgs().Peek("ty"))
	szr := string(ctx.PostArgs().Peek("zr"))
	sqp := string(ctx.PostArgs().Peek("qp"))
	sdj := string(ctx.PostArgs().Peek("dj"))
	sdz := string(ctx.PostArgs().Peek("dz"))
	scp := string(ctx.PostArgs().Peek("cp"))
	planID := string(ctx.PostArgs().Peek("plan_id"))
	agencyType := string(ctx.PostArgs().Peek("agency_type")) //391团队393普通
	if len(maintainName) == 0 {
		maintainName = ""
	}

	vs := model.RebateScale()
	ty, err := decimal.NewFromString(sty)
	if err != nil || ty.IsNegative() || ty.GreaterThan(vs.TY) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
	}

	zr, err := decimal.NewFromString(szr)
	if err != nil || zr.IsNegative() || zr.GreaterThan(vs.ZR) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
	}

	qp, err := decimal.NewFromString(sqp)
	if err != nil || qp.IsNegative() || qp.GreaterThan(vs.QP) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
	}

	dj, err := decimal.NewFromString(sdj)
	if err != nil || dj.IsNegative() || dj.GreaterThan(vs.DJ) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
	}

	dz, err := decimal.NewFromString(sdz)
	if err != nil || dz.IsNegative() || dz.GreaterThan(vs.DZ) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
	}

	cp, err := decimal.NewFromString(scp)
	if err != nil || cp.IsNegative() || cp.GreaterThan(vs.CP) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
	}

	if !validator.CheckUName(name, 4, 9) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	if !validator.CtypeAlnum(maintainName) {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if !validator.CheckUPassword(password, 8, 15) {
		helper.Print(ctx, false, helper.PasswordFMTErr)
		return
	}

	if agencyType != "391" && agencyType != "393" {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if agencyType == "391" && len(groupName) < 1 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	mr := model.MemberRebate{
		TY: ty.StringFixed(1),
		ZR: zr.StringFixed(1),
		QP: qp.StringFixed(1),
		DJ: dj.StringFixed(1),
		DZ: dz.StringFixed(1),
		CP: cp.StringFixed(1),
	}
	createdAt := uint32(ctx.Time().Unix())

	if !validator.CheckStringDigit(planID) {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	// 检测佣金方案是否存在
	_, err = model.CommissionPlanFind(g.Ex{"id": planID})
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	// 添加下级代理
	err = model.MemberInsert(name, password, remark, maintainName, groupName, agencyType, planID, createdAt, mr)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// Balance 查询钱包余额
func (that *MemberController) Balance(ctx *fasthttp.RequestCtx) {

	username := string(ctx.QueryArgs().Peek("username"))
	if !validator.CheckUName(username, 4, 9) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	balance, err := model.MemberBalance(username)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	data, err := jettison.Marshal(balance)
	if err != nil {
		helper.Print(ctx, false, helper.FormatErr)
		return
	}

	helper.PrintJson(ctx, true, string(data))
}

// 修改用户状态
func (that *MemberController) UpdateState(ctx *fasthttp.RequestCtx) {

	params := memberStateParam{}
	err := validator.Bind(ctx, &params)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	// 验证用户名
	names := strings.Split(params.Username, ",")
	for _, v := range names {
		if !validator.CheckUName(v, 4, 9) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}
	}

	err = model.MemberRemarkInsert("", params.Remark, admin["name"], names, ctx.Time().Unix())
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	err = model.MemberUpdateState(names, params.State)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

/**
 * @Description: List 会员列表
 * @Author: parker
 * @Date: 2021/4/14 16:38
 * @LastEditTime: 2021/4/14 19:00
 * @LastEditors: parker
 */
func (that *MemberController) List(ctx *fasthttp.RequestCtx) {

	ty := ctx.PostArgs().GetUintOrZero("ty")                  //1 批量匹配
	username := string(ctx.PostArgs().Peek("username"))       //会员帐号
	realname := string(ctx.PostArgs().Peek("realname"))       //会员姓名
	phone := string(ctx.PostArgs().Peek("phone"))             //手机号
	agent := string(ctx.PostArgs().Peek("agent"))             //代理帐号
	tag := string(ctx.PostArgs().Peek("tag"))                 //会员标签
	state := ctx.PostArgs().GetUintOrZero("state")            //状态 0:全部,1:启用,2:禁用
	regStartTime := string(ctx.PostArgs().Peek("start_time")) //注册开始时间
	regEndTime := string(ctx.PostArgs().Peek("end_time"))     //注册结束时间
	email := string(ctx.PostArgs().Peek("email"))             //邮箱
	ipFlag := ctx.PostArgs().GetUintOrZero("ip_flag")         //1:最近登录ip,2:注册IP
	ip := string(ctx.PostArgs().Peek("ip"))                   //精确ip
	deviceFlag := ctx.PostArgs().GetUintOrZero("device_flag") //设备类型1:登录设备号,2:注册设备号
	device := string(ctx.PostArgs().Peek("device"))           //设备号
	page := ctx.PostArgs().GetUintOrZero("page")              //页码
	pageSize := ctx.PostArgs().GetUintOrZero("page_size")     //每页数量

	if ty != 1 {

		r := map[int]bool{
			1: true,
			2: true,
		}
		if state > 0 {
			if _, ok := r[state]; !ok {
				helper.Print(ctx, false, helper.ParamErr)
				return
			}
		}

		if ipFlag > 0 {
			if _, ok := r[ipFlag]; !ok {
				helper.Print(ctx, false, helper.ParamErr)
				return
			}
		}

		if deviceFlag > 0 {
			if _, ok := r[deviceFlag]; !ok {
				helper.Print(ctx, false, helper.ParamErr)
				return
			}
		}
	}

	if page < 1 {
		page = 1
	}

	if pageSize < 10 || pageSize > 200 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	ex := g.Ex{}
	if username != "" {

		// 多个会员名用,分隔
		sName := strings.Split(username, ",")
		var usernames []string
		for _, name := range sName {
			if !validator.CheckUName(name, 4, 9) {
				helper.Print(ctx, false, helper.UsernameErr)
				return
			}

			usernames = append(usernames, name)
		}

		if ty == 0 && len(usernames) > 10 {
			ex["username"] = usernames[:10]
		} else {
			ex["username"] = usernames
		}

		data, err := model.MemberList(page, pageSize, "", "", "", ex)
		if err != nil {
			helper.Print(ctx, false, err.Error())
			return
		}

		helper.Print(ctx, true, data)
		return
	}

	if agent != "" {
		if !validator.CheckUName(agent, 4, 9) && len(username) > 50 {
			helper.Print(ctx, false, helper.AgentNameErr)
			return
		}

		ex["parent_name"] = agent
	}

	if ip != "" {

		if ipFlag == 2 {
			ex["regip"] = ip
		} else {
			ex["last_login_ip"] = ip
		}

	}

	if phone != "" {
		if !validator.CheckStringDigit(phone) {
			helper.Print(ctx, false, helper.PhoneFMTErr)
			return
		}

		ex["phone_hash"] = fmt.Sprintf("%d", model.MurmurHash(phone, 0))
	}

	if email != "" {
		if !strings.Contains(email, "@") {
			helper.Print(ctx, false, helper.EmailFMTErr)
			return
		}

		ex["email_hash"] = fmt.Sprintf("%d", model.MurmurHash(email, 0))
	}

	if state > 0 {
		ex["state"] = state
	}

	if realname != "" {
		ex["realname_hash"] = fmt.Sprintf("%d", model.MurmurHash(realname, 0))
	}

	if device != "" {
		// 最后登录设备号
		if deviceFlag == 1 {
			ex["last_login_device"] = device
		} else if deviceFlag == 2 { // 注册设备号
			ex["reg_device"] = device
		}
	}

	data, err := model.MemberList(page, pageSize, tag, regStartTime, regEndTime, ex)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *MemberController) Agency(ctx *fasthttp.RequestCtx) {

	username := string(ctx.PostArgs().Peek("username"))          //会员帐号
	maintainName := string(ctx.PostArgs().Peek("maintain_name")) //维护人
	state := ctx.PostArgs().GetUintOrZero("state")               //状态 0:全部,1:启用,2:禁用
	regStartTime := string(ctx.PostArgs().Peek("start_time"))    //注册开始时间
	regEndTime := string(ctx.PostArgs().Peek("end_time"))        //注册结束时间
	page := ctx.PostArgs().GetUintOrZero("page")                 //页码
	pageSize := ctx.PostArgs().GetUintOrZero("page_size")        //每页数量
	parentID := string(ctx.PostArgs().Peek("uid"))
	sortField := string(ctx.PostArgs().Peek("sort_field"))
	isAsc := ctx.PostArgs().GetUintOrZero("is_asc")

	if page < 1 {
		page = 1
	}

	if pageSize < 10 || pageSize > 200 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	var press = exp.NewExpressionList(exp.AndType, g.C("uid").Eq(g.C("top_uid")))
	if parentID != "" {
		press = exp.NewExpressionList(exp.AndType, g.Or(g.C("parent_uid").Eq(parentID), g.C("uid").Eq(parentID)))
	}

	if state > 0 {
		press = press.Append(g.C("state").Eq(state))
	}

	if username != "" {
		if !validator.CheckUName(username, 4, 9) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}
		press = press.Append(g.C("username").Eq(username))
	}

	if maintainName != "" {
		press = press.Append(g.C("maintain_name").Eq(maintainName))
	}

	if sortField != "" {
		sortFields := map[string]bool{
			"deposit":      true,
			"withdraw":     true,
			"valid_amount": true,
			"rebate":       true,
			"net_amount":   true,
		}

		if _, ok := sortFields[sortField]; !ok {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}

		if !validator.CheckIntScope(strconv.Itoa(isAsc), 0, 1) {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}
	}

	data, err := model.AgencyList(press, parentID, username, maintainName, regStartTime, regEndTime, sortField, isAsc, page, pageSize)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 修改会员信息
func (that *MemberController) Update(ctx *fasthttp.RequestCtx) {

	phone := string(ctx.PostArgs().Peek("phone"))
	email := string(ctx.PostArgs().Peek("email"))
	tagsID := string(ctx.PostArgs().Peek("tags_id"))
	realname := string(ctx.PostArgs().Peek("real_name"))
	username := string(ctx.PostArgs().Peek("username"))

	if !validator.CheckUName(username, 4, 9) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	param := map[string]string{}
	if realname != "" {
		param["realname"] = realname
	}

	if phone != "" {
		if !validator.IsVietnamesePhone(phone) {
			helper.Print(ctx, false, helper.PhoneFMTErr)
			return
		}

		param["phone"] = phone
	}

	if email != "" {
		if !strings.Contains(email, "@") {
			helper.Print(ctx, false, helper.EmailFMTErr)
			return
		}

		param["email"] = email
	}

	var userTagsId []string
	if tagsID != "" {
		if !validator.CheckStringCommaDigit(tagsID) {
			helper.Print(ctx, false, helper.UserTagErr)
			return
		}

		userTagsId = strings.Split(tagsID, ",")
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	err = model.MemberUpdate(username, admin["id"], param, userTagsId)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// Tags 用户标签
func (that *MemberController) Tags(ctx *fasthttp.RequestCtx) {

	uid := string(ctx.QueryArgs().Peek("uid"))
	if !validator.CheckStringDigit(uid) {
		helper.Print(ctx, false, helper.UIDErr)
		return
	}

	data, err := model.MemberTagsList(uid)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)

}

// SetTags 设置用户标签
// 为单个用户设置标签时batch=0,uid为用户id, 为多个用户批量设置标签的时候batch=1,uid用`,`分割
func (that *MemberController) SetTags(ctx *fasthttp.RequestCtx) {

	params := setTagParam{}
	err := validator.Bind(ctx, &params)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	// 校验tags，并拆解组装成slice
	var tags []string
	for _, v := range strings.Split(params.tags, ",") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}

		tags = append(tags, v)
	}
	if len(tags) == 0 {
		helper.Print(ctx, false, helper.UserTagErr)
		return
	}

	// 校验uid，并拆解组装成slice
	uids := strings.Split(params.uid, ",")
	var ids []string
	for _, v := range uids {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}

		ids = append(ids, v)
	}

	if len(ids) == 0 {
		helper.Print(ctx, false, helper.UIDErr)
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	err = model.MemberTagsSet(params.Batch, admin["id"], ids, tags, ctx.Time().Unix())
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// CancelTags 取消用户标签
// 取消单个用户标签时uid为用户id, 批量取消多个用户标签的时候uid用`,`分割
func (that *MemberController) CancelTags(ctx *fasthttp.RequestCtx) {

	params := setTagParam{}
	err := validator.Bind(ctx, &params)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	err = model.MemberTagsCancel(params.uid, params.tags)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 会员管理-会员列表-解除密码限制/解除短信限制/场馆钱包限制
func (that *MemberController) RetryReset(ctx *fasthttp.RequestCtx) {

	param := retryResetParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if param.Ty == model.WALLET {
		if !validator.CtypeDigit(param.Pid) {
			helper.Print(ctx, false, helper.PlatIDErr)
			return
		}
	}

	err = model.MemberRetryReset(param.Username, param.Ty, param.Pid)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 会员备注添加
func (that *MemberController) RemarkLogInsert(ctx *fasthttp.RequestCtx) {

	params := remarkLogParams{}
	err := validator.Bind(ctx, &params)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	if !validator.CheckStringLength(params.Msg, 1, 300) {
		helper.Print(ctx, false, helper.ContentLengthErr)
		return
	}

	if params.File != "" && !validator.CheckUrl(params.File) {
		helper.Print(ctx, false, helper.FileURLErr)
		return
	}

	if params.Username == "" {
		// 会员名校验
		if !validator.CheckUName(params.Username, 4, 9) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}
	}
	// 验证用户名
	names := strings.Split(params.Username, ",")
	for _, v := range names {
		if !validator.CheckUName(v, 4, 9) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}
	}

	err = model.MemberRemarkInsert(params.File, params.Msg, admin["name"], names, ctx.Time().Unix())
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 会员管理-会员列表-数据概览
func (that MemberController) Overview(ctx *fasthttp.RequestCtx) {

	username := string(ctx.QueryArgs().Peek("username"))
	startTime := string(ctx.QueryArgs().Peek("start_time"))
	endTime := string(ctx.QueryArgs().Peek("end_time"))

	if !validator.CheckUName(username, 4, 9) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	_, err := time.Parse("2006-01-02 15:04:05", startTime)
	if err != nil {
		helper.Print(ctx, false, helper.DateTimeErr)
		return
	}

	_, err = time.Parse("2006-01-02 15:04:05", endTime)
	if err != nil {
		helper.Print(ctx, false, helper.DateTimeErr)
		return
	}

	data, err := model.MemberDataOverview(username, startTime, endTime)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 会员日志列表
func (that *MemberController) RemarkLogList(ctx *fasthttp.RequestCtx) {

	uid := string(ctx.QueryArgs().Peek("uid"))
	adminName := string(ctx.QueryArgs().Peek("admin_name"))
	startTime := string(ctx.QueryArgs().Peek("start_time"))
	endTime := string(ctx.QueryArgs().Peek("end_time"))
	sPage := string(ctx.QueryArgs().Peek("page"))
	sPageSize := string(ctx.QueryArgs().Peek("page_size"))

	if !validator.CheckStringDigit(uid) {
		helper.Print(ctx, false, helper.UIDErr)
		return
	}

	if !validator.CheckStringDigit(sPage) || !validator.CheckStringDigit(sPageSize) {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if adminName != "" && !validator.CheckAName(adminName, 5, 20) {
		helper.Print(ctx, false, helper.AdminNameErr)
		return
	}

	if startTime != "" {
		_, err := time.Parse("2006-01-02 15:04:05", startTime)
		if err != nil {
			helper.Print(ctx, false, helper.DateTimeErr)
			return
		}
	}

	if endTime != "" {
		_, err := time.Parse("2006-01-02 15:04:05", endTime)
		if err != nil {
			helper.Print(ctx, false, helper.DateTimeErr)
			return
		}
	}

	page, _ := strconv.Atoi(sPage)
	pageSize, _ := strconv.Atoi(sPageSize)
	data, err := model.MemberRemarkLogList(uid, adminName, startTime, endTime, page, pageSize)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.PrintJson(ctx, true, data)
}

func (that *MemberController) UpdatePwd(ctx *fasthttp.RequestCtx) {

	username := string(ctx.PostArgs().Peek("username"))
	pwd := string(ctx.PostArgs().Peek("pwd"))
	ty := ctx.PostArgs().GetUintOrZero("ty")
	if username == "" || pwd == "" {
		helper.Print(ctx, false, helper.ParamNull)
		return
	}

	if ty > 1 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	// 会员名校验
	if !validator.CheckUName(username, 4, 9) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	// 会员密码校验
	if !validator.CheckUPassword(pwd, 8, 15) {
		helper.Print(ctx, false, helper.PasswordFMTErr)
		return
	}

	err := model.MemberUpdatePwd(username, pwd, ty, ctx)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// History 查询用户真实姓名/邮箱/手机号/银行卡号修改历史
func (that *MemberController) History(ctx *fasthttp.RequestCtx) {

	id := string(ctx.PostArgs().Peek("id"))
	field := string(ctx.PostArgs().Peek("field"))
	encrypt := ctx.PostArgs().GetBool("encrypt")

	if !validator.CheckStringDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	if _, ok := model.MemberHistoryField[field]; !ok {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	data, err := model.MemberHistory(id, field, encrypt)
	if err != nil {
		helper.Print(ctx, false, helper.ServerErr)
		return
	}

	helper.PrintJson(ctx, true, data)
}

// Full 查询用户真实姓名/邮箱/手机号/银行卡号明文信息
func (that *MemberController) Full(ctx *fasthttp.RequestCtx) {

	id := string(ctx.PostArgs().Peek("id"))
	field := string(ctx.PostArgs().Peek("field"))
	if !validator.CheckStringDigit(id) {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if _, ok := model.MemberHistoryField[field]; !ok {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	data, err := model.MemberFull(id, field)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// UpdateTopMember 修改密码以及返水比例
func (that *MemberController) UpdateTopMember(ctx *fasthttp.RequestCtx) {

	username := string(ctx.PostArgs().Peek("username"))
	password := string(ctx.PostArgs().Peek("password"))
	remarks := string(ctx.PostArgs().Peek("remarks"))
	sty := string(ctx.PostArgs().Peek("ty"))
	szr := string(ctx.PostArgs().Peek("zr"))
	sqp := string(ctx.PostArgs().Peek("qp"))
	sdj := string(ctx.PostArgs().Peek("dj"))
	sdz := string(ctx.PostArgs().Peek("dz"))
	state := ctx.PostArgs().GetUintOrZero("state") // 状态 1正常 2禁用
	planID := string(ctx.PostArgs().Peek("plan_id"))

	if !validator.CheckUName(username, 4, 9) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	if planID != "" {
		if !validator.CheckStringDigit(planID) {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}
	}

	mb, err := model.MemberFindOne(username)
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	MaxRebate, err := model.MemberMaxRebateFindOne(mb.UID)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	ty, err := decimal.NewFromString(sty) //下级会员体育返水比例
	if err != nil || ty.IsNegative() || ty.LessThan(MaxRebate.TY) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	zr, err := decimal.NewFromString(szr) //下级会员真人返水比例
	if err != nil || zr.IsNegative() || zr.LessThan(MaxRebate.ZR) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	qp, err := decimal.NewFromString(sqp) //下级会员棋牌返水比例
	if err != nil || qp.IsNegative() || qp.LessThan(MaxRebate.QP) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	dj, err := decimal.NewFromString(sdj) //下级会员电竞返水比例
	if err != nil || dj.IsNegative() || dj.LessThan(MaxRebate.DJ) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	dz, err := decimal.NewFromString(sdz) //下级会员电子返水比例
	if err != nil || dz.IsNegative() || dz.LessThan(MaxRebate.DZ) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	if mb.ParentUid != "0" && mb.ParentUid != "" {
		ParentRabte, err := model.MemberParentRebate(mb.ParentUid)
		if err != nil {
			helper.Print(ctx, false, err.Error())
			return
		}
		//大于上级棋牌返水比例
		if ParentRabte.QP.LessThan(qp) {
			helper.Print(ctx, false, helper.RebateOutOfRange)
			return
		}
		//大于上级体育返水比例
		if ParentRabte.TY.LessThan(ty) {
			helper.Print(ctx, false, helper.RebateOutOfRange)
			return
		}
		//大于上级真人返水比例
		if ParentRabte.ZR.LessThan(zr) {
			helper.Print(ctx, false, helper.RebateOutOfRange)
			return
		}
		//大于上级电子游戏返水比例
		if ParentRabte.DZ.LessThan(dz) {
			helper.Print(ctx, false, helper.RebateOutOfRange)
			return
		}
		//大于上级电竞返水比例
		if ParentRabte.DJ.LessThan(dj) {
			helper.Print(ctx, false, helper.RebateOutOfRange)
			return
		}
	}

	recd := g.Record{}
	if password != "" {
		if !validator.CheckUPassword(password, 8, 15) {
			helper.Print(ctx, false, helper.PasswordFMTErr)
			return
		}
		recd["password"] = fmt.Sprintf("%d", model.MurmurHash(password, mb.CreatedAt))
	}

	if state != 0 {
		if state > 2 || state < 1 {
			helper.Print(ctx, false, helper.PasswordFMTErr)
			return
		}
		recd["state"] = state
	}

	if remarks != "" {
		recd["remarks"] = remarks
	}

	mr := model.MemberRebate{
		TY: ty.StringFixed(1),
		ZR: zr.StringFixed(1),
		QP: qp.StringFixed(1),
		DJ: dj.StringFixed(1),
		DZ: dz.StringFixed(1),
	}

	// 更新代理
	err = model.MemberUpdateInfo(mb.UID, planID, recd, mr)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// UpdateMaintainName 修改维护人
func (that *MemberController) UpdateMaintainName(ctx *fasthttp.RequestCtx) {

	username := string(ctx.PostArgs().Peek("username"))
	maintainName := string(ctx.PostArgs().Peek("maintain_name"))

	if !validator.CheckUName(username, 4, 9) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	if !validator.CtypeAlnum(maintainName) {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	mb, err := model.MemberFindOne(username)
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	// 更新代理
	err = model.MemberUpdateMaintanName(mb.UID, maintainName)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *MemberController) MemberList(ctx *fasthttp.RequestCtx) {
	param := model.MemberListParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if param.Username != "" {
		if !validator.CheckUName(param.Username, 4, 9) {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}
	}

	if param.ParentName != "" {
		if !validator.CheckUName(param.ParentName, 4, 9) {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}
	}

	data, err := model.AgencyMemberList(param)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}
