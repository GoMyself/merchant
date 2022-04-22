package model

import (
	"database/sql"
	"errors"
	"fmt"
	"merchant2/contrib/helper"
	"merchant2/contrib/session"
	"merchant2/contrib/validator"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/nwf2013/schema"
	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/olivere/elastic/v7"
	"github.com/shopspring/decimal"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"github.com/wI2L/jettison"
)

var (
	PWD            = uint8(1) // 密码
	SMS            = uint8(2) // 短信
	WALLET         = uint8(3) // 钱包
	memberUnfreeze = map[uint8]string{
		PWD:    "MPE:%s",    // 密码尝试次数
		SMS:    "smsfgt:%s", // 短信发送次数
		WALLET: "P:%s:%s",   // 钱包限制
	}
	MemberHistoryField = map[string]bool{
		"realname": true, // 会员真实姓名
		"phone":    true, // 会员手机号
		"email":    true, // 会员邮箱
		"bankcard": true, // 会员银行卡查询
	}
)

type PlatBalance struct {
	ID      string `db:"id" json:"id"`
	Balance string `db:"balance" json:"balance"`
}

type memberTags struct {
	Uid     string `db:"uid" json:"uid"`
	TagId   int64  `db:"tag_id" json:"tags_id"`
	TagName string `db:"tag_name" json:"tag_name"`
}

type tag struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type memberTag struct {
	Uid  string `db:"uid" json:"uid"`
	Tags []tag  `json:"tags"`
}

type memberDeviceReg struct {
	RegDevice string `db:"reg_device" json:"reg_device"`
	Uid       uint64 `db:"uid" json:"uid"`
}

// MemberDataOverviewData 会员管理-会员列表-数据概览 response structure
type MemberDataOverviewData struct {
	NetAmount      float64 `json:"net_amount"`       // 总输赢
	ValidBetAmount float64 `json:"valid_bet_amount"` // 总有效投注
	Deposit        float64 `json:"deposit"`          // 总存款
	Withdraw       float64 `json:"withdraw"`         // 总提款
	Dividend       float64 `json:"dividend"`         // 总红利
	Rebate         float64 `json:"rebate"`           // 总返水
}

// MemberListData 会员列表
type MemberListData struct {
	T    int                      `json:"t"`
	S    int                      `json:"s"`
	D    []MemberListCol          `json:"d"`
	Agg  map[string]MemberAggData `json:"agg"`
	Info map[string]memberInfo    `json:"info"`
}

type memberInfo struct {
	UID          string `db:"uid" json:"uid"`
	Username     string `db:"username" json:"username"`           //会员名
	LastLoginIp  string `db:"last_login_ip" json:"last_login_ip"` //最后登陆ip
	TopUid       string `db:"top_uid" json:"top_uid"`             //总代uid
	TopName      string `db:"top_name" json:"top_name"`           //总代代理
	ParentUid    string `db:"parent_uid" json:"parent_uid"`       //上级uid
	ParentName   string `db:"parent_name" json:"parent_name"`     //上级代理
	State        uint8  `db:"state" json:"state"`                 //状态 1正常 2禁用
	Remarks      string `db:"remarks" json:"remarks"`             //备注
	MaintainName string `db:"maintain_name" json:"maintain_name"` //
}

type MemberListCol struct {
	UID         string  `json:"uid" db:"uid"`
	Deposit     float64 `json:"deposit" db:"deposit"`
	Withdraw    float64 `json:"withdraw" db:"withdraw"`
	ValidAmount float64 `json:"valid_amount" db:"valid_amount"`
	Rebate      float64 `json:"rebate" db:"rebate"`
	NetAmount   float64 `json:"net_amount" db:"net_amount"`
	TY          string  `json:"ty" db:"ty"`
	ZR          string  `json:"zr" db:"zr"`
	QP          string  `json:"qp" db:"qp"`
	DJ          string  `json:"dj" db:"dj"`
	DZ          string  `json:"dz" db:"dz"`
	CP          string  `json:"cp" db:"cp"`
	Lvl         int     `json:"lvl" db:"-"`
	PlanID      string  `json:"plan_id" db:"-"`
	PlanName    string  `json:"plan_name" db:"-"`
}

type MemberAggData struct {
	MemCount       int    `db:"mem_count" json:"mem_count"`
	RegistCountNew int    `db:"regist_count" json:"regist_count"`
	UID            string `db:"uid" json:"uid"`
}

type MemberRebateResult_t struct {
	ZR decimal.Decimal
	QP decimal.Decimal
	TY decimal.Decimal
	DZ decimal.Decimal
	DJ decimal.Decimal
	CP decimal.Decimal
}

func MemberInsert(username, password, remark, maintainName, groupName, agencyType, planID string, createdAt uint32, mr MemberRebate) error {

	userName := strings.ToLower(username)
	if MemberExist(userName) {
		return errors.New(helper.UsernameExist)
	}

	uid := helper.GenId()
	mr.UID = uid
	mr.CreatedAt = createdAt
	mr.ParentUID = "0"
	mr.Prefix = meta.Prefix
	agtype, _ := strconv.ParseInt(agencyType, 10, 64)
	m := Member{
		UID:                uid,
		Username:           userName,
		Password:           fmt.Sprintf("%d", MurmurHash(password, createdAt)),
		Prefix:             meta.Prefix,
		State:              1,
		CreatedAt:          createdAt,
		LastLoginIp:        "",
		LastLoginAt:        createdAt,
		LastLoginDevice:    "",
		LastLoginSource:    0,
		ParentUid:          "0",
		TopUid:             uid,
		TopName:            userName,
		FirstDepositAmount: "0.000",
		FirstBetAmount:     "0.000",
		Balance:            "0.000",
		LockAmount:         "0.000",
		Commission:         "0.000",
		Remarks:            remark,
		MaintainName:       maintainName,
		GroupName:          groupName,
		AgencyType:         agtype,
	}

	tx, err := meta.MerchantDB.Begin() // 开启事务
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	query, _, _ := dialect.Insert("tbl_members").Rows(&m).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	query, _, _ = dialect.Insert("tbl_member_rebate_info").Rows(&mr).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	treeNode := MemberClosureInsert(uid, "0")
	_, err = tx.Exec(treeNode)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(fmt.Errorf("sql : %s, error : %s", treeNode, err.Error()), helper.DBErr)
	}

	// 维护佣金方案
	recd := g.Record{
		"id":      helper.GenId(),
		"uid":     uid,
		"plan_id": planID,
		"prefix":  meta.Prefix,
	}
	query, _, _ = dialect.Insert("tbl_commission_conf").Rows(recd).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	_, err = session.Set([]byte(m.Username), m.UID)
	if err != nil {
		return errors.New(helper.SessionErr)
	}

	return nil
}

/**
 * @Description: Transfer 会员列表-帐户信息
 * @Author: parker
 * @Date: 2021/4/7 10:43
 * @LastEditTime: 2021/4/7 10:43
 * @LastEditors: parker
 */
func MemberAccountInfo(username string) ([]PlatBalance, error) {

	var data []PlatBalance

	mb, err := MemberFindOne(username)
	if err != nil || len(mb.Username) == 0 {
		return data, errors.New(helper.UsernameErr)
	}

	//添加中心钱包余额
	data = append(data, PlatBalance{ID: "1", Balance: mb.Balance})
	data = append(data, memberPlatformBalance(username)...)

	return data, nil
}

func MemberExist(username string) bool {

	var uid uint64
	query, _, _ := dialect.From("tbl_members").Select("uid").Where(g.Ex{"username": username, "prefix": meta.Prefix}).ToSQL()
	err := meta.MerchantDB.Get(&uid, query)
	if err == nil && uid != 0 {
		return true
	}

	return false
}

// 批量获取会员标签
func MemberBatchTag(uids []string) (string, error) {

	var tags []memberTags

	t := dialect.From("tbl_member_tags")
	query, _, _ := t.Select("uid", "tag_id", "tag_name").Where(g.Ex{"uid": uids, "prefix": meta.Prefix}).ToSQL()
	err := meta.MerchantDB.Select(&tags, query)
	if err != nil {
		return "", pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	var result []memberTag
	for _, id := range uids {
		item := memberTag{Uid: id, Tags: []tag{}}
		for _, v := range tags {
			if strings.EqualFold(v.Uid, id) {
				item.Tags = append(item.Tags, tag{ID: v.TagId, Name: v.TagName})
			}
		}

		result = append(result, item)
	}

	data, err := jettison.Marshal(result)
	if err != nil {
		return "", errors.New(helper.FormatErr)
	}

	return string(data), nil
}

// 更新用户状态
func MemberUpdateState(sliceName []string, state int8) error {

	query, _, _ := dialect.Update("tbl_members").
		Set(g.Record{"state": state}).Where(g.Ex{"username": sliceName, "prefix": meta.Prefix}).ToSQL()
	_, err := meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	// 更新用户redis state
	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()

	for k := range sliceName {
		pipe.HSet(ctx, sliceName[k], "state", state)
		pipe.Persist(ctx, sliceName[k])
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	return nil
}

/**
 * @Description: MemberList 会员列表
 * @Author: parker
 * @Date: 2021/4/14 16:38
 * @LastEditTime: 2021/4/14 19:00
 * @LastEditors: parker
 */
func MemberList(page, pageSize int, tag, startTime, endTime string, ex g.Ex) (MemberPageData, error) {

	data := MemberPageData{}
	var err error
	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.TimeToLoc(endTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}
		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}

	if tag != "" {

		fsc := elastic.NewFetchSourceContext(true).Include("uid")
		boolQuery := elastic.NewBoolQuery()
		boolQuery.Must(elastic.NewTermQuery("prefix", meta.Prefix))
		boolQuery.Filter(elastic.NewWildcardQuery("tag_name", fmt.Sprintf("*%s*", tag)))
		distinct := elastic.NewCollapseBuilder("uid")

		resTag, err := meta.ES.Search(esPrefixIndex("tbl_member_tags")).FetchSourceContext(fsc).Query(boolQuery).Collapse(distinct).Size(10000).Do(ctx)
		if err != nil {
			return data, pushLog(err, "es")
		}

		var p fastjson.Parser
		var ids []uint64
		for _, v := range resTag.Hits.Hits {

			tag, err := p.ParseBytes(v.Source)
			if err != nil {
				return data, errors.New(helper.FormatErr)
			}

			ids = append(ids, tag.GetUint64("uid"))
		}

		if len(ids) == 0 {
			return data, nil
		}

		ex["uid"] = ids
	}

	data, err = memberList(page, pageSize, ex)
	if err != nil {
		return data, err
	}

	if len(data.D) < 1 {
		return data, nil
	}

	var res []schema.Dec_t
	for _, v := range data.D {
		nameRecs := schema.Dec_t{
			Field: "realname",
			Hide:  true,
			ID:    v.UID,
		}
		res = append(res[0:], nameRecs)
		emailRecs := schema.Dec_t{
			Field: "email",
			Hide:  true,
			ID:    v.UID,
		}
		res = append(res[0:], emailRecs)
		phoneRecs := schema.Dec_t{
			Field: "phone",
			Hide:  true,
			ID:    v.UID,
		}
		res = append(res[0:], phoneRecs)
	}

	record, err := rpcGet(res)
	if err != nil {
		return data, errors.New(helper.GetRPCErr)
	}

	rpcLen := len(record)

	for k := range data.D {

		data.D[k].Password = ""
		data.D[k].RealName = ""
		if rpcLen > k*3+0 && record[k*3].Err == "" {
			data.D[k].RealName = record[k*3].Res
		}

		data.D[k].Email = ""
		if rpcLen > k*3+1 && record[k*3+1].Err == "" {
			data.D[k].Email = record[k*3+1].Res
		}

		data.D[k].Phone = ""
		if rpcLen > k*3+2 && record[k*3+2].Err == "" {
			data.D[k].Phone = record[k*3+2].Res
		}
	}

	return data, nil
}

func memberList(page, pageSize int, ex g.Ex) (MemberPageData, error) {

	data := MemberPageData{}
	ex["prefix"] = meta.Prefix
	t := dialect.From("tbl_members")
	if page == 1 {
		totalQuery, _, _ := t.Select(g.COUNT("uid")).Where(ex).ToSQL()
		err := meta.MerchantDB.Get(&data.T, totalQuery)
		if err != nil && err != sql.ErrNoRows {
			return data, pushLog(err, helper.DBErr)
		}

		if data.T == 0 {
			return data, nil
		}
	}
	offset := (page - 1) * pageSize
	var d []Member
	query, _, _ := t.Select(colsMember...).Where(ex).Order(g.I("created_at").Desc()).Offset(uint(offset)).Limit(uint(pageSize)).ToSQL()
	err := meta.MerchantDB.Select(&d, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	if len(d) == 0 {
		return data, nil
	}

	for _, v := range d {
		val := MemberData{Member: v}
		data.D = append(data.D[0:], val)
	}
	data.S = uint(pageSize)
	return data, nil
}

func AgencyList(ex exp.ExpressionList, parentID, username, maintainName, startTime, endTime, sortField string, isAsc, page, pageSize, agencyType int) (MemberListData, error) {

	data := MemberListData{}
	startAt, err := helper.TimeToLoc(startTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	endAt, err := helper.TimeToLoc(endTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	if startAt >= endAt {
		return data, errors.New(helper.QueryTimeRangeErr)
	}

	data.S = pageSize

	if sortField != "" && username == "" { // 排序
		data.D, data.T, err = memberListSort(ex, parentID, sortField, startAt, endAt, isAsc, page, pageSize)
		if err != nil {
			return data, err
		}
	} else {
		data.D, data.T, err = agencyList(ex, startAt, endAt, page, pageSize, parentID, agencyType)
		if err != nil {
			return data, err
		}
	}

	if len(data.D) == 0 {
		return data, nil
	}

	var (
		ids []string
	)
	for _, v := range data.D {
		ids = append(ids, v.UID)

	}

	// 获取用户状态 最后登录ip
	members, err := memberInfoFindBatch(ids)
	if err != nil {
		return data, err
	}

	data.Info = members

	// 获取用户的反水比例
	rebate, err := MemberRebateSelect(ids)
	if err != nil {
		return data, err
	}

	lvParams := make(map[string]string)
	for _, member := range members {
		lvParams[member.UID] = member.TopUid
	}

	// 获取代理层级  佣金方案
	lvls := memberLvl(lvParams)

	// 佣金方案
	plans, err := memberPlan(ids)
	if err != nil {
		return data, err
	}

	for i, v := range data.D {
		if rb, ok := rebate[v.UID]; ok {
			data.D[i].DJ = rb.DJ
			data.D[i].TY = rb.TY
			data.D[i].ZR = rb.ZR
			data.D[i].QP = rb.QP
			data.D[i].DZ = rb.DZ
			data.D[i].CP = rb.CP
		}

		if lv, ok := lvls[v.UID]; ok {
			data.D[i].Lvl = lv
		}

		if plan, ok := plans[v.UID]; ok {
			if planID, ok := plan["plan_id"]; ok {
				data.D[i].PlanID = planID
			}

			if plannName, ok := plan["name"]; ok {
				data.D[i].PlanName = plannName
			}
		}

	}

	// 直属下级人数 新增注册人数
	data.Agg, err = MemberAgg(ids, startAt, endAt)
	return data, err
}

// 获取佣金方案
func memberPlan(ids []string) (map[string]map[string]string, error) {

	ex := g.Ex{
		"uid":    ids,
		"prefix": meta.Prefix,
	}

	var conf []CommssionConf
	query, _, _ := dialect.From("tbl_commission_conf").Select("id", "uid", "plan_id").Where(ex).ToSQL()
	err := meta.MerchantDB.Select(&conf, query)
	if err != nil && err != sql.ErrNoRows {
		return nil, pushLog(err, helper.DBErr)
	}

	if len(conf) == 0 {
		return nil, nil
	}

	var planID []string
	for _, v := range conf {
		planID = append(planID, v.PlanID)
	}

	var plans []CommissionPlan
	ex = g.Ex{
		"id":     planID,
		"prefix": meta.Prefix,
	}
	query, _, _ = dialect.From("tbl_commission_plan").Select(colsCommPlan...).Where(ex).ToSQL()
	err = meta.MerchantDB.Select(&plans, query)
	if err != nil && err != sql.ErrNoRows {
		return nil, pushLog(err, helper.DBErr)
	}

	if len(plans) == 0 {
		return nil, nil
	}

	planMap := make(map[string]string)
	for _, v := range plans {
		planMap[v.ID] = v.Name
	}

	data := make(map[string]map[string]string)
	for _, v := range conf {

		data[v.UID] = map[string]string{
			"plan_id": v.PlanID,
		}

		if name, ok := planMap[v.PlanID]; ok {
			data[v.UID]["name"] = name
		}

	}

	return data, nil
}

// 获取代理层级
func memberLvl(params map[string]string) map[string]int {

	var or []exp.Expression

	for k, v := range params {
		or = append(or, g.And(
			g.C("ancestor").Eq(v),   // 总代id
			g.C("descendant").Eq(k), // 代理id
		))
	}

	var trees []MembersTree
	query, _, _ := dialect.From("tbl_members_tree").Where(g.Or(or...)).ToSQL()
	err := meta.MerchantDB.Select(&trees, query)
	if err != nil {
		return nil
	}

	data := make(map[string]int, len(trees))
	for _, v := range trees {
		data[v.Descendant] = v.Lvl
	}

	return data
}

func memberListSort(ex exp.ExpressionList, parentID, sortField string, startAt, endAt int64, isAsc, page, pageSize int) ([]MemberListCol, int, error) {

	var data []MemberListCol

	exC := g.Ex{
		"report_time": g.Op{"between": exp.NewRangeVal(startAt, endAt)},
		"report_type": 2, // 投注时间2结算时间3投注时间月报4结算时间月报
		"prefix":      meta.Prefix,
	}

	ex = ex.Append(exC)

	number := 0
	if page == 1 {

		query, _, _ := dialect.From("tbl_report_agency").Select(g.COUNT(g.DISTINCT("uid"))).Where(ex).ToSQL()
		err := meta.ReportDB.Get(&number, query)
		if err != nil && err != sql.ErrNoRows {
			return data, 0, pushLog(err, helper.DBErr)
		}

		if number == 0 {
			return data, 0, nil
		}
	}

	orderField := g.L("report_time")
	if sortField != "" {
		orderField = g.L(sortField)
	}

	orderBy := orderField.Desc()
	if isAsc == 1 {
		orderBy = orderField.Asc()
	}

	and := g.And(ex, g.C("uid").Neq(g.C("parent_uid")))
	if parentID != "" {
		and = g.And(
			exC,
			g.Or(
				g.And(
					g.C("uid").In(parentID),
					g.C("data_type").Eq(2),
				),
				g.And(
					g.C("data_type").Eq(1),
					g.C("parent_uid").Eq(parentID),
				),
			),
		)
	}

	offset := (page - 1) * pageSize
	query, _, _ := dialect.From("tbl_report_agency").Select(
		"uid",
		g.SUM("deposit_amount").As("deposit"),
		g.SUM("withdrawal_amount").As("withdraw"),
		g.SUM("valid_bet_amount").As("valid_amount"),
		g.SUM("rebate_amount").As("rebate"),
		g.SUM("company_net_amount").As("net_amount"),
	).GroupBy("uid").
		Where(and).
		Offset(uint(offset)).
		Limit(uint(pageSize)).
		Order(orderBy).
		ToSQL()
	err := meta.ReportDB.Select(&data, query)
	if err != nil {
		return data, number, pushLog(err, helper.DBErr)
	}

	return data, number, nil
}

func agencyList(ex exp.ExpressionList, startAt, endAt int64, page, pageSize int, parentID string, agencyType int) ([]MemberListCol, int, error) {

	var data []MemberListCol
	number := 0
	ex = ex.Append(g.C("prefix").Eq(meta.Prefix))
	if agencyType != 0 {
		ex = ex.Append(g.C("agency_type").Eq(391))
	}
	if page == 1 {
		query, _, _ := dialect.From("tbl_members").Select(g.COUNT(1)).Where(ex).ToSQL()
		err := meta.MerchantDB.Get(&number, query)
		if err != nil && err != sql.ErrNoRows {
			return data, number, pushLog(err, helper.DBErr)
		}

		if number == 0 {
			return data, number, nil
		}
	}

	var members []Member
	offset := (page - 1) * pageSize
	query, _, _ := dialect.From("tbl_members").Select("uid").Where(ex).Offset(uint(offset)).
		Limit(uint(pageSize)).Order(g.L("created_at").Desc()).ToSQL()
	err := meta.MerchantDB.Select(&members, query)
	if err != nil {
		return data, number, pushLog(err, helper.DBErr)
	}

	// 补全数据
	var ids []string
	idMap := make(map[string]bool, len(members))
	for _, member := range members {
		if member.UID != parentID {
			ids = append(ids, member.UID)
		}
		idMap[member.UID] = true
	}

	// 获取统计数据
	and := g.And(
		g.C("report_type").Eq(2),
		g.C("prefix").Eq(meta.Prefix),
		g.C("report_time").Between(exp.NewRangeVal(startAt, endAt)),
	)

	if parentID == "" {
		and = and.Append(
			g.And(
				g.C("uid").Neq(g.C("parent_uid")),
				g.C("uid").In(ids),
			),
		)
	} else {

		or := g.Or(
			g.And(
				g.C("uid").In(parentID),
				g.C("parent_uid").Eq(g.C("uid")),
			),
		)

		if len(ids) > 0 {
			or = or.Append(
				g.And(
					g.C("uid").Neq(g.C("parent_uid")),
					g.C("uid").In(ids),
				),
			)
		}

		and = and.Append(or)
	}

	query, _, _ = dialect.From("tbl_report_agency").Where(and).
		Select(
			"uid",
			g.SUM("deposit_amount").As("deposit"),
			g.SUM("withdrawal_amount").As("withdraw"),
			g.SUM("valid_bet_amount").As("valid_amount"),
			g.SUM("rebate_amount").As("rebate"),
			g.SUM("company_net_amount").As("net_amount"),
		).GroupBy("uid").
		ToSQL()
	err = meta.ReportDB.Select(&data, query)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err.Error())
		return data, number, pushLog(err, helper.DBErr)
	}

	if len(ids) == len(data) {
		return data, number, nil
	}

	// 可能有会员未生成报表数据 这时需要给未生成报表的会员 赋值默认返回值
	//否则会出现total和data length 不一致的问题
	for _, v := range data {
		delete(idMap, v.UID)
	}

	for id := range idMap {
		data = append(data, MemberListCol{UID: id})
	}

	return data, number, nil
}

// MemberAgg 获取直属下级人数
func MemberAgg(ids []string, startTIme, endTime int64) (map[string]MemberAggData, error) {

	aggs := make(map[string]MemberAggData)
	var data []MemberAggData

	for _, id := range ids {
		aggs[id] = MemberAggData{UID: id}
	}

	ex := g.Ex{
		"uid":         ids,
		"report_time": g.Op{"between": exp.NewRangeVal(startTIme, endTime)},
		"report_type": 2,
		"prefix":      meta.Prefix,
	}

	query, _, _ := dialect.From("tbl_report_agency").Select(g.MAX("subordinate_count").As("mem_count"),
		g.SUM("regist_count").As("regist_count"), "uid").Where(ex).GroupBy("uid").ToSQL()
	err := meta.ReportDB.Select(&data, query)
	if err != nil && err != sql.ErrNoRows {
		return nil, pushLog(err, helper.DBErr)
	}

	if len(data) == 0 {
		return aggs, nil
	}

	for _, v := range data {
		aggs[v.UID] = v
	}

	return aggs, nil
}

func MemberRebateSelect(ids []string) (map[string]MemberRebate, error) {

	var own []MemberRebate
	query, _, _ := dialect.From("tbl_member_rebate_info").Select(colsMemberRebate...).Where(g.Ex{"uid": ids, "prefix": meta.Prefix}).ToSQL()
	err := meta.MerchantDB.Select(&own, query)
	if err != nil {
		return nil, pushLog(err, helper.DBErr)
	}

	data := make(map[string]MemberRebate)
	for _, v := range own {
		data[v.UID] = v
	}
	return data, nil
}

// 更新用户信息
func MemberUpdate(username, adminID string, param map[string]string, tagsId []string) error {

	if len(param["phone"]) > 0 &&
		(meta.Lang == "vn" && !validator.IsVietnamesePhone(param["phone"])) { //越南手机号
		return errors.New(helper.PhoneFMTErr)
	}

	mb, err := MemberFindOne(username)
	if err != nil {
		return err
	}

	if len(mb.Username) == 0 {
		return errors.New(helper.UsernameErr)
	}

	param["uid"] = mb.UID
	record := g.Record{}
	if gender, ok := param["gender"]; ok {
		if gender != "0" {
			record["gender"] = param["gender"]
		}
	}

	var (
		insertRes []schema.Enc_t
		updateRes []schema.Enc_t
	)
	if _, ok := param["realname"]; ok {

		if meta.Lang == "vn" && !validator.CheckStringVName(param["realname"]) {
			return errors.New(helper.RealNameFMTErr)
		}

		realNameHash := MurmurHash(param["realname"], 0)
		if realNameHash != mb.RealnameHash {

			param["realname_hash"] = fmt.Sprintf("%d", realNameHash)
			record["realname_hash"] = param["realname_hash"]
			recs := schema.Enc_t{
				Field: "realname",
				Value: param["realname"],
				ID:    mb.UID,
			}

			if mb.RealnameHash == 0 {
				insertRes = append(insertRes, recs)
			} else {
				updateRes = append(updateRes, recs)
			}
		}
	}

	if _, ok := param["phone"]; ok {

		phoneHash := MurmurHash(param["phone"], 0)
		if memberBindCheck(g.Ex{"phone_hash": fmt.Sprintf("%d", phoneHash)}) {
			return errors.New(helper.PhoneExist)
		}

		if phoneHash != mb.PhoneHash {

			param["phone_hash"] = fmt.Sprintf("%d", phoneHash)
			record["phone_hash"] = param["phone_hash"]
			recs := schema.Enc_t{
				Field: "phone",
				Value: param["phone"],
				ID:    mb.UID,
			}

			if mb.PhoneHash == 0 {
				insertRes = append(insertRes, recs)
			} else {
				updateRes = append(updateRes, recs)
			}
		}
	}

	if _, ok := param["email"]; ok {

		emailHash := MurmurHash(param["email"], 0)
		if memberBindCheck(g.Ex{"email_hash": fmt.Sprintf("%d", emailHash)}) {
			return errors.New(helper.EmailExist)
		}

		if emailHash != mb.EmailHash {

			param["email_hash"] = fmt.Sprintf("%d", emailHash)
			record["email_hash"] = param["email_hash"]
			recs := schema.Enc_t{
				Field: "email",
				Value: param["email"],
				ID:    mb.UID,
			}

			if mb.EmailHash == 0 {
				insertRes = append(insertRes, recs)
			} else {
				updateRes = append(updateRes, recs)
			}
		}
	}

	tags := map[string]string{}
	if len(tagsId) > 0 {

		var tagls []Tags
		// 查询标签
		query, _, _ := dialect.From("tbl_tags").Select(colsTags...).Where(g.Ex{"id": tagsId, "prefix": meta.Prefix}).ToSQL()
		err = meta.MerchantDB.Select(&tagls, query)
		if err != nil {
			return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
		}

		for _, v := range tagls {
			tags[fmt.Sprintf("%d", v.ID)] = v.Name
		}
	}

	if len(updateRes) > 0 {
		_, err = rpcUpdate(updateRes)
		if err != nil {
			return errors.New(helper.UpdateRPCErr)
		}
	}

	if len(insertRes) > 0 {
		_, err = rpcInsert(insertRes)
		if err != nil {
			return errors.New(helper.UpdateRPCErr)
		}
	}

	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	uid := param["uid"]
	ex := g.Ex{"uid": uid, "prefix": meta.Prefix}
	if len(record) > 0 {
		// 更新会员信息
		query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
		_, err = tx.Exec(query)
		if err != nil {
			return pushLog(err, helper.DBErr)
		}
	}

	// 删除该用户的所有标签
	query, _, _ := dialect.Delete("tbl_member_tags").Where(ex).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	if len(tags) > 0 {

		var data []MemberTags
		for k, v := range tags {
			tag := MemberTags{
				ID:        helper.GenId(),
				UID:       uid,
				AdminID:   adminID,
				TagID:     k,
				TagName:   v,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
				Prefix:    meta.Prefix,
			}
			data = append(data, tag)
		}

		query, _, _ = dialect.Insert("tbl_member_tags").Rows(data).ToSQL()
		_, err = tx.Exec(query)
		if err != nil {
			_ = tx.Rollback()
			return pushLog(err, helper.DBErr)
		}
	}

	err = tx.Commit()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

// 会员管理-会员列表-解除密码限制/解除短信限制
func MemberRetryReset(username string, ty uint8, pid string) error {

	if _, ok := memberUnfreeze[ty]; !ok {
		return errors.New(helper.UnfreezeTyErr)
	}

	switch ty {
	case WALLET: // 解锁钱包限制
		return memberPlatformRetryReset(username, pid)

	case PWD, SMS: // 解除密码限制/解除短信限制
		err := meta.MerchantRedis.Del(ctx, fmt.Sprintf(memberUnfreeze[ty], username)).Err()
		if err != nil {
			return pushLog(err, "redis")
		}
	}

	return nil
}

// 会员列表 用户日志写入
func MemberRemarkInsert(file, msg, adminName string, names []string, createdAt int64) error {

	// 获取所有用户的uid
	members, err := memberFindBatch(names)
	if err != nil {
		return err
	}

	if len(members) != len(names) {
		return errors.New(helper.UsernameErr)
	}

	for username, member := range members {

		//log := map[string]string{
		//	"id":         helper.GenId(),
		//	"created_at": fmt.Sprintf("%d", createdAt),
		//	"admin_name": adminName,
		//	"msg":        msg,
		//	"files":      file,
		//	"uid":        member.UID,
		//	"username":   username,
		//}
		//err = tdlog.WriteLog("member_remark_log", log)
		//if err != nil {
		//	fmt.Println("member write member_remark_log error")
		//}

		log := MemberRemarksLog{
			ID:        helper.GenId(),
			CreatedAt: createdAt,
			AdminName: adminName,
			File:      file,
			Msg:       msg,
			UID:       member.UID,
			Username:  username,
			Prefix:    meta.Prefix,
		}
		err = meta.Zlog.Post(esPrefixIndex("member_remarks_log"), log)
		if err != nil {
			fmt.Println("member write member_remarks_log error")
		}
	}

	return nil
}

// 会员管理-会员列表-数据概览
func MemberDataOverview(username, startTime, endTime string) (MemberDataOverviewData, error) {

	data := MemberDataOverviewData{}

	// 获取uid
	mb, err := MemberFindOne(username)
	if err != nil {
		return data, err
	}

	ss, err := helper.TimeToLoc(startTime, loc)
	if err != nil {
		return data, errors.New(helper.TimeTypeErr)
	}

	se, err := helper.TimeToLoc(endTime, loc)
	if err != nil {
		return data, errors.New(helper.TimeTypeErr)
	}

	// 毫秒级时间戳
	mss, err := helper.TimeToLocMs(startTime, loc)
	if err != nil {
		return data, errors.New(helper.TimeTypeErr)
	}

	mse, err := helper.TimeToLocMs(endTime, loc)
	if err != nil {
		return data, errors.New(helper.TimeTypeErr)
	}

	if mss > mse {
		return data, errors.New(helper.QueryTimeRangeErr)
	}

	// 总输赢 && 总有效投注
	filters := []elastic.Query{
		elastic.NewRangeQuery("bet_time").Gte(mss).Lte(mse),
		elastic.NewTermQuery("uid", mb.UID),
		elastic.NewTermQuery("prefix", meta.Prefix),
	}
	boolQuery := elastic.NewBoolQuery().Filter(filters...)
	esService := meta.ES.Search().
		Query(boolQuery).
		TrackTotalHits(true).
		Sort("created_at", false).
		Aggregation("net_amount", elastic.NewSumAggregation().Field("net_amount")).
		Aggregation("valid_bet_amount", elastic.NewSumAggregation().Field("valid_bet_amount"))

	res, err := esService.Index(pullPrefixIndex("tbl_game_record")).Do(ctx)
	if err != nil {
		return data, pushLog(err, "es")
	}

	winLose, _ := res.Aggregations.Sum("net_amount")
	data.NetAmount = *winLose.Value
	validBet, _ := res.Aggregations.Sum("valid_bet_amount")
	data.ValidBetAmount = *validBet.Value

	// 总存款
	ex := g.Ex{"uid": mb.UID,
		"prefix":     meta.Prefix,
		"state":      DepositSuccess,
		"confirm_at": g.Op{"between": exp.NewRangeVal(ss, se)},
	}
	query, _, _ := dialect.From("tbl_deposit").
		Select(g.COALESCE(g.SUM("amount"), 0).As("dividend")).Where(ex).ToSQL()
	err = meta.MerchantDB.Get(&data.Deposit, query)
	if err != nil {
		return data, pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	// 总提款
	wex := g.Ex{"uid": mb.UID,
		"prefix":     meta.Prefix,
		"state":      WithdrawSuccess,
		"confirm_at": g.Op{"between": exp.NewRangeVal(ss, se)},
	}
	query, _, _ = dialect.From("tbl_withdraw").
		Select(g.COALESCE(g.SUM("amount"), 0).As("withdraw")).Where(wex).ToSQL()
	err = meta.MerchantDB.Get(&data.Withdraw, query)
	if err != nil {
		return data, pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	// 总红利
	dex := g.Ex{"uid": mb.UID,
		"prefix":         meta.Prefix,
		"hand_out_state": DividendSuccess,
		"apply_at":       g.Op{"between": exp.NewRangeVal(mss, mse)},
	}
	query, _, _ = dialect.From("tbl_member_dividend").
		Select(g.COALESCE(g.SUM("amount"), 0).As("dividend")).Where(dex).ToSQL()
	err = meta.MerchantDB.Get(&data.Dividend, query)
	if err != nil {
		return data, pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	// 总返水
	//rex := g.Ex{"uid": mb.UID,
	//	"prefix":    meta.Prefix,
	//	"state":     helper.RebateReviewPass,
	//	"ration_at": g.Op{"between": exp.NewRangeVal(ss, se)},
	//}
	//query, _, _ = dialect.From("tbl_member_rebate_info").
	//	Select(g.COALESCE(g.SUM("rebate_amount"), 0).As("rebate")).Where(rex).ToSQL()
	//err = meta.MerchantDB.Get(&data.Rebate, query)
	//if err != nil {
	//	return data, pushLog(fmt.Errorf("%s,[%s]", err.Error(), query),helper.DBErr)
	//}

	return data, nil
}

func MemberUpdatePwd(username, pwd string, ty int, ctx *fasthttp.RequestCtx) error {

	mb, err := MemberFindOne(username)
	if err != nil || mb.Username == "" {
		return errors.New(helper.UsernameErr)
	}

	admin, err := AdminToken(ctx)
	if err != nil || admin["name"] == "" {
		return errors.New(helper.AccessTokenExpires)
	}

	record := g.Record{}
	if ty == 1 {
		record["withdraw_pwd"] = fmt.Sprintf("%d", MurmurHash(pwd, mb.CreatedAt))
	} else {
		record["password"] = fmt.Sprintf("%d", MurmurHash(pwd, mb.CreatedAt))
	}
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(g.Ex{"uid": mb.UID}).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

func MemberHistory(id, field string, encrypt bool) (string, error) {

	recs := schema.Res_t{
		Field: field,
		Hide:  encrypt,
		ID:    id,
	}

	resp, err := meta.Grpc.Call("History", recs)
	if err != nil {
		return "", errors.New(helper.GetRPCErr)
	}

	res, ok := resp.(string)
	if !ok {
		return "", fmt.Errorf("type assertion error")
	}

	return res, nil
}

func MemberFull(id, field string) (string, error) {

	arg := []schema.Dec_t{
		{Field: field, Hide: false, ID: id},
	}
	resp, err := rpcGet(arg)
	fmt.Println(resp, err)
	if err != nil {
		return "", fmt.Errorf("%s,%s", helper.ServerErr, err.Error())
	}

	if resp[0].Err != "" {
		return "", fmt.Errorf("%s,%s", helper.ServerErr, resp[0].Err)
	}

	return resp[0].Res, nil
}

//根据 uid数组，redis批量获取用户余额
func MemberBalance(username string) (MBBalance, error) {

	mb := MBBalance{}

	t := dialect.From("tbl_members")
	query, _, _ := t.Select(colsMemberBalance...).Where(g.Ex{"username": username, "prefix": meta.Prefix}).ToSQL()
	err := meta.MerchantDB.Get(&mb, query)
	if err != nil && err != sql.ErrNoRows {
		return mb, pushLog(err, helper.DBErr)
	}

	if err == sql.ErrNoRows {
		return mb, errors.New(helper.UsernameErr)
	}

	return mb, nil
}

//根据 uid数组，redis批量获取用户余额
func MemberBalanceBatch(uids []string) (string, error) {

	var mbs []MBBalance
	t := dialect.From("tbl_members")
	query, _, _ := t.Select(colsMemberBalance...).Where(g.Ex{"uid": uids}).ToSQL()
	err := meta.MerchantDB.Select(&mbs, query)
	if err != nil {
		return "", pushLog(err, helper.DBErr)
	}

	data, err := jettison.Marshal(mbs)
	if err != nil {
		return "", errors.New(helper.FormatErr)
	}

	return string(data), nil
}

//根据 uid数组，redis批量获取用户余额
func memMapBalanceBatch(uids []string) (map[string]MBBalance, error) {

	var mbs []MBBalance
	mbsm := map[string]MBBalance{}
	t := dialect.From("tbl_members")
	query, _, _ := t.Select(colsMemberBalance...).Where(g.Ex{"uid": uids}).ToSQL()
	err := meta.MerchantDB.Select(&mbs, query)
	if err != nil {
		return nil, pushLog(err, helper.DBErr)
	}

	for _, v := range mbs {
		mbsm[v.UID] = v
	}

	return mbsm, nil
}

// 解锁场馆钱包限制
func memberPlatformRetryReset(username, pid string) error {

	user, err := MemberFindOne(username)
	if err != nil {
		return err
	}

	param, err := memberPlatPromoInfo(user.UID, pid)
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	if param == nil {
		return errors.New(helper.PlatNoPromoApply)
	}

	// 余额解锁活动
	key := fmt.Sprintf(memberUnfreeze[WALLET], user.UID, pid)
	meta.MerchantRedis.Unlink(ctx, key)

	// 活动状态变更类型为解锁活动
	param["alter_ty"] = fmt.Sprintf("%d", PromoUnlock)
	// 解锁类型为余额解锁
	param["unlock_ty"] = fmt.Sprintf("%d", PromoUnlockAdmin)
	// 投递消息队列，异步处理会员场馆活动解锁
	_, _ = BeanPut("promo", param, 0)

	return nil
}

func PlatToMap(m MemberPlatform) map[string]interface{} {

	data := map[string]interface{}{
		"id":                      m.ID,
		"username":                m.Username,
		"password":                m.Password,
		"pid":                     m.Pid,
		"balance":                 m.Balance,
		"state":                   m.State,
		"created_at":              m.CreatedAt,
		"transfer_in":             m.TransferIn,
		"transfer_in_processing":  m.TransferInProcessing,
		"transfer_out":            m.TransferOut,
		"transfer_out_processing": m.TransferOutProcessing,
		"extend":                  m.Extend,
	}

	return data
}

func LoadMemberPlatform() error {

	var (
		total    uint = 0
		index    uint = 0
		pageSize uint = 100
	)

	t := dialect.From("tbl_member_platform")
	query, _, _ := t.Select(g.COUNT(1)).Where(g.Ex{"prefix": meta.Prefix}).ToSQL()
	err := meta.MerchantDB.Get(&total, query)
	if err != nil {
		return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	if total == 0 {
		return nil
	}

	pageMax := total / pageSize
	t = t.Select(colsMemberPlatform...)

	for index = 0; index <= pageMax; index++ {
		var data []MemberPlatform
		offset := index * pageSize
		query, _, _ = t.Offset(offset).Limit(pageSize).ToSQL()
		err = meta.MerchantDB.Select(&data, query)
		if err != nil {
			continue
		}

		pipe := meta.MerchantRedis.Pipeline()

		for _, v := range data {
			key := fmt.Sprintf("%s:%s", v.Username, v.Pid)
			pipe.Unlink(ctx, key)
			pipe.HMSet(ctx, key, PlatToMap(v))
			pipe.Persist(ctx, key)
		}

		_, _ = pipe.Exec(ctx)
		_ = pipe.Close()
	}

	return nil
}

// MemberDeviceRegLoad 加载同设备注册数
func MemberDeviceRegLoad() error {

	var total uint
	query, _, _ := dialect.From("tbl_members").Select(g.COUNT("uid")).ToSQL()
	err := meta.MerchantDB.Get(&total, query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	err = memberDeviceRegClear(total)
	if err != nil {
		return err
	}

	var (
		page     uint = 0
		pageSize uint = 1000
	)
	pageTotal := total / pageSize
	for page = 0; page <= pageTotal; page++ {

		var deviceReg []memberDeviceReg
		offset := page * pageSize
		query, _, _ := dialect.From("tbl_members").Select("reg_device", "uid").Offset(offset).Limit(pageSize).ToSQL()
		fmt.Println(query)
		err = meta.MerchantDB.Select(&deviceReg, query)
		if err != nil {
			return pushLog(err, helper.DBErr)
		}

		if len(deviceReg) == 0 {
			continue
		}

		err = memberDevice(deviceReg)
		if err != nil {
			return err
		}
	}

	return nil
}

// memberDeviceRegClear 清除所有key
func memberDeviceRegClear(total uint) error {

	var (
		page     uint = 0
		pageSize uint = 1000
	)
	pageTotal := total / pageSize
	delDevices := map[string]bool{}

	for page = 0; page <= pageTotal; page++ {

		devices := map[string]bool{}
		var deviceReg []memberDeviceReg
		offset := page * pageSize
		query, _, _ := dialect.From("tbl_members").Select("reg_device", "uid").Offset(offset).Limit(pageSize).ToSQL()
		err := meta.MerchantDB.Select(&deviceReg, query)
		if err != nil {
			return pushLog(err, helper.DBErr)
		}

		if len(deviceReg) == 0 {
			continue
		}

		for _, v := range deviceReg {
			_, ok := devices[v.RegDevice]
			_, okDel := delDevices[v.RegDevice]
			if !ok && !okDel {
				devices[v.RegDevice] = true
				delDevices[v.RegDevice] = true
			}
		}

		pipe := meta.MerchantRedis.Pipeline()

		for k := range devices {
			deviceNum := MurmurHash(k, 0)
			deviceNo := fmt.Sprintf("D:%d", deviceNum)
			pipe.Unlink(ctx, deviceNo)
		}

		_, err = pipe.Exec(ctx)
		if err != nil {
			return errors.New(helper.RedisErr)
		}

		_ = pipe.Close()
	}

	return nil
}

//memberDevice 同设备累计
func memberDevice(deviceReg []memberDeviceReg) error {

	total := len(deviceReg)
	page := 0
	pageSize := 50
	pageTotal := total / pageSize

	for page = 0; page <= pageTotal; page++ {

		index := 0
		offset := page * pageSize
		pipe := meta.MerchantRedis.Pipeline()

		for index = 0; index < 50; index++ {
			if index+offset >= total {
				break
			}

			item := deviceReg[index+offset]
			if len(item.RegDevice) < 3 {
				continue
			}

			deviceNo := fmt.Sprintf("D:%d", MurmurHash(item.RegDevice, 0))
			pipe.IncrBy(ctx, deviceNo, 1)
		}

		_, err := pipe.Exec(ctx)
		if err != nil {
			return errors.New(helper.RedisErr)
		}

		_ = pipe.Close()
	}

	return nil
}

// 检测手机号，email，是否已经被会员绑定
// 仅用来检测会员信息绑定
func memberBindCheck(ex g.Ex) bool {

	var id string

	t := dialect.From("tbl_members")
	query, _, _ := t.Select("uid").Where(ex).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&id, query)
	return err != sql.ErrNoRows
}

// 通过用户名获取用户在redis中的数据
func MemberFindOne(name string) (Member, error) {

	m := Member{}
	if name == "" {
		return m, errors.New(helper.UsernameErr)
	}

	t := dialect.From("tbl_members")
	query, _, _ := t.Select(colsMember...).Where(g.Ex{"username": name, "prefix": meta.Prefix}).ToSQL()
	err := meta.MerchantDB.Get(&m, query)
	if err != nil && err != sql.ErrNoRows {
		return m, pushLog(err, helper.DBErr)
	}

	if err == sql.ErrNoRows {
		return m, errors.New(helper.UsernameErr)
	}

	return m, nil
}

func memberFindBatch(names []string) (map[string]Member, error) {

	data := map[string]Member{}

	if len(names) == 0 {
		return data, errors.New(helper.ParamNull)
	}

	var mbs []Member
	t := dialect.From("tbl_members")
	query, _, _ := t.Select(colsMember...).Where(g.Ex{"username": names, "prefix": meta.Prefix}).ToSQL()
	err := meta.MerchantDB.Select(&mbs, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	if len(mbs) > 0 {
		for _, v := range mbs {
			if v.Username != "" {
				data[v.Username] = v
			}
		}
	}

	return data, nil
}

func memberInfoFindBatch(ids []string) (map[string]memberInfo, error) {

	if len(ids) == 0 {
		return nil, errors.New(helper.ParamNull)
	}

	var mbs []memberInfo
	query, _, _ := dialect.From("tbl_members").Select(colsMemberInfo...).Where(g.Ex{"uid": ids}).ToSQL()
	err := meta.MerchantDB.Select(&mbs, query)
	if err != nil {
		return nil, pushLog(err, helper.DBErr)
	}

	if len(mbs) == 0 {
		return nil, nil
	}

	data := make(map[string]memberInfo, len(mbs))
	for _, v := range mbs {
		data[v.UID] = v
	}

	return data, nil
}

// 获取会员指定场馆的活动锁定信息
func memberPlatPromoInfo(uid, pid string) (map[string]interface{}, error) {

	param := map[string]interface{}{}
	key := fmt.Sprintf("P:%s:%s", uid, pid)
	fields := []string{
		"pid",
		"pname",
		"cash_type",
		"apply_at",
		"water_flow",
	}
	rs, err := meta.MerchantRedis.HMGet(ctx, key, fields...).Result()
	if err != nil {
		return nil, err
	}

	if len(fields) != len(rs) {
		return nil, err
	}

	for k, v := range rs {
		if v == nil {
			return nil, err
		}
		param[fields[k]] = v
	}

	return param, nil
}

//获取会员可转账的场馆
func memberPlatformBalance(username string) []PlatBalance {

	var p []PlatBalance
	t := dialect.From("tbl_member_platform")
	ex := g.Ex{
		"username": username,
		"id": []string{
			"6798510151614082003",
			"7219886347116135962",
			"5864536520308659696",
			"1958997188942770517",
			"2658175169982643138",
			"2306868265751172637",
			"6238858173568905466",
			"1055235995899664907",
			"1371916058167324188",
			"2854120181948444476",
			"1846182857231915191",
			"2299282204811996672",
			"7591876028427516934",
			"8840968482572372234",
			"6982718883667836955",
			"1386624620395927266",
			"1794601907316741515",
			"6861705028422769039",
		},
		"prefix": meta.Prefix,
	}
	query, _, _ := t.Select(colsPlatBalance...).Where(ex).ToSQL()
	err := meta.MerchantDB.Select(&p, query)
	if err != nil {
		_ = pushLog(err, helper.DBErr)
	}

	return p
}

func MemberUpdateInfo(uid, planID string, mbRecod g.Record, mr MemberRebate) error {

	tx, err := meta.MerchantDB.Begin() // 开启事务
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	subEx := g.Ex{
		"uid":    uid,
		"prefix": meta.Prefix,
	}

	if len(mbRecod) > 0 {
		query, _, _ := dialect.Update("tbl_members").Set(&mbRecod).Where(subEx).ToSQL()
		_, err = tx.Exec(query)
		if err != nil {
			_ = tx.Rollback()
			return pushLog(err, helper.DBErr)
		}
	}

	recd := g.Record{
		"ty": mr.TY,
		"zr": mr.ZR,
		"qp": mr.QP,
		"dj": mr.DJ,
		"dz": mr.DZ,
	}
	query, _, _ := dialect.Update("tbl_member_rebate_info").Set(&recd).Where(subEx).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	if planID != "" {

		recd = g.Record{
			"plan_id": planID,
		}
		query, _, _ := dialect.Update("tbl_commission_conf").Set(&recd).Where(subEx).ToSQL()
		_, err = tx.Exec(query)
		if err != nil {
			_ = tx.Rollback()
			return pushLog(err, helper.DBErr)
		}
	}

	err = tx.Commit()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

func MemberUpdateMaintanName(uid, maintainName string) error {

	tx, err := meta.MerchantDB.Begin() // 开启事务
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	subEx := g.Ex{
		"uid": uid,
	}
	recd := g.Record{
		"maintain_name": maintainName,
	}
	query, _, _ := dialect.Update("tbl_members").Set(&recd).Where(subEx).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

func MemberMaxRebateFindOne(uid string) (MemberRebateResult_t, error) {

	data := MemberMaxRebate{}
	res := MemberRebateResult_t{}

	t := dialect.From("tbl_member_rebate_info")
	query, _, _ := t.Select(
		g.MAX("zr").As("zr"),
		g.MAX("qp").As("qp"),
		g.MAX("dz").As("dz"),
		g.MAX("dj").As("dj"),
		g.MAX("ty").As("ty"),
		g.MAX("cp").As("cp"),
	).Where(g.Ex{"parent_uid": uid, "prefix": meta.Prefix}).ToSQL()
	err := meta.MerchantDB.Get(&data, query)
	if err != nil {
		return res, pushLog(err, helper.DBErr)
	}

	res.ZR = decimal.NewFromFloat(data.ZR.Float64)
	res.QP = decimal.NewFromFloat(data.QP.Float64)
	res.TY = decimal.NewFromFloat(data.TY.Float64)
	res.DJ = decimal.NewFromFloat(data.DJ.Float64)
	res.DZ = decimal.NewFromFloat(data.DZ.Float64)
	res.CP = decimal.NewFromFloat(data.CP.Float64)

	res.ZR = res.ZR.Truncate(1)
	res.QP = res.QP.Truncate(1)
	res.TY = res.TY.Truncate(1)
	res.DJ = res.DJ.Truncate(1)
	res.DZ = res.DZ.Truncate(1)
	res.CP = res.DZ.Truncate(1)

	return res, nil
}

func MemberParentRebate(uid string) (MemberRebateResult_t, error) {

	data := MemberMaxRebate{}
	res := MemberRebateResult_t{}

	t := dialect.From("tbl_member_rebate_info")
	query, _, _ := t.Select(
		g.C("zr").As("zr"),
		g.C("qp").As("qp"),
		g.C("dz").As("dz"),
		g.C("dj").As("dj"),
		g.C("ty").As("ty"),
		g.C("cp").As("cp"),
	).Where(g.Ex{"uid": uid, "prefix": meta.Prefix}).ToSQL()
	err := meta.MerchantDB.Get(&data, query)
	if err != nil {
		return res, pushLog(err, helper.DBErr)
	}

	res.ZR = decimal.NewFromFloat(data.ZR.Float64)
	res.QP = decimal.NewFromFloat(data.QP.Float64)
	res.TY = decimal.NewFromFloat(data.TY.Float64)
	res.DJ = decimal.NewFromFloat(data.DJ.Float64)
	res.DZ = decimal.NewFromFloat(data.DZ.Float64)
	res.CP = decimal.NewFromFloat(data.CP.Float64)

	res.ZR = res.ZR.Truncate(1)
	res.QP = res.QP.Truncate(1)
	res.TY = res.TY.Truncate(1)
	res.DJ = res.DJ.Truncate(1)
	res.DZ = res.DZ.Truncate(1)
	res.CP = res.CP.Truncate(1)

	return res, nil
}

//代理管理 下级会员
func AgencyMemberList(param MemberListParam) (AgencyMemberData, error) {

	res := AgencyMemberData{}
	//查询MySQL,必须是代理的下级会员
	ex := g.Ex{}

	if param.ParentName != "" {
		ex["parent_name"] = param.ParentName
	}

	if param.State != 0 {
		ex["state"] = param.State
	}

	if param.Username != "" {
		ex["username"] = param.Username
	}

	if param.RegStart != "" && param.RegEnd != "" {

		startAt, err := helper.DayOfStart(param.RegStart, loc)
		if err != nil {
			return res, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.DayOfEnd(param.RegEnd, loc)
		if err != nil {
			return res, errors.New(helper.TimeTypeErr)
		}

		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}

	t := dialect.From("tbl_members")
	if param.Page == 1 {
		countQuery, _, _ := t.Select(g.COUNT(1)).Where(ex).ToSQL()
		err := meta.MerchantDB.Get(&res.T, countQuery)
		if err != nil {
			return res, pushLog(fmt.Errorf("%s,[%s]", err.Error(), countQuery), helper.DBErr)
		}

		if res.T == 0 {
			return res, nil
		}
	}

	var memberList []memberListShow
	offset := (param.Page - 1) * param.PageSize
	query, _, _ := t.Select(colsMemberListShow...).
		Where(ex).Offset(uint(offset)).Limit(uint(param.PageSize)).Order(g.C("created_at").Desc()).ToSQL()
	err := meta.MerchantDB.Select(&memberList, query)
	if err != nil {
		return res, pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	var (
		uids        []string
		agencyNames []string
	)

	for _, val := range memberList {
		uids = append(uids, val.UID)

		if val.ParentName != "" && val.ParentName != "root" {
			agencyNames = append(agencyNames, val.ParentName)
		}
	}

	// 用户中心钱包余额
	balanceMap, err := memMapBalanceBatch(uids)
	if err != nil {
		return res, err
	}

	rangeParam := map[string][]interface{}{}
	if param.StartAt != "" && param.EndAt != "" {

		startAt, err := helper.DayOfStart(param.StartAt, loc)
		if err != nil {
			return res, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.DayOfEnd(param.EndAt, loc)
		if err != nil {
			return res, errors.New(helper.TimeTypeErr)
		}

		rangeParam["report_time"] = []interface{}{startAt, endAt}
	}

	// 获取用户数据
	md, err := MemberSumByRange(param.StartAt, param.EndAt, uids)
	if err != nil {
		return res, err
	}

	for _, m := range memberList {

		val := memberListData{memberListShow: m}
		if md, ok := md[m.UID]; ok {
			val.NetAmount, _ = decimal.NewFromFloat(md.NetAmount).Truncate(4).Float64()
			val.Deposit, _ = decimal.NewFromFloat(md.DepositAmount).Truncate(4).Float64()
			val.Withdraw, _ = decimal.NewFromFloat(md.WithdrawAmount).Truncate(4).Float64()
			val.BetAmount, _ = decimal.NewFromFloat(md.ValidBetAmount).Truncate(4).Float64()
			val.RebateAmount, _ = decimal.NewFromFloat(md.RebateAmount).Truncate(4).Float64()
			val.DividendAmount, _ = decimal.NewFromFloat(md.DividendAmount).Truncate(4).Float64()
			val.DividendAgency, _ = decimal.NewFromFloat(md.DividendAgency).Truncate(4).Float64()
		}

		if _, o := balanceMap[m.UID]; o {
			val.Balance = balanceMap[m.UID].Balance
		}

		res.D = append(res.D, val)
	}

	return res, nil
}

func MemberSumByRange(start, end string, uids []string) (map[string]AgencyBaseSumField, error) {

	if start != "" && end != "" {

		startAt, err := helper.DayOfStart(start, loc)
		if err != nil && err != sql.ErrNoRows {
			return nil, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.DayOfEnd(end, loc)
		if err != nil && err != sql.ErrNoRows {
			return nil, errors.New(helper.TimeTypeErr)
		}

		var (
			result = map[string]AgencyBaseSumField{}
			data   []MemReport
			num    int
		)
		ex := g.Ex{
			"uid":         uids,
			"report_time": g.Op{"between": g.Range(startAt, endAt)},
			"report_type": 2,
			"data_type":   2,
		}
		query, _, _ := dialect.From("tbl_report_agency").
			Select(g.COUNT("uid").As("num")).Where(ex).Order(g.C("uid").Desc()).ToSQL()
		fmt.Println(query)
		err = meta.ReportDB.Get(&num, query)
		if num > 0 {

			query, _, _ = dialect.From("tbl_report_agency").
				Select(g.C("uid").As("uid"), g.SUM("deposit_amount").As("deposit_amount"), g.SUM("withdrawal_amount").As("withdrawal_amount"),
					g.SUM("adjust_amount").As("adjust_amount"), g.SUM("valid_bet_amount").As("valid_bet_amount"),
					g.SUM("company_net_amount").As("company_net_amount"), g.SUM("dividend_amount").As("dividend_amount"),
					g.SUM("rebate_amount").As("rebate_amount"),
				).Where(ex).GroupBy("uid").Order(g.C("uid").Desc()).ToSQL()
			fmt.Println(query)
			err = meta.ReportDB.Select(&data, query)
			if err != nil {
				return result, err
			}
			for _, v := range data {
				obj := AgencyBaseSumField{
					DepositAmount:  v.DepositAmount,
					WithdrawAmount: v.WithdrawalAmount,
					ValidBetAmount: v.ValidBetAmount,
					NetAmount:      v.CompanyNetAmount,
					DividendAmount: v.DividendAmount,
					RebateAmount:   v.RebateAmount,
					AdjustAmount:   v.AdjustAmount,
				}
				result[v.Uid] = obj
			}
			return result, nil
		}
	}
	return nil, nil

}
