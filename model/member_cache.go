package model

import (
	"database/sql"
	"errors"
	"fmt"
	"merchant2/contrib/helper"

	g "github.com/doug-martin/goqu/v9"
)

type tbl_members_t struct {
	Zalo                string  `db:"zalo" json:"zalo"`         // 会员名
	RealName            string  `db:"realname" json:"realname"` // 会员名
	Phone               string  `db:"phone" json:"phone"`       // 会员名
	Email               string  `db:"email" json:"email"`       // 会员名
	Uid                 string  `db:"uid" json:"uid"`
	Username            string  `db:"username" json:"username"`                           // 会员名
	RealnameHash        string  `db:"realname_hash" json:"realname_hash"`                 // 真实姓名哈希
	EmailHash           string  `db:"email_hash" json:"email_hash"`                       // 邮件地址哈希
	PhoneHash           string  `db:"phone_hash" json:"phone_hash"`                       // 电话号码哈希
	ZaloHash            string  `db:"zalo_hash" json:"zalo_hash"`                         // zalo哈希
	Regip               string  `db:"regip" json:"regip"`                                 // 注册IP
	RegDevice           string  `db:"reg_device" json:"reg_device"`                       // 注册设备号
	CreatedAt           int64   `db:"created_at" json:"created_at"`                       // 注册时间
	LastLoginIp         string  `db:"last_login_ip" json:"last_login_ip"`                 // 最后登陆ip
	LastLoginAt         int64   `db:"last_login_at" json:"last_login_at"`                 // 最后登陆时间
	SourceId            int64   `db:"source_id" json:"source_id"`                         // 注册来源 1:pc 2:h5 3:app 4:手动
	FirstDepositAt      int64   `db:"first_deposit_at" json:"first_deposit_at"`           // 首充时间
	FirstBetAt          int64   `db:"first_bet_at" json:"first_bet_at"`                   // 首投时间
	FirstBetAmount      string  `db:"first_bet_amount" json:"first_bet_amount"`           // 首投金额
	FirstDepositAmount  string  `db:"first_deposit_amount" json:"first_deposit_amount"`   // 首充金额
	SecondDepositAt     int64   `db:"second_deposit_at" json:"second_deposit_at"`         // 二存时间
	SecondDepositAmount string  `db:"second_deposit_amount" json:"second_deposit_amount"` // 二存金额
	TopUid              string  `db:"top_uid" json:"top_uid"`                             // 总代uid
	TopName             string  `db:"top_name" json:"top_name"`                           // 总代代理
	ParentUid           string  `db:"parent_uid" json:"parent_uid"`                       // 上级uid
	ParentName          string  `db:"parent_name" json:"parent_name"`                     // 上级代理
	BankcardTotal       int64   `db:"bankcard_total" json:"bankcard_total"`               // 用户绑定银行卡的数量
	LastLoginDevice     string  `db:"last_login_device" json:"last_login_device"`         // 最后登陆设备
	LastLoginSource     int64   `db:"last_login_source" json:"last_login_source"`         // 上次登录设备来源:1=pc,2=h5,3=ios,4=andriod
	Remarks             string  `db:"remarks" json:"remarks"`                             // 备注
	Balance             float64 `db:"balance" json:"balance"`                             // 余额
	LockAmount          float64 `db:"lock_amount" json:"lock_amount"`                     // 锁定金额
	Commission          float64 `db:"commission" json:"commission"`                       // 佣金
	State               int64   `db:"state" json:"state"`                                 // 状态 1正常 2禁用
	WithdrawPwd         string  `db:"withdraw_pwd" json:"withdraw_pwd"`                   // 取款密码
	Level               int64   `db:"level" json:"level"`                                 // 用户等级
	MaintainName        string  `db:"maintain_name" json:"maintain_name"`                 // 维护人
	//LastUpDownAt        int64   `db:"last_up_down_at" json:"last_up_down_at"`             // 最后升级降级时间
	AgencyType int64  `db:"agency_type" json:"agency_type"` // 代理类型 391团队代理 393普通代理
	GroupName  string `db:"group_name" json:"group_name"`   // 团队名称 仅agency_type=391有
	Address    string `db:"address" json:"address"`         // 收货地址
}

func memberInfoCache(username string) (tbl_members_t, error) {

	m := tbl_members_t{}

	t := dialect.From("tbl_members")
	query, _, _ := t.Select("uid", "username", "realname_hash", "email_hash", "phone_hash", "zalo_hash", "regip", "reg_device", "created_at", "last_login_ip", "last_login_at", "source_id", "first_deposit_at", "first_bet_at", "first_bet_amount", "first_deposit_amount", "second_deposit_at", "second_deposit_amount", "top_uid", "top_name", "parent_uid", "parent_name", "bankcard_total", "last_login_device", "last_login_source", "remarks", "balance", "lock_amount", "commission", "state", "withdraw_pwd", "level", "maintain_name", "agency_type", "group_name", "address").Where(g.Ex{"username": username, "prefix": meta.Prefix}).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&m, query)
	if err != nil && err != sql.ErrNoRows {

		fmt.Println("memberInfoCache err = ", err)
		fmt.Println("memberInfoCache query = ", query)
		return m, pushLog(err, helper.DBErr)
	}

	if err == sql.ErrNoRows {
		return m, errors.New(helper.UsernameErr)
	}

	/*
		pipe := meta.MerchantRedis.TxPipeline()

		exist := pipe.Exists(ctx, username)
		rs := pipe.HMGet(ctx, username, "uid", "username", "realname_hash", "email_hash", "phone_hash", "zalo_hash", "regip", "reg_device", "created_at", "last_login_ip", "last_login_at", "source_id", "first_deposit_at", "first_bet_at", "first_bet_amount", "first_deposit_amount", "second_deposit_at", "second_deposit_amount", "top_uid", "top_name", "parent_uid", "parent_name", "bankcard_total", "last_login_device", "last_login_source", "remarks", "balance", "lock_amount", "commission", "state", "withdraw_pwd", "level", "maintain_name", "last_up_down_at", "agency_type", "group_name", "address")

		_, err := pipe.Exec(ctx)
		pipe.Close()
		if err != nil {
			fmt.Println("memberInfoCache pipe.Exec err = ", err.Error())
			return m, errors.New(helper.RedisErr)
		}

		num, err := exist.Result()
		if num == 0 {
			fmt.Println("memberInfoCache exist.Result err = ", err.Error())
			return m, errors.New(helper.UsernameErr)
		}

		if err = rs.Scan(&m); err != nil {
			fmt.Println("memberInfoCache rs.Scan err = ", err.Error())
			return m, errors.New(helper.RedisErr)
		}
	*/
	return m, nil
}

func MemberInfo(username string) (tbl_members_t, error) {

	res, err := memberInfoCache(username)
	if err != nil {
		return res, err
	}

	encRes := []string{}
	if res.RealnameHash != "0" {

		encRes = append(encRes, "realname")
	}
	if res.PhoneHash != "0" {

		encRes = append(encRes, "phone")
	}
	if res.EmailHash != "0" {

		encRes = append(encRes, "email")
	}
	if res.ZaloHash != "0" {

		encRes = append(encRes, "zalo")
	}

	res.Zalo = ""
	res.RealName = ""
	res.Phone = ""
	res.Email = ""

	if len(encRes) > 0 {
		recs, err := grpc_t.Decrypt(res.Uid, true, encRes)
		if err != nil {

			//fmt.Println("MemberInfo res.MemberInfos.UID = ", res.MemberInfos.UID)
			//fmt.Println("MemberInfo grpc_t.Decrypt err = ", err.Error())
			return res, errors.New(helper.UpdateRPCErr)
		}

		res.Zalo = recs["zalo"]
		res.RealName = recs["realname"]
		res.Phone = recs["phone"]
		res.Email = recs["email"]
	}

	return res, nil
}
