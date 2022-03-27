package model

import (
	"database/sql"
	"errors"
	"fmt"
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"strconv"
	"time"

	"bitbucket.org/nwf2013/schema"
	g "github.com/doug-martin/goqu/v9"
)

func BankcardInsert(realName, bankcard string, data BankCard) error {

	var res []schema.Enc_t
	// 获取会员真实姓名
	mb, err := MemberFindOne(data.Username)
	if err != nil {
		return err
	}

	// 判断会员银行卡数目
	if mb.BankcardTotal >= 3 {
		return errors.New(helper.MaxThreeBankCard)
	}

	//判断卡号是否存在
	bankcardHash := MurmurHash(bankcard, 0)
	idx := bankcardHash % 10
	key := fmt.Sprintf("bl:bc%d", idx)
	ok, err := meta.MerchantRedis.SIsMember(ctx, key, bankcardHash).Result()
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	if ok {
		return errors.New(helper.BankcardBan)
	}

	//判断卡号是否存在
	cardNoHash := fmt.Sprintf("%d", bankcardHash)
	ex := g.Ex{
		"bank_card_hash": cardNoHash,
		"prefix":         meta.Prefix,
	}
	bcd, err := BankCardFindOne(ex)
	if err != nil {
		return err
	}

	var isDelete bool
	// 存在记录
	if bcd.UID != "" {

		// 已存在正常银行卡/冻结银行卡
		if bcd.State == 1 || bcd.State == 3 {
			return errors.New(helper.BankCardExistErr)
		}

		// 已删除的银行卡
		if bcd.State == 2 {
			isDelete = true
		}
	}

	ex = g.Ex{
		"uid": mb.UID,
	}
	record := g.Record{
		"bankcard_total": g.L("bankcard_total+1"),
	}
	// 会员未绑定真实姓名，更新第一次绑定银行卡的真实姓名到会员信息
	if mb.RealnameHash == 0 {
		// 第一次新增银行卡判断真实姓名是否为越南语
		if meta.Lang == "vn" && !validator.CheckStringVName(realName) {
			return errors.New(helper.RealNameFMTErr)
		}

		recs := schema.Enc_t{
			Field: "realname",
			Value: realName,
			ID:    mb.UID,
		}

		res = append(res, recs)
		realNameHash := fmt.Sprintf("%d", MurmurHash(realName, 0))
		// 会员信息更新真实姓名和真实姓名hash
		record["realname_hash"] = realNameHash
	}

	bc := g.Record{
		"id":               data.ID,
		"uid":              mb.UID,
		"prefix":           meta.Prefix,
		"username":         data.Username,
		"bank_address":     data.BankAddress,
		"bank_id":          data.BankID,
		"bank_branch_name": data.BankAddress,
		"bank_card_hash":   cardNoHash,
		"created_at":       fmt.Sprintf("%d", data.CreatedAt),
	}

	recs := schema.Enc_t{
		Field: "bankcard",
		Value: bankcard,
		ID:    data.ID,
	}

	res = append(res, recs)
	_, err = rpcInsert(res)
	if err != nil {
		return errors.New(helper.UpdateRPCErr)
	}

	// 会员银行卡插入加锁
	key = fmt.Sprintf("bc:%s", data.Username)
	err = Lock(key)
	if err != nil {
		return err
	}

	defer Unlock(key)

	//开启事务
	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	if isDelete {
		query := fmt.Sprintf("delete from tbl_member_bankcard where bank_card_hash = %s and state = 2", cardNoHash)
		_, err = tx.Exec(query)
		if err != nil {
			_ = tx.Rollback()
			return pushLog(err, helper.DBErr)
		}
	}

	// 更新会员银行卡信息
	queryInsert, _, _ := dialect.Insert("tbl_member_bankcard").Rows(bc).ToSQL()
	_, err = tx.Exec(queryInsert)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	// 更新会员信息
	queryUpdate, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = tx.Exec(queryUpdate)
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

func BankCardFindOne(ex g.Ex) (BankCard, error) {

	data := BankCard{}
	ex["prefix"] = meta.Prefix
	t := dialect.From("tbl_member_bankcard")
	query, _, _ := t.Select(colsBankcard...).Where(ex).Order(g.C("state").Asc()).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&data, query)
	if err != nil && err != sql.ErrNoRows {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func BankcardList(username, bankcard string) ([]BankcardData, error) {

	var (
		res  []schema.Dec_t
		data []BankcardData
	)

	// h后台查询查询，必须带username或bankcard参数
	ex := g.Ex{
		"state":  []int{1, 3},
		"prefix": meta.Prefix,
	}
	if username != "" {
		// 判断会员是否存在
		if !MemberExist(username) {
			return data, errors.New(helper.UsernameErr)
		}

		ex["username"] = username
	}
	// 银行卡号参数可选
	if bankcard != "" {
		ex["bank_card_hash"] = fmt.Sprintf("%d", MurmurHash(bankcard, 0))
	}
	var cardList []BankCard
	t := dialect.From("tbl_member_bankcard")
	query, _, _ := t.Select(colsBankcard...).Where(ex).Order(g.C("created_at").Desc()).ToSQL()
	err := meta.MerchantDB.Select(&cardList, query)
	if err != nil && err != sql.ErrNoRows {
		return data, pushLog(err, helper.DBErr)
	}

	length := len(cardList)
	if length == 0 {
		return data, nil
	}

	for _, v := range cardList {
		recs := schema.Dec_t{
			Field: "bankcard",
			Hide:  true,
			ID:    v.ID,
		}
		res = append(res, recs)
	}

	recs := schema.Dec_t{
		Field: "realname",
		Hide:  true,
		ID:    cardList[0].UID,
	}
	res = append(res, recs)
	record, err := rpcGet(res)
	if err != nil {
		return data, errors.New(helper.GetRPCErr)
	}

	rpcLen := len(record)
	for k, v := range cardList {

		card := ""
		if rpcLen > k && record[k].Err == "" {
			card = record[k].Res
		}

		realName := ""
		if rpcLen > length && record[length].Err == "" {
			realName = record[length].Res
		}

		val := BankcardData{
			BankCard: v,
			RealName: realName,
			Bankcard: card,
		}
		data = append(data, val)
	}

	return data, nil
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

func BankcardUpdate(bid, bankID, bankAddr, bankcard string) error {

	data, err := BankCardFindOne(g.Ex{"id": bid})
	if err != nil {
		return err
	}

	if data.Username == "" {
		return errors.New(helper.BankCardNotExist)
	}

	// 冻结删除的银行卡不允许编辑
	if data.State == 2 || data.State == 3 {
		return errors.New(helper.OperateFailed)
	}

	ex := g.Ex{
		"id":     bid,
		"prefix": meta.Prefix,
	}
	record := g.Record{}
	if bankID != "" {
		record["bank_id"] = bankID
	}

	if bankAddr != "" {
		record["bank_branch_name"] = bankAddr
		record["bank_address"] = bankAddr
	}

	if bankcard != "" {

		bankcardHash := fmt.Sprintf("%d", MurmurHash(bankcard, 0))
		if BankCardExist(g.Ex{"bank_card_hash": bankcardHash}) {
			return errors.New(helper.BankCardExistErr)
		}

		record["bank_card_hash"] = bankcardHash
		var res []schema.Enc_t
		recs := schema.Enc_t{
			Field: "bankcard",
			Value: bankcard,
			ID:    bid,
		}
		res = append(res, recs)
		_, err = rpcUpdate(res)
		if err != nil {
			return errors.New(helper.UpdateRPCErr)
		}
	}

	query, _, _ := dialect.Update("tbl_member_bankcard").Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

func BankcardDelete(bid string, adminUID, adminName string) error {

	ex := g.Ex{
		"id":    bid,
		"state": []int{1, 3},
	}
	data, err := BankCardFindOne(ex)
	if err != nil {
		return err
	}

	if data.Username == "" {
		return errors.New(helper.BankCardNotExist)
	}

	// 获取会员真实姓名
	mb, err := MemberFindOne(data.Username)
	if err != nil {
		return err
	}

	if mb.UID == "" {
		return errors.New(helper.UsernameErr)
	}

	// 删除冻结的银行卡，直接删除
	if data.State == 3 {

		tx, err := meta.MerchantDB.Begin()
		if err != nil {
			return pushLog(err, helper.DBErr)
		}

		query, _, _ := dialect.Update("tbl_member_bankcard").Set(g.Record{"state": 2}).Where(g.Ex{"id": bid, "prefix": meta.Prefix}).ToSQL()
		_, err = tx.Exec(query)
		if err != nil {
			_ = tx.Rollback()
			return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
		}

		record := g.Record{
			"bankcard_total": g.L("bankcard_total-1"),
		}
		query, _, _ = dialect.Update("tbl_members").Set(record).Where(g.Ex{"uid": mb.UID}).ToSQL()
		_, err = tx.Exec(query)
		if err != nil {
			_ = tx.Rollback()
			return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
		}

		err = tx.Commit()
		if err != nil {
			return pushLog(err, helper.DBErr)
		}

		return nil
	}

	hash, _ := strconv.ParseUint(data.BankcardHash, 10, 64)
	//开启事务
	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	t := dialect.Update("tbl_member_bankcard")
	query, _, _ := t.Set(g.Record{"state": 2}).Where(g.Ex{"id": bid, "prefix": meta.Prefix}).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	query, _, _ = dialect.Update("tbl_members").
		Set(g.Record{"bankcard_total": g.L("bankcard_total-1")}).Where(g.Ex{"username": data.Username, "prefix": meta.Prefix}).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	record := g.Record{
		"id":           bid,
		"ty":           TyBankcard,
		"value":        data.BankcardHash,
		"remark":       "delete",
		"accounts":     data.Username,
		"created_at":   time.Now().Unix(),
		"created_uid":  adminUID,
		"created_name": adminName,
	}
	query, _, _ = dialect.Insert("tbl_blacklist").Rows(&record).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	// 会员删除银行卡，加入黑名单
	idx := hash % 10
	key := fmt.Sprintf("bl:bc%d", idx)
	// 加入values set
	_, err = meta.MerchantRedis.SAdd(ctx, key, data.BankcardHash).Result()
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	return nil
}
