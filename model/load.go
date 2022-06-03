package model

import (
	"database/sql"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"merchant/contrib/helper"
)

func LoadLink() {

	var total int

	t := dialect.From("tbl_members")
	query, _, _ := t.Select(g.COUNT("uid")).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Get(&total, query)
	if err != nil {
		fmt.Println(query, err)
		return
	}

	if total > 0 {

		p := total / LINK_PAGE
		if total%LINK_PAGE > 0 {
			p += 1
		}
		for i := 0; i < p; i++ {

			var (
				uids []string
				data []Link_t
			)
			query, _, _ = t.Where(g.Ex{}).Select("uid").Offset(uint(i * LINK_PAGE)).Limit(LINK_PAGE).ToSQL()
			//fmt.Println(query)
			err := meta.MerchantDB.Select(&uids, query)
			if err != nil {
				if err != sql.ErrNoRows {
					_ = pushLog(err, helper.DBErr)
				}
				return
			}

			ex := g.Ex{
				"uid": uids,
			}
			query, _, _ = dialect.From("tbl_member_link").Where(ex).Select(colsLink...).ToSQL()
			//fmt.Println(query)
			err = meta.MerchantDB.Select(&data, query)
			if err != nil {
				if err != sql.ErrNoRows {
					_ = pushLog(err, helper.DBErr)
				}
				return
			}

			bcs := make(map[string]map[string]Link_t)
			for _, v := range data {
				key := fmt.Sprintf("%s:lk:%s", meta.Prefix, v.UID)
				if _, ok := bcs[key]; ok {
					bcs[key]["$"+v.ID] = v
				} else {
					bcs[key] = map[string]Link_t{
						"$" + v.ID: v,
					}
				}
			}

			pipe := meta.MerchantRedis.TxPipeline()

			for k, v := range bcs {

				value, err := helper.JsonMarshal(&v)
				if err != nil {
					_ = pushLog(err, helper.FormatErr)
					return
				}

				pipe.Unlink(ctx, k)
				pipe.Do(ctx, "JSON.SET", k, ".", string(value))
				pipe.Persist(ctx, k)

				//fmt.Println(k, string(value))
			}

			_, err = pipe.Exec(ctx)
			if err != nil {
				_ = pushLog(err, helper.RedisErr)
				return
			}

			_ = pipe.Close()
		}
	}
}

func LoadMembers() {

	var total int

	t := dialect.From("tbl_members")
	query, _, _ := t.Select(g.COUNT("uid")).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Get(&total, query)
	if err != nil {
		_ = pushLog(err, helper.DBErr)
		return
	}

	if total > 0 {

		p := total / MEMBER_PAGE
		if total%MEMBER_PAGE > 0 {
			p += 1
		}
		for i := 0; i < p; i++ {

			var data []Member
			query, _, _ = t.Where(g.Ex{}).Select(colsMember...).Offset(uint(i * MEMBER_PAGE)).Limit(MEMBER_PAGE).ToSQL()
			fmt.Println(query)
			err := meta.MerchantDB.Select(&data, query)
			if err != nil {
				_ = pushLog(err, helper.DBErr)
				return
			}

			pipe := meta.MerchantRedis.TxPipeline()
			for _, v := range data {
				key := meta.Prefix + ":member:" + v.Username
				fields := []interface{}{"uid", v.UID, "username", v.Username, "password", v.Password, "birth", v.Birth, "birth_hash", v.BirthHash, "realname_hash", v.RealnameHash, "email_hash", v.EmailHash, "phone_hash", v.PhoneHash, "zalo_hash", v.ZaloHash, "prefix", v.Prefix, "tester", v.Tester, "withdraw_pwd", v.WithdrawPwd, "regip", v.Regip, "reg_device", v.RegDevice, "reg_url", v.RegUrl, "created_at", v.CreatedAt, "last_login_ip", v.LastLoginIp, "last_login_at", v.LastLoginAt, "source_id", v.SourceId, "first_deposit_at", v.FirstDepositAt, "first_deposit_amount", v.FirstDepositAmount, "first_bet_at", v.FirstBetAt, "first_bet_amount", v.FirstBetAmount, "", v.SecondDepositAt, "", v.SecondDepositAmount, "top_uid", v.TopUid, "top_name", v.TopName, "parent_uid", v.ParentUid, "parent_name", v.ParentName, "bankcard_total", v.BankcardTotal, "last_login_device", v.LastLoginDevice, "last_login_source", v.LastLoginSource, "remarks", v.Remarks, "state", v.State, "level", v.Level, "balance", v.Balance, "lock_amount", v.LockAmount, "commission", v.Commission, "group_name", v.GroupName, "agency_type", v.AgencyType, "address", v.Address, "avatar", v.Avatar}
				pipe.Del(ctx, key)
				pipe.HMSet(ctx, key, fields...)
				pipe.Persist(ctx, key)
			}
			_, err = pipe.Exec(ctx)
			if err != nil {
				_ = pushLog(err, helper.RedisErr)
				return
			}

			_ = pipe.Close()
		}
	}
}

func LoadMemberRebate() error {

	var data []MemberRebate
	query, _, _ := dialect.From("tbl_member_rebate_info").Select(colsMemberRebate...).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Select(&data, query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	pipe := meta.MerchantRedis.Pipeline()

	for _, mr := range data {

		key := fmt.Sprintf("%s:m:rebate:%s", meta.Prefix, mr.UID)
		vals := []interface{}{"zr", mr.ZR, "qp", mr.QP, "ty", mr.TY, "dj", mr.DJ, "dz", mr.DZ, "cp", mr.CP, "fc", mr.FC, "by", mr.BY, "cg_high_rebate", mr.CgHighRebate, "cg_official_rebate", mr.CgOfficialRebate}

		pipe.Unlink(ctx, key)
		pipe.HMSet(ctx, key, vals...)
		pipe.Persist(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		_ = pushLog(err, helper.DBErr)
	}

	_ = pipe.Close()

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

	var total int

	t := dialect.From("tbl_member_platform")
	query, _, _ := t.Select(g.COUNT(1)).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Get(&total, query)
	if err != nil {
		return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	if total == 0 {
		return nil
	}

	p := total / MEMBER_PAGE
	if total%MEMBER_PAGE > 0 {
		p += 1
	}
	for i := 0; i < p; i++ {

		var data []MemberPlatform
		query, _, _ = t.Select(colsMemberPlatform...).Offset(uint(i * MEMBER_PAGE)).Limit(MEMBER_PAGE).ToSQL()
		fmt.Println(query)
		err = meta.MerchantDB.Select(&data, query)
		if err != nil {
			_ = pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
			continue
		}

		pipe := meta.MerchantRedis.Pipeline()

		for _, v := range data {
			key := fmt.Sprintf("%s:m:plat:%s:%s", meta.Prefix, v.Username, v.Pid)
			pipe.Unlink(ctx, key)
			pipe.HMSet(ctx, key, PlatToMap(v))
			pipe.Persist(ctx, key)
		}

		_, err = pipe.Exec(ctx)
		if err != nil {
			_ = pushLog(err, helper.DBErr)
		}

		_ = pipe.Close()
	}

	return nil
}
