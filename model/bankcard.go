package model

import (
	"database/sql"
	"errors"
	"fmt"
	"merchant/contrib/helper"
	"merchant/contrib/validator"

	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"
)

type BankcardData struct {
	D []BankCard_t `json:"d"`
	T int64        `json:"t"`
	S uint         `json:"s"`
}

type BankCard_t struct {
	ID           string `db:"id" json:"id"`
	UID          string `db:"uid" json:"uid"`
	RealName     string `json:"realname"`
	Bankcard     string `json:"bankcard_no"`
	Username     string `db:"username" json:"username"`
	BankAddress  string `db:"bank_address" json:"bank_address"`
	BankID       string `db:"bank_id" json:"bank_id"`
	BankBranch   string `db:"bank_branch_name" json:"bank_branch_name"`
	State        int    `db:"state" json:"state"`
	BankcardHash string `db:"bank_card_hash" json:"bank_card_hash"`
	CreatedAt    uint64 `db:"created_at" json:"created_at"`
	Prefix       string `db:"prefix" json:"prefix"`
}

func BankcardInsert(realName, bankcardNo string, data BankCard_t) error {

	encRes := [][]string{}

	// 获取会员真实姓名
	mb, err := MemberFindOne(data.Username)
	if err != nil {
		return err
	}

	// 判断会员银行卡数目
	if mb.BankcardTotal >= 5 {
		return errors.New(helper.MaxThreeBankCard)
	}

	//判断卡号是否存在
	err = BankCardExistRedis(bankcardNo)
	if err != nil {
		return err
	}

	memberEx := g.Ex{
		"uid": mb.UID,
	}
	memberRecord := g.Record{
		"bankcard_total": g.L("bankcard_total+1"),
	}
	// 会员未绑定真实姓名，更新第一次绑定银行卡的真实姓名到会员信息
	if mb.RealnameHash == "0" {
		// 第一次新增银行卡判断真实姓名是否为越南语
		if meta.Lang == "vn" && !validator.CheckStringVName(realName) {
			return errors.New(helper.RealNameFMTErr)
		}

		encRes = append(encRes, []string{"realname", realName})
		// 会员信息更新真实姓名和真实姓名hash
		memberRecord["realname_hash"] = fmt.Sprintf("%d", MurmurHash(realName, 0))
	}

	bankcardRecord := g.Record{
		"id":               data.ID,
		"uid":              mb.UID,
		"prefix":           meta.Prefix,
		"username":         data.Username,
		"bank_address":     data.BankAddress,
		"bank_id":          data.BankID,
		"bank_branch_name": data.BankAddress,
		"bank_card_hash":   fmt.Sprintf("%d", MurmurHash(bankcardNo, 0)),
		"created_at":       fmt.Sprintf("%d", data.CreatedAt),
		"state":            data.State,
	}

	encRes = append(encRes, []string{"bankcard" + data.ID, bankcardNo})

	// 会员银行卡插入加锁
	lkey := fmt.Sprintf("bc:%s", data.Username)
	err = Lock(lkey)
	if err != nil {
		return err
	}

	defer Unlock(lkey)

	//开启事务
	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	// 更新会员银行卡信息
	queryInsert, _, _ := dialect.Insert("tbl_member_bankcard").Rows(bankcardRecord).ToSQL()
	_, err = tx.Exec(queryInsert)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	// 更新会员信息
	queryUpdate, _, _ := dialect.Update("tbl_members").Set(memberRecord).Where(memberEx).ToSQL()
	_, err = tx.Exec(queryUpdate)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	err = grpc_t.Encrypt(mb.UID, encRes)
	if err != nil {
		return errors.New(helper.UpdateRPCErr)
	}

	key := fmt.Sprintf("%s:merchant:bankcard_exist", meta.Prefix)
	_ = meta.MerchantRedis.Do(ctx, "CF.ADD", key, bankcardNo).Err()

	BankcardUpdateCache(data.Username)
	_ = MemberUpdateCache("", data.Username)

	//fmt.Println("BankcardInsert CF.ADD = ", err)

	return nil
}

func BankCardFindOne(ex g.Ex) (BankCard_t, error) {

	data := BankCard_t{}

	ex["prefix"] = meta.Prefix

	t := dialect.From("tbl_member_bankcard")
	query, _, _ := t.Select(colsBankcard...).Where(ex).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&data, query)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("BankCardFindOne query = ", query)
		fmt.Println("BankCardFindOne err = ", err)

		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

/// 黑名单获取一个
func BlackListFindOne(ex g.Ex, table string) (BlackList_t, error) {

	data := BlackList_t{}

	ex["prefix"] = meta.Prefix

	//"tbl_blacklist"
	t := dialect.From(table)
	query, _, _ := t.Select(colsBankcard...).Where(ex).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&data, query)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("BlackListFindOne query = ", query)
		fmt.Println("BlackListFindOne err = ", err)

		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func BankcardList(page, pageSize uint, username, bankcard string) (BankcardData, error) {

	var (
		ids  []string
		data BankcardData
	)

	// h后台查询查询，必须带username或bankcard参数
	ex := g.Ex{
		"prefix": meta.Prefix,
	}
	if username != "" {
		/*
			mb, err := MemberFindOne(username)
			// 判断会员是否存在
			if err != nil {
				return data, errors.New(helper.UserNotExist)
			}
		*/
		ex["username"] = username
	}
	// 银行卡号参数可选
	if bankcard != "" {
		ex["bank_card_hash"] = fmt.Sprintf("%d", MurmurHash(bankcard, 0))
	}

	fmt.Println(ex)

	t := dialect.From("tbl_member_bankcard")
	if page == 1 {
		query, _, _ := t.Select(g.COUNT("id")).Where(ex).ToSQL()
		fmt.Println(query)
		err := meta.MerchantDB.Get(&data.T, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}

		if data.T == 0 {
			return data, nil
		}
	}

	offset := pageSize * (page - 1)
	query, _, _ := t.Select(colsBankcard...).Where(ex).Offset(offset).Limit(pageSize).Order(g.C("created_at").Desc()).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Select(&data.D, query)
	if err != nil && err != sql.ErrNoRows {
		return data, pushLog(err, helper.DBErr)
	}

	encFields := []string{"realname"}

	for _, v := range data.D {
		ids = append(ids, v.UID)
		encFields = append(encFields, "bankcard"+v.ID)
	}

	encRes, err := grpc_t.DecryptAll(ids, true, encFields)
	if err != nil {
		fmt.Println("grpc_t.Decrypt err = ", err)
		return data, errors.New(helper.GetRPCErr)
	}

	for i, v := range data.D {

		data.D[i].RealName = encRes[v.UID]["realname"]
		data.D[i].Bankcard = encRes[v.UID]["bankcard"+v.ID]
	}

	data.S = pageSize
	return data, nil
}

func BankCardExistRedis(bankcardNo string) error {

	pipe := meta.MerchantRedis.Pipeline()
	defer pipe.Close()

	key := fmt.Sprintf("%s:merchant:bankcard_exist", meta.Prefix)
	ex1Temp := pipe.Do(ctx, "CF.EXISTS", key, bankcardNo)
	key = fmt.Sprintf("%s:merchant:bankcard_blacklist", meta.Prefix)
	ex2Temp := pipe.Do(ctx, "CF.EXISTS", key, bankcardNo)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return errors.New(helper.RedisErr)
	}

	ex1 := ex1Temp.Val()
	ex2 := ex2Temp.Val()
	fmt.Printf("WARNING bankcardNo:%+v\n redis CF.EXISTS:merchant:bankcard_exist:%+v\n", bankcardNo, ex1)
	fmt.Printf("WARNING bankcardNo:%+v\n redis CF.EXISTS:merchant:bankcard_blacklist:%+v\n", bankcardNo, ex2)

	if v, ok := ex1.(int64); ok && v == 1 {
		return errors.New(helper.BankCardExistErr)
	}

	if v, ok := ex2.(int64); ok && v == 1 {
		return errors.New(helper.BankcardBan)
	}
	return nil
}

// 满足条件的银行卡数量
func BankCardExist(ex g.Ex) bool {

	var id string
	ex["prefix"] = meta.Prefix
	t := dialect.From("tbl_member_bankcard")
	query, _, _ := t.Select("uid").Where(ex).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&id, query)
	return err != sql.ErrNoRows
}

func BankcardUpdate(bid, bankID, bankAddr, bankcardNo, state string) error {

	data, err := BankCardFindOne(g.Ex{"id": bid})
	if err != nil {
		return err
	}

	if data.Username == "" {
		return errors.New(helper.BankCardNotExist)
	}

	ex := g.Ex{
		"id": bid,
	}
	record := g.Record{
		"state": state,
	}
	if bankID != "" {
		record["bank_id"] = bankID
	}

	if bankAddr != "" {
		record["bank_branch_name"] = bankAddr
		record["bank_address"] = bankAddr
	}

	if bankcardNo != "" {

		//判断卡号是否存在
		err = BankCardExistRedis(bankcardNo)
		if err != nil {
			return err
		}

		src := [][]string{
			{"bankcard" + bid, bankcardNo},
		}
		err := grpc_t.Encrypt(data.UID, src)
		if err != nil {
			fmt.Println("grpc_t.Encrypt = ", err)
			return errors.New(helper.UpdateRPCErr)
		}

		record["bank_card_hash"] = fmt.Sprintf("%d", MurmurHash(bankcardNo, 0))
	}

	query, _, _ := dialect.Update("tbl_member_bankcard").Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return errors.New(helper.DBErr)
	}

	BankcardUpdateCache(data.Username)

	return nil
}

func BankcardUpdateCache(username string) {

	var data []BankCard_t

	ex := g.Ex{
		"prefix":   meta.Prefix,
		"username": username,
		//"state":    "1",
	}

	t := dialect.From("tbl_member_bankcard")
	query, _, _ := t.Select(colsBankcard...).Where(ex).Order(g.C("created_at").Desc()).ToSQL()

	err := meta.MerchantDB.Select(&data, query)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("BankcardUpdateCache err = ", err)
		return
	}

	key := fmt.Sprintf("%s:merchant:cbc:%s", meta.Prefix, username)

	pipe := meta.MerchantRedis.Pipeline()

	pipe.Del(ctx, key)
	if len(data) > 0 {

		value, err := helper.JsonMarshal(data)
		if err == nil {
			pipe.Set(ctx, key, string(value), 0).Err()
			//fmt.Println("JSON.SET err = ", err)
		}
	}

	pipe.Exec(ctx)
	pipe.Close()
}

func BankcardDelete(fctx *fasthttp.RequestCtx, bid string) error {
	var key string
	var will_update_blacklist bool

	user, err := AdminToken(fctx)
	if err != nil {
		return errors.New(helper.AccessTokenExpires)
	}

	ex := g.Ex{
		"id": bid,
	}
	data, err := BankCardFindOne(ex)
	if err != nil {
		return err
	}

	if data.Username == "" {
		return errors.New(helper.BankCardNotExist)
	}

	// 获取会员真实信息
	mb, errm := MemberInfo(data.Username)
	if errm != nil {
		return errors.New(helper.InviteUsernameErr)
	}

	enckey := "bankcard" + bid
	// encRes:map[bankcard142491282874077388:02312645320]    银行卡hash值  和 银行卡号
	encRes, err := grpc_t.Decrypt(mb.UID, false, []string{enckey})
	if err != nil {
		return errors.New(helper.GetRPCErr)
	}

	// 插入 mysql 黑名单 数据库前 查询redis，如果存在则 不重新插入mysql 并更新 redis
	key = fmt.Sprintf("%s:merchant:bankcard_blacklist", meta.Prefix)
	cmd := meta.MerchantRedis.Do(ctx, "CF.EXISTS", key, encRes[enckey])
	err = cmd.Err()
	ex1 := cmd.Val()
	if v, ok := ex1.(int64); ok && v == 1 {
		// 此卡已经 在 黑名单 中
		will_update_blacklist = false
	} else {
		// 此卡不在黑名单，需要更新
		will_update_blacklist = true
	}

	// 删除冻结的银行卡，直接删除
	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	query, _, _ := dialect.Delete("tbl_member_bankcard").Where(g.Ex{"id": bid}).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return errors.New(helper.DBErr)
	}

	record := g.Record{
		"bankcard_total": g.L("bankcard_total-1"),
	}
	query, _, _ = dialect.Update("tbl_members").Set(record).Where(g.Ex{"uid": mb.UID}).ToSQL()
	_, err = tx.Exec(query)

	if err != nil {
		_ = tx.Rollback()
		return errors.New(helper.DBErr)
	}

	// 会员删除银行卡，加入黑名单
	bankcard_blacklist_record := g.Record{
		"id":           helper.GenId(),
		"prefix":       meta.Prefix,
		"value":        encRes[enckey],
		"remark":       "",
		"ty":           "5",
		"created_at":   fctx.Time().In(loc).Unix(),
		"created_uid":  user["id"],
		"created_name": user["name"],
	}

	if will_update_blacklist {
		/// 黑名单还没有 该银行卡，更新 mysql tbl_blacklist
		query, _, _ = dialect.Insert("tbl_blacklist").Rows(bankcard_blacklist_record).ToSQL()
		_, err = tx.Exec(query)
		if err != nil {
			_ = tx.Rollback()
			return errors.New(helper.DBErr)
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.New(helper.DBErr)
	}

	/// 更新redis
	pipe := meta.MerchantRedis.Pipeline()
	defer pipe.Close()

	key = fmt.Sprintf("%s:merchant:bankcard_exist", meta.Prefix)
	pipe.Do(ctx, "CF.DEL", key, encRes[enckey])
	if will_update_blacklist {
		key = fmt.Sprintf("%s:merchant:bankcard_blacklist", meta.Prefix)
		pipe.Do(ctx, "CF.ADD", key, encRes[enckey])
	}

	_, _ = pipe.Exec(ctx)
	//key := "cbc:" + data.Username
	//path := fmt.Sprintf(".$%s", data.ID)

	//meta.MerchantRedis.Do(ctx, "JSON.DEL", key, path).Err()

	BankcardUpdateCache(data.Username)
	return nil
}
