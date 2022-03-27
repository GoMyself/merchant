package router

import (
	"fmt"
	"merchant2/controller"
	"runtime/debug"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

var (
	ApiTimeoutMsg = `{"status": "false","data":"服务器响应超时，请稍后重试"}`
	ApiTimeout    = time.Second * 30
	router        *fasthttprouter.Router
	buildInfo     BuildInfo
)

type BuildInfo struct {
	GitReversion   string
	BuildTime      string
	BuildGoVersion string
}

func apiServerPanic(ctx *fasthttp.RequestCtx, rcv interface{}) {

	err := rcv.(error)
	fmt.Println(err)
	debug.PrintStack()

	if r := recover(); r != nil {
		fmt.Println("recovered failed", r)
	}

	ctx.SetStatusCode(500)
	return
}

func Version(ctx *fasthttp.RequestCtx) {

	ctx.SetContentType("text/html; charset=utf-8")
	fmt.Fprintf(ctx, "merchant<br />Git reversion = %s<br />Build Time = %s<br />Go version = %s<br />System Time = %s<br />",
		buildInfo.GitReversion, buildInfo.BuildTime, buildInfo.BuildGoVersion, ctx.Time())

	//ctx.Request.Header.VisitAll(func (key, value []byte) {
	//	fmt.Fprintf(ctx, "%s: %s<br/>", string(key), string(value))
	//})
}

// SetupRouter 设置路由列表
func SetupRouter(b BuildInfo) *fasthttprouter.Router {

	router = fasthttprouter.New()
	router.PanicHandler = apiServerPanic

	buildInfo = b

	groupCtl := new(controller.GroupController)
	// 权限管理
	privCtl := new(controller.PrivController)
	// 用户管理
	adminCtl := new(controller.AdminController)
	// tree
	treeCtl := new(controller.TreeController)
	//ip区域
	areaCtl := new(controller.AreaController)
	//场馆列表
	platformCtl := new(controller.PlatformController)
	// 游戏列表
	slotsCtl := new(controller.SlotsController)
	//会员
	memberCtl := new(controller.MemberController)
	// 账户调整
	adjustCtl := new(controller.AdjustController)
	//银行卡管理
	bankCtl := new(controller.BankcardController)
	//红利
	dividendCtl := new(controller.DividendController)
	// 标签管理
	tagsCtl := new(controller.TagsController)
	// 广告
	bannerCtl := new(controller.BannerController)
	// app 版本管理
	appUpCtl := new(controller.AppUpgradeController)
	//黑名单管理
	blacklistCtl := new(controller.BlacklistController)
	// 日志管理
	logCtl := new(controller.LogController)
	// 返水管理
	rebateCtl := new(controller.RebateController)
	//公告管理
	noticeCtl := new(controller.NoticeController)
	//转账，游戏，账变记录
	recordCtl := new(controller.RecordController)
	// 佣金
	commissionCtl := new(controller.CommissionController)

	get("/merchant/version", Version)

	// 权限管理-用户组管理-新增分组
	post("/merchant/group/insert", groupCtl.Insert)
	// 权限管理-用户组管理-修改分组
	post("/merchant/group/update", groupCtl.Update)
	// 权限管理-用户组管理列表
	get("/merchant/group/list", groupCtl.List)
	// 权限管理-获取分组列表
	get("/merchant/priv/list", privCtl.List)

	// 后台用户登录
	post("/merchant/admin/login", adminCtl.Login)
	// 权限管理-后台账号管理-后台账号列表
	get("/merchant/admin/list", adminCtl.List)
	// 权限管理-后台账号管理-新建账号
	post("/merchant/admin/insert", adminCtl.Insert)
	// 权限管理-后台账号管理-编辑更新账号
	post("/merchant/admin/update", adminCtl.Update)
	// 权限管理-后台账号管理-启用
	// 权限管理-后台账号管理-禁用
	get("/merchant/admin/update/state", adminCtl.UpdateState)

	//获取静态json
	get("/merchant/tree", treeCtl.List)
	//ip查询
	post("/merchant/area/view", areaCtl.View)

	// 站点管理-场馆管理
	get("/merchant/platform/list", platformCtl.List)
	// 站点管理-场馆编辑
	// 站点管理-锁定钱包
	// 站点管理-维护
	post("/merchant/platform/update", platformCtl.Update)
	// 站点管理-场馆列表
	get("/merchant/platform/plats", platformCtl.PlatList)
	// 场馆费率
	get("/merchant/platform/rate", platformCtl.PlatRate)

	// 站点管理-场馆管理-游戏列表
	post("/merchant/slots/list", slotsCtl.List)
	// 站点管理-场馆管理-游戏列表-编辑游戏
	post("/merchant/slots/update", slotsCtl.Update)
	// 站点管理-场馆管理-游戏列表-上线下线
	post("/merchant/slots/updatestate", slotsCtl.UpdateState)

	// 会员管理-会员列表-新增总代
	post("/merchant/member/insert", memberCtl.Insert)
	// 会员管理-会员列表-帐户信息
	get("/merchant/member/info", memberCtl.AccountInfo)
	// 会员管理-会员中心钱包余额
	get("/merchant/member/balance", memberCtl.Balance)
	//批量获取用户余额
	post("/merchant/member/balancebatch", memberCtl.BalanceBatch)
	//批量获取用户标签
	post("/merchant/member/tagsbatch", memberCtl.TagBatch)
	// 会员管理-会员列表-修改会员状态 (批量禁用启用)
	post("/merchant/member/updatestate", memberCtl.UpdateState)
	// 会员管理-会员列表-修改会员信息
	post("/merchant/member/update", memberCtl.Update)
	// 会员管理-会员列表
	post("/merchant/member/list", memberCtl.List)
	// 代理管理-代理列表
	post("/merchant/agency/list", memberCtl.Agency)
	// 代理管理-代理编辑 总代
	post("/merchant/agency/update", memberCtl.UpdateTopMember)
	// 代理管理-代理编辑维护人
	post("/merchant/agency/updatemaintain", memberCtl.UpdateMaintainName)
	// 会员管理-会员列表-用户标签
	get("/merchant/member/tags", memberCtl.Tags)
	// 会员管理-会员列表-批量添加标签/编辑标签
	post("/merchant/member/settags", memberCtl.SetTags)
	//会员列表-修改会员密码
	post("/merchant/member/updatepwd", memberCtl.UpdatePwd)
	// 会员管理-会员列表-批量取消标签
	post("/merchant/member/tags/cancel", memberCtl.CancelTags)
	// 会员管理-会员列表-解除密码限制/接触短信限制/场馆钱包限制
	get("/merchant/member/retry/reset", memberCtl.RetryReset)
	// 会员管理-会员列表-添加备注
	post("/merchant/member/remark/insert", memberCtl.RemarkLogInsert)
	// 会员管理-会员列表-数据概览
	get("/merchant/member/overview", memberCtl.Overview)
	// 会员管理-会员列表-基本信息-备注信息
	get("/merchant/member/remark/list", memberCtl.RemarkLogList)
	// 查询用户真实姓名/邮箱/手机号/银行卡号修改历史
	post("/merchant/member/history", memberCtl.History)
	// 查询用户真实姓名/邮箱/手机号/银行卡号明文信息
	post("/merchant/member/full", memberCtl.Full)

	// 会员列表-账户调整
	post("/merchant/adjust/insert", adjustCtl.Insert)
	// 会员管理-账户调整审核列表
	get("/merchant/adjust/list", adjustCtl.List)
	// 后台会员管理-账户调整审核
	post("/merchant/adjust/review", adjustCtl.Review)

	//新增银行卡
	post("/merchant/bankcard/insert", bankCtl.Insert)
	//查询银行卡
	post("/merchant/bankcard/list", bankCtl.List)
	//修改银行卡信息
	post("/merchant/bankcard/update", bankCtl.Update)
	//删除银行卡
	get("/merchant/bankcard/delete", bankCtl.Delete)

	// 运营管理-红利管理-单会员发放
	post("/merchant/dividend/insert", dividendCtl.Insert)
	// 运营管理-红利管理-审核列表
	get("/merchant/dividend/list", dividendCtl.List)
	// 运营管理-会员管理-基本信息-红利列表
	post("/merchant/dividend/member/list", dividendCtl.MemberList)
	// 运营管理-红利管理-更新
	post("/merchant/dividend/update", dividendCtl.Update)
	// 运营管理-红利管理-修改审核备注
	post("/merchant/dividend/review", dividendCtl.ReviewRemarkUpdate)

	// 会员管理-会员配置-标签管理-新增
	post("/merchant/tags/insert", tagsCtl.Insert)
	// 会员管理-会员配置-标签管理-列表
	post("/merchant/tags/list", tagsCtl.List)
	// 会员管理-会员配置-标签管理-修改
	post("/merchant/tags/update", tagsCtl.Update)
	// 会员管理-会员配置-标签管理-删除
	get("/merchant/tags/delete", tagsCtl.Delete)

	// 运营管理-广告管理-新增
	post("/merchant/banner/insert", bannerCtl.Insert)
	// 运营管理-广告管理-修改
	post("/merchant/banner/update", bannerCtl.Update)
	// 运营管理-广告管理-列表
	post("/merchant/banner/list", bannerCtl.List)
	// 运营管理-广告管理-删除
	get("/merchant/banner/delete", bannerCtl.Delete)
	// 运营管理-广告管理-状态(启用|停用)
	post("/merchant/banner/updatestate", bannerCtl.UpdateState)
	// 运营管理-广告管理-上传图片
	post("/merchant/banner/upload", bannerCtl.UploadFile)

	// App 升级配置更新
	post("/merchant/app/update", appUpCtl.Update)
	// App 升级配置列表
	get("/merchant/app/list", appUpCtl.List)

	//查询会员登录日志
	get("/merchant/blacklist/assoclog", blacklistCtl.AssociateList)
	//查询会员登录日志
	get("/merchant/blacklist/loginlog", blacklistCtl.LogList)

	//风控管理-黑名单查询
	get("/merchant/blacklist/list", blacklistCtl.List)
	//风控管理-黑名单新增
	post("/merchant/blacklist/insert", blacklistCtl.Insert)
	//风控管理-黑名单修改
	post("/merchant/blacklist/update", blacklistCtl.Update)
	//风控管理-黑名单删除
	get("/merchant/blacklist/delete", blacklistCtl.Delete)

	// 系统管理-日志管理-登录日志
	get("/merchant/sys/log/login/list", logCtl.AdminLoginLog)
	// 系统管理-日志管理-系统日志
	post("/merchant/sys/log/system/list", logCtl.SystemLog)

	//运营管理-内容管理-系统公告-增加
	post("/merchant/notice/insert", noticeCtl.Insert)
	//运营管理-内容管理-系统公告-列表
	post("/merchant/notice/list", noticeCtl.List)
	//运营管理-内容管理-系统公告-编辑
	post("/merchant/notice/update", noticeCtl.Update)
	//运营管理-内容管理-系统公告-停用/启用
	post("/merchant/notice/updatestate", noticeCtl.UpdateState)
	//运营管理-内容管理-系统公告-删除
	get("/merchant/notice/delete", noticeCtl.Delete)

	// 获取返水上限
	get("/merchant/rebate/scale", rebateCtl.Scale)

	// 会员管理-会员列表-账变记录
	get("/merchant/record/transaction", recordCtl.Transaction)
	// 记录管理-平台转帐
	post("/merchant/record/transfer", recordCtl.Transfer)
	// 会员管理-会员列表-有效投注查询
	// 会员管理-投注管理
	// 会员管理-投注管理-会员游戏记录详情列表
	// 会员列表->输赢信息
	// 会员列表->输赢信息->场馆合并
	// 会员列表->输赢信息->日期合并
	post("/merchant/record/game", recordCtl.RecordGame)

	//总代佣金列表
	get("/merchant/commission/list", commissionCtl.TopList)
	// 发放总代佣金
	post("/merchant/commission/ration", commissionCtl.Ration)
	// 代理管理-佣金记录-列表
	get("/merchant/commission/record/list", commissionCtl.RecordList)
	// 代理管理-佣金记录-审核
	post("/merchant/commission/record/review", commissionCtl.RecordReview)

	// 添加佣金方案
	post("/merchant/commission/plan/insert", commissionCtl.PlanInsert)
	// 更新佣金方案
	post("/merchant/commission/plan/update", commissionCtl.PlanUpdate)
	// 佣金方案列表
	post("/merchant/commission/plan/list", commissionCtl.PlanList)
	// 佣金方案详情
	get("/merchant/commission/plan/detail", commissionCtl.PlanDetail)
	return router
}

// get is a shortcut for router.GET(path string, handle fasthttp.RequestHandler)
func get(path string, handle fasthttp.RequestHandler) {
	router.GET(path, fasthttp.TimeoutHandler(handle, ApiTimeout, ApiTimeoutMsg))
}

// head is a shortcut for router.HEAD(path string, handle fasthttp.RequestHandler)
func head(path string, handle fasthttp.RequestHandler) {
	router.HEAD(path, fasthttp.TimeoutHandler(handle, ApiTimeout, ApiTimeoutMsg))
}

// options is a shortcut for router.OPTIONS(path string, handle fasthttp.RequestHandler)
func options(path string, handle fasthttp.RequestHandler) {
	router.OPTIONS(path, fasthttp.TimeoutHandler(handle, ApiTimeout, ApiTimeoutMsg))
}

// post is a shortcut for router.POST(path string, handle fasthttp.RequestHandler)
func post(path string, handle fasthttp.RequestHandler) {
	router.POST(path, fasthttp.TimeoutHandler(handle, ApiTimeout, ApiTimeoutMsg))
}

// put is a shortcut for router.PUT(path string, handle fasthttp.RequestHandler)
func put(path string, handle fasthttp.RequestHandler) {
	router.PUT(path, fasthttp.TimeoutHandler(handle, ApiTimeout, ApiTimeoutMsg))
}

// patch is a shortcut for router.PATCH(path string, handle fasthttp.RequestHandler)
func patch(path string, handle fasthttp.RequestHandler) {
	router.PATCH(path, fasthttp.TimeoutHandler(handle, ApiTimeout, ApiTimeoutMsg))
}

// delete is a shortcut for router.DELETE(path string, handle fasthttp.RequestHandler)
func delete(path string, handle fasthttp.RequestHandler) {
	router.DELETE(path, fasthttp.TimeoutHandler(handle, ApiTimeout, ApiTimeoutMsg))
}
