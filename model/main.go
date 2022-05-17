package model

import (
	"context"
	"fmt"
	"merchant2/contrib/helper"
	"merchant2/contrib/tracerr"
	"strings"

	"github.com/hprose/hprose-golang/v3/rpc/core"
	rpchttp "github.com/hprose/hprose-golang/v3/rpc/http"
	. "github.com/hprose/hprose-golang/v3/rpc/http/fasthttp"
	"github.com/nats-io/nats.go"
	"github.com/shopspring/decimal"
	cpool "github.com/silenceper/pool"

	"time"

	"errors"

	g "github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/elastic/v7"
	"github.com/spaolacci/murmur3"
)

type log_t struct {
	ID      string `json:"id" msg:"id"`
	Project string `json:"project" msg:"project"`
	Flags   string `json:"flags" msg:"flags"`
	Fn      string `json:"fn" msg:"fn"`
	File    string `json:"file" msg:"file"`
	Content string `json:"content" msg:"content"`
}

type VenueRebateScale struct {
	ZR decimal.Decimal
	QP decimal.Decimal
	TY decimal.Decimal
	DZ decimal.Decimal
	DJ decimal.Decimal
	CP decimal.Decimal
	FC decimal.Decimal
}

var grpc_t struct {
	View       func(uid, field string, hide bool) ([]string, error)
	Encrypt    func(uid string, data [][]string) error
	Decrypt    func(uid string, hide bool, field []string) (map[string]string, error)
	DecryptAll func(uids []string, hide bool, field []string) (map[string]map[string]string, error)
}

type MetaTable struct {
	VenueRebate    VenueRebateScale
	MerchantRedis  *redis.Client
	MerchantTD     *sqlx.DB
	MerchantDB     *sqlx.DB
	ReportDB       *sqlx.DB
	BetDB          *sqlx.DB
	PromoteConfig  map[string]map[string]interface{}
	BeanPool       cpool.Pool
	BeanBetPool    cpool.Pool
	ES             *elastic.Client
	AccessEs       *elastic.Client
	NatsConn       *nats.Conn
	AutoCommission bool
	Prefix         string
	EsPrefix       string
	PullPrefix     string
	Lang           string
	GcsDoamin      string
	Program        string
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
	colsBankcard             = helper.EnumFields(BankCard{})
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
	colsCommPlanDetail       = helper.EnumFields(CommissionDetail{})
	colsTblCommissions       = helper.EnumFields(Commissions{})
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
	dividendFields           = []string{"id", "prefix", "uid", "parent_uid", "parent_name", "wallet", "batch", "batch_id", "level", "ty", "agency_type", "water_limit", "platform_id", "username", "amount", "hand_out_amount", "water_flow", "notify", "state", "hand_out_state", "remark", "review_remark", "apply_at", "apply_uid", "apply_name", "review_at", "review_uid", "review_name"}
	adjustFields             = []string{"id", "prefix", "uid", "parent_uid", "parent_name", "username", "agent_id", "agency_type", "amount", "adjust_type", "adjust_mode", "is_turnover", "turnover_multi", "pid", "apply_remark", "review_remark", "agent_name", "state", "hand_out_state", "images", "level", "svip", "is_agent", "apply_at", "apply_uid", "apply_name", "review_at", "review_uid", "review_name"}
	depositFields            = []string{"id", "parent_name", "prefix", "oid", "channel_id", "finance_type", "uid", "level", "parent_uid", "agency_type", "username", "cid", "pid", "amount", "state", "automatic", "created_at", "created_uid", "created_name", "confirm_at", "confirm_uid", "confirm_name", "review_remark"}
	withdrawFields           = []string{"id", "parent_name", "prefix", "bid", "flag", "finance_type", "oid", "uid", "level", "parent_uid", "agency_type", "username", "pid", "amount", "state", "automatic", "created_at", "confirm_at", "confirm_uid", "review_remark", "withdraw_at", "confirm_name", "withdraw_uid", "withdraw_name", "withdraw_remark", "bank_name", "card_name", "card_no"}
	loginLogFields           = []string{"username", "ip", "ips", "device", "device_no", "date", "serial", "agency", "parents"}
	smsLogFields             = []string{"username", "ip", "create_at", "code", "phone", "phone_hash"}
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
	//client.Use(log.Plugin)

	client.UseService(&grpc_t)

	meta.VenueRebate = VenueRebateScale{
		ZR: decimal.NewFromFloat(1.0).Truncate(1),
		QP: decimal.NewFromFloat(1.2).Truncate(1),
		TY: decimal.NewFromFloat(1.5).Truncate(1),
		DZ: decimal.NewFromFloat(1.2).Truncate(1),
		DJ: decimal.NewFromFloat(1.1).Truncate(1),
		CP: decimal.NewFromFloat(1.1).Truncate(1),
		FC: decimal.NewFromFloat(1.5).Truncate(1),
	}

	//_, _ = meta.NatsConn.Subscribe(meta.Prefix+":merchant_notify", func(m *nats.Msg) {
	//	fmt.Printf("Nats received a message: %s\n", string(m.Data))
	//})
}

func Load() {

	AppUpgradeLoadCache()
	//_ = GameToMinio()
	//_ = PlatToMinio()
	_ = PrivRefresh()
	_ = GroupRefresh()
	_ = LoadMemberPlatform()
	_ = BlacklistLoadCache()
	_ = BannersLoadCache()
	_ = TreeLoadToRedis()
}

func MurmurHash(str string, seed uint32) uint64 {

	h64 := murmur3.New64WithSeed(seed)
	h64.Write([]byte(str))
	v := h64.Sum64()
	h64.Reset()

	return v
}

func pushLog(err error, code string) error {

	err = tracerr.Wrap(err)
	ts := time.Now()
	fields := g.Record{
		"id":       helper.GenId(),
		"content":  err.Error(),
		"project":  meta.Program,
		"flags":    code,
		"filename": tracerr.SprintSource(err, 2, 2),
		"ts":       ts.In(loc).UnixMilli(),
	}

	query, _, _ := dialect.Insert("goerror").Rows(&fields).ToSQL()
	//fmt.Println(query)
	_, err1 := meta.MerchantTD.Exec(query)
	if err1 != nil {
		fmt.Println("insert SMS = ", err1.Error())
	}

	note := fmt.Sprintf("Hệ thống lỗi %s", fields["id"])
	return errors.New(note)
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
