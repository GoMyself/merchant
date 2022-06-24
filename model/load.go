package model

import (
	"database/sql"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"merchant/contrib/helper"
)

func LoadLinks() {

	var data []Link_t
	query, _, _ := dialect.From("tbl_member_link").Where(g.Ex{}).Select(colsLink...).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Select(&data, query)
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

	for k, v := range bcs {

		value, err := helper.JsonMarshal(&v)
		if err != nil {
			_ = pushLog(err, helper.FormatErr)
			return
		}

		pipe := meta.MerchantRedis.TxPipeline()
		pipe.Unlink(ctx, k)
		pipe.Do(ctx, "JSON.SET", k, ".", string(value))
		pipe.Persist(ctx, k)

		_, err = pipe.Exec(ctx)
		if err != nil {
			fmt.Println(k, string(value), err)
			_ = pushLog(err, helper.RedisErr)
			return
		}

		_ = pipe.Close()
	}
}

func LoadMembers() {

	//query := "update tbl_member_rebate_info set zr = 0.4,ty=0.4,fc=0.4,qp=0.4,dj=0.3,dz=0.2,`by`=0.2,cg_high_rebate=9.8,cg_official_rebate=9.7 where uid = 4722355249852325 or parent_uid = 4722355249852325;"
	//_, _ = meta.MerchantDB.Exec(query)

	var data []Member
	query, _, _ := dialect.From("tbl_members").Where(g.Ex{}).Select(colsMember...).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Select(&data, query)
	if err != nil {
		_ = pushLog(err, helper.DBErr)
		return
	}

	total := len(data)
	p := total / LOAD_PAGE
	if total%LOAD_PAGE > 0 {
		p += 1
	}

	phoneKey := fmt.Sprintf("%s:phoneExist", meta.Prefix)
	realnameKey := fmt.Sprintf("%s:realnameExist", meta.Prefix)
	emailKey := fmt.Sprintf("%s:emailExist", meta.Prefix)
	zaloKey := fmt.Sprintf("%s:zaloExist", meta.Prefix)

	pipe := meta.MerchantRedis.TxPipeline()
	pipe.Unlink(ctx, phoneKey)
	pipe.Unlink(ctx, realnameKey)
	pipe.Unlink(ctx, emailKey)
	pipe.Unlink(ctx, zaloKey)
	_, err = pipe.Exec(ctx)
	if err != nil {
		_ = pushLog(err, helper.RedisErr)
		return
	}
	_ = pipe.Close()

	for i := 0; i < p; i++ {

		pd := data[i*LOAD_PAGE:]
		if i != p-1 {
			pd = data[i*LOAD_PAGE : (i+1)*LOAD_PAGE]
		}
		var uids []string
		pipe := meta.MerchantRedis.TxPipeline()
		for _, v := range pd {
			uids = append(uids, v.UID)
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

		d, err := grpc_t.DecryptAll(uids, false, []string{"realname", "email", "phone", "zalo"})
		if err != nil {
			_ = pushLog(err, helper.GetRPCErr)
			return
		}

		pipe1 := meta.MerchantRedis.TxPipeline()
		for _, v := range uids {
			if d[v]["realname"] != "" {
				pipe1.SAdd(ctx, realnameKey, d[v]["realname"])
			}
			if d[v]["email"] != "" {
				pipe1.SAdd(ctx, emailKey, d[v]["email"])
			}
			if d[v]["phone"] != "" {
				pipe1.SAdd(ctx, phoneKey, d[v]["phone"])
			}
			if d[v]["zalo"] != "" {
				pipe1.SAdd(ctx, zaloKey, d[v]["zalo"])
			}
		}
		_, err = pipe1.Exec(ctx)
		if err != nil {
			_ = pushLog(err, helper.RedisErr)
			return
		}
		_ = pipe1.Close()
	}
}

func LoadMemberRebates() error {

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

func LoadMemberPlatforms() error {

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

	p := total / LOAD_PAGE
	if total%LOAD_PAGE > 0 {
		p += 1
	}
	for i := 0; i < p; i++ {

		var data []MemberPlatform
		query, _, _ = t.Select(colsMemberPlatform...).Offset(uint(i * LOAD_PAGE)).Limit(LOAD_PAGE).ToSQL()
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
