package model

import (
	"database/sql"
	"errors"
	"fmt"
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"time"

	g "github.com/doug-martin/goqu/v9"
)

func BankcardInsert(realName, bankcardNo string, data BankCard) error {

	encRes := [][]string{}

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
	err = BankCardExistRedis(bankcardNo)
	if err != nil {
		return err
	}
	member_ex := g.Ex{
		"uid": mb.UID,
	}
	member_record := g.Record{
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
		member_record["realname_hash"] = fmt.Sprintf("%d", MurmurHash(realName, 0))
	}

	bankcard_record := g.Record{
		"id":               data.ID,
		"uid":              mb.UID,
		"prefix":           meta.Prefix,
		"username":         data.Username,
		"bank_address":     data.BankAddress,
		"bank_id":          data.BankID,
		"bank_branch_name": data.BankAddress,
		"bank_card_hash":   fmt.Sprintf("%d", MurmurHash(bankcardNo, 0)),
		"created_at":       fmt.Sprintf("%d", data.CreatedAt),
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
	queryInsert, _, _ := dialect.Insert("tbl_member_bankcard").Rows(bankcard_record).ToSQL()
	_, err = tx.Exec(queryInsert)
	if err != nil {
		_ = tx.Rollback()
		fmt.Println("queryInsert = ", queryInsert)
		return pushLog(err, helper.DBErr)
	}

	// 更新会员信息
	queryUpdate, _, _ := dialect.Update("tbl_members").Set(member_record).Where(member_ex).ToSQL()
	_, err = tx.Exec(queryUpdate)
	if err != nil {
		_ = tx.Rollback()
		fmt.Println("queryUpdate = ", queryUpdate)
		return pushLog(err, helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	err = grpc_t.Encrypt(mb.UID, encRes)
	if err != nil {
		fmt.Println("grpc_t.Encrypt = ", err)
		return errors.New(helper.UpdateRPCErr)
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
		uid  string
		ids  []string
		data []BankcardData
	)

	// h后台查询查询，必须带username或bankcard参数
	ex := g.Ex{
		"prefix": meta.Prefix,
	}
	if username != "" {
		mb, err := MemberFindOne(username)
		if err != nil && err != sql.ErrNoRows {
			return data, pushLog(err, helper.DBErr)
		}

		// 判断会员是否存在
		if err == sql.ErrNoRows {
			return data, errors.New(helper.UsernameErr)
		}

		uid = mb.UID
		ex["username"] = username
	}
	// 银行卡号参数可选
	if bankcard != "" {
		ex["bank_card_hash"] = fmt.Sprintf("%d", MurmurHash(bankcard, 0))
	}

	fmt.Println("ex = ", ex)
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

	encFields := []string{"realname"}

	for _, v := range cardList {
		ids = append(ids, v.ID)
		encFields = append(encFields, "bankcard"+v.ID)
	}

	encRes, err := grpc_t.Decrypt(uid, true, encFields)
	if err != nil {
		fmt.Println("grpc_t.Decrypt err = ", err)
		return data, errors.New(helper.GetRPCErr)
	}

	for _, v := range cardList {

		key := "bankcard" + v.ID
		val := BankcardData{
			BankCard: v,
			RealName: encRes["realname"],
			Bankcard: encRes[key],
		}

		data = append(data, val)
	}

	return data, nil
}

func BankCardExistRedis(bankcardNo string) error {

	pipe := meta.MerchantRedis.Pipeline()
	ex1_temp := pipe.Do(ctx, "CF.EXISTS", "bankcard_exist", bankcardNo)
	ex2_temp := pipe.Do(ctx, "CF.EXISTS", "bankcard_blacklist", bankcardNo)
	_, err := pipe.Exec(ctx)
	pipe.Close()
	if err != nil {
		return errors.New(helper.RedisErr)
	}

	if val, ok := ex1_temp.Val().(string); ok && val == "1" {
		return errors.New(helper.BankCardExistErr)
	}

	if val, ok := ex2_temp.Val().(string); ok && val == "1" {
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

func BankcardUpdate(bid, bankID, bankAddr, bankcardNo string) error {

	data, err := BankCardFindOne(g.Ex{"id": bid})
	if err != nil {
		return err
	}

	if data.Username == "" {
		return errors.New(helper.BankCardNotExist)
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

	if bankcardNo != "" {

		//判断卡号是否存在
		err = BankCardExistRedis(bankcardNo)
		if err != nil {
			return err
		}

		src := [][]string{
			{"bankcard" + bid, bankcardNo},
		}
		err := grpc_t.Encrypt(bid, src)
		if err != nil {
			fmt.Println("grpc_t.Encrypt = ", err)
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
		"id": bid,
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

	enckey := "bankcard" + bid
	encRes, err := grpc_t.Decrypt(mb.UID, true, []string{enckey})
	if err != nil {
		return errors.New(helper.GetRPCErr)
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
		"id":               helper.GenId(),
		"prefix":           meta.Prefix,
		"bank_card_no":     encRes["enckey"],
		"bank_branch_name": data.BankBranch,
		"bank_address":     data.BankAddress,
		"bank_id":          data.BankID,
		"created_at":       time.Now().Unix(),
	}
	query, _, _ = dialect.Insert("tbl_member_bankcard_blacklist").Rows(bankcard_blacklist_record).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return errors.New(helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		return errors.New(helper.DBErr)
	}

	pipe := meta.MerchantRedis.Pipeline()
	pipe.Do(ctx, "CF.DEL", "bankcard_exist", encRes["enckey"])
	pipe.Do(ctx, "CF.ADD", "bankcard_blacklist", encRes["enckey"])
	pipe.Exec(ctx)
	pipe.Close()

	return nil
}
