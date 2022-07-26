package model

import (
	"context"
	"fmt"
	"merchant/contrib/helper"
	"runtime"
	"strings"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/beanstalkd/go-beanstalk"
	"github.com/hprose/hprose-golang/v3/rpc/core"
	rpchttp "github.com/hprose/hprose-golang/v3/rpc/http"
	. "github.com/hprose/hprose-golang/v3/rpc/http/fasthttp"
	"github.com/nats-io/nats.go"
	"github.com/shopspring/decimal"

	"time"

	g "github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/spaolacci/murmur3"
)

var grpc_t struct {
	View       func(uid, field string, hide bool) ([]string, error)
	Encrypt    func(uid string, data [][]string) error
	Decrypt    func(uid string, hide bool, field []string) (map[string]string, error)
	DecryptAll func(uids []string, hide bool, field []string) (map[string]map[string]string, error)
}

type MetaTable struct {
	VenueRebate   MemberRebateResult_t
	MerchantRedis *redis.ClusterClient
	MerchantPika  *redis.Client
	MerchantTD    *sqlx.DB
	MerchantDB    *sqlx.DB
	MerchantBean  *beanstalk.Conn
	MerchantMQ    rocketmq.Producer
	ReportDB      *sqlx.DB
	BetDB         *sqlx.DB
	TiDB          *sqlx.DB
	PromoteConfig map[string]map[string]interface{}
	NatsConn      *nats.Conn
	IsDev         bool
	IndexUrl      string
	Prefix        string
	EsPrefix      string
	PullPrefix    string
	Lang          string
	GcsDoamin     string
	Program       string
}

var (
	meta                     *MetaTable
	loc                      *time.Location
	ctx                      = context.Background()
	zero                     = decimal.NewFromInt(0)
	dialect                  = g.Dialect("mysql")
	colsGroup                = helper.EnumFields(Group{})
	colsAdmin                = helper.EnumFields(Admin{})
	colsMember               = helper.EnumFields(Member{})
	colsMemberLevel          = helper.EnumFields(MemberLevel{})
	colsBankcard             = helper.EnumFields(BankCard_t{})
	colsPlatform             = helper.EnumFields(Platform{})
	colsPlatJson             = helper.EnumFields(platJson{})
	colsMemberBalance        = helper.EnumFields(MBBalance{})
	colsPlatBalance          = helper.EnumFields(PlatBalance{})
	colsTags                 = helper.EnumFields(Tags{})
	colsMemberTags           = helper.EnumFields(MemberTags{})
	colsMemberPlatform       = helper.EnumFields(MemberPlatform{})
	colsMemberAdjust         = helper.EnumFields(MemberAdjust{})
	colsDividend             = helper.EnumFields(MemberDividend{})
	colsAppUpgrade           = helper.EnumFields(AppUpgrade{})
	colsBanner               = helper.EnumFields(Banner{})
	colsBlacklist            = helper.EnumFields(Blacklist{})
	colsGameList             = helper.EnumFields(GameLists{})
	colsShowGame             = helper.EnumFields(showGameJson{})
	colsNotice               = helper.EnumFields(Notice{})
	colsTransfer             = helper.EnumFields(Transfer{})
	colsTransaction          = helper.EnumFields(Transaction{})
	colsCommPlan             = helper.EnumFields(CommissionPlan{})
	colsCommissionTransfer   = helper.EnumFields(CommissionTransfer{})
	colsMemberRebate         = helper.EnumFields(MemberRebate{})
	colsMemberInfo           = helper.EnumFields(memberInfo{})
	colLevelRecord           = helper.EnumFields(MemberLevelRecord{})
	colsMemberListShow       = helper.EnumFields(memberListShow{})
	colsAgencyTransfer       = helper.EnumFields(AgencyTransfer{})
	colsAgencyTransferRecord = helper.EnumFields(AgencyTransferRecord{})
	colsMessage              = helper.EnumFields(Message{})
	colsPromoRecord          = helper.EnumFields(PromoRecord{})
	colWithdrawRecord        = helper.EnumFields(WithdrawRecord{})
	colsPromoData            = helper.EnumFields(PromoData{})
	colsPromoInspection      = helper.EnumFields(PromoInspection{})
	colsLink                 = helper.EnumFields(Link_t{})
	colsMessageTD            = helper.EnumFields(MessageTD{})
	colsDeposit              = helper.EnumFields(Deposit{})
	colsWithdraw             = helper.EnumFields(Withdraw{})
	colsBankcardLog          = helper.EnumFields(BankcardLog{})
	colsMemberRemarksLog     = helper.EnumFields(MemberRemarksLog{})
	colRebateReport          = helper.EnumFields(RebateReportItem{})
	colsGameRecord           = helper.EnumFields(GameRecord{})
	dividendFields           = []string{"id", "uid", "prefix", "ty", "username", "top_uid", "top_name", "parent_uid", "parent_name", "amount", "review_at", "review_uid", "review_name", "water_multiple", "water_flow"}
	adjustFields             = []string{"id", "prefix", "uid", "parent_uid", "parent_name", "username", "agent_id", "agency_type", "amount", "adjust_type", "adjust_mode", "is_turnover", "turnover_multi", "pid", "apply_remark", "review_remark", "agent_name", "state", "hand_out_state", "images", "level", "svip", "is_agent", "apply_at", "apply_uid", "apply_name", "review_at", "review_uid", "review_name"}
	depositFields            = []string{"id", "parent_name", "prefix", "oid", "channel_id", "finance_type", "uid", "level", "parent_uid", "agency_type", "username", "cid", "pid", "amount", "state", "automatic", "created_at", "created_uid", "created_name", "confirm_at", "confirm_uid", "confirm_name", "review_remark"}
	withdrawFields           = []string{"id", "parent_name", "prefix", "bid", "flag", "finance_type", "oid", "uid", "level", "parent_uid", "agency_type", "username", "pid", "amount", "state", "automatic", "created_at", "confirm_at", "confirm_uid", "review_remark", "withdraw_at", "confirm_name", "withdraw_uid", "withdraw_name", "withdraw_remark", "bank_name", "card_name", "card_no"}
)

func Constructor(mt *MetaTable, rpc string) {

	meta = mt
	if meta.Lang == "cn" {
		loc, _ = time.LoadLocation("Asia/Shanghai")
	} else if meta.Lang == "vn" || meta.Lang == "th" {
		loc, _ = time.LoadLocation("Asia/Bangkok")
	}

	rpchttp.RegisterHandler()
	RegisterTransport()

	client := core.NewClient(rpc)
	client.UseService(&grpc_t)

	meta.VenueRebate = MemberRebateResult_t{
		ZR:               decimal.NewFromFloat(1.0).Truncate(1),
		QP:               decimal.NewFromFloat(1.2).Truncate(1),
		TY:               decimal.NewFromFloat(1.5).Truncate(1),
		DZ:               decimal.NewFromFloat(1.2).Truncate(1),
		DJ:               decimal.NewFromFloat(1.1).Truncate(1),
		CP:               decimal.NewFromFloat(1.1).Truncate(1),
		FC:               decimal.NewFromFloat(1.5).Truncate(1),
		BY:               decimal.NewFromFloat(1.2).Truncate(1),
		CGHighRebate:     decimal.NewFromFloat(10.00).Truncate(2),
		CGOfficialRebate: decimal.NewFromFloat(10.00).Truncate(2),
	}
}

func Load() {

	LoadAppUpgrades()
	LoadLinks()
	LoadMembers()
	LoadMemberLevels()

	//_ = LoadBankcards()
	_ = LoadPrivs()
	_ = LoadGroups()
	_ = LoadMemberPlatforms()
	_ = LoadBlacklists(0)
	_ = LoadBanners()
	_ = LoadMemberRebates()
	_ = LoadTrees()
	_ = LoadPlatforms()
	_ = LoadSMSChannels()
	_ = LoadGameLists()
}

func MurmurHash(str string, seed uint32) uint64 {

	h64 := murmur3.New64WithSeed(seed)
	h64.Write([]byte(str))
	v := h64.Sum64()
	h64.Reset()

	return v
}

func pushLog(err error, code string) error {

	_, file, line, _ := runtime.Caller(1)
	paths := strings.Split(file, "/")
	l := len(paths)
	if l > 2 {
		file = paths[l-2] + "/" + paths[l-1]
	}
	path := fmt.Sprintf("%s:%d", file, line)

	ts := time.Now()
	id := helper.GenId()

	fields := g.Record{
		"id":       id,
		"content":  err.Error(),
		"project":  meta.Program,
		"flags":    code,
		"filename": path,
		"ts":       ts.In(loc).UnixMicro(),
	}
	query, _, _ := dialect.Insert("goerror").Rows(&fields).ToSQL()
	fmt.Println("insert SMS = sql ", query)
	_, err1 := meta.MerchantTD.Exec(query)
	if err1 != nil {
		fmt.Println("insert SMS = sql ", query)
		fmt.Println("insert SMS = error ", err1.Error())
	}

	return fmt.Errorf("hệ thống lỗi %s", id)
}

func PushMerchantNotify(format, applyName, username, amount string) error {

	msg := fmt.Sprintf(format, applyName, username, amount, applyName, username, amount, applyName, username, amount)
	msg = strings.TrimSpace(msg)
	err := meta.NatsConn.Publish(meta.Prefix+":merchant_notify", []byte(msg))
	fmt.Printf("Nats send a message: %s\n", msg)
	if err != nil {
		fmt.Printf("Nats send message error: %s\n", err.Error())
		return err
	}

	_ = meta.NatsConn.Flush()
	return nil
}

func esPrefixIndex(index string) string {
	return meta.EsPrefix + index
}

func pullPrefixIndex(index string) string {
	return meta.PullPrefix + index
}

func Close() {
	_ = meta.ReportDB.Close()
	_ = meta.MerchantDB.Close()
	_ = meta.MerchantRedis.Close()
}
