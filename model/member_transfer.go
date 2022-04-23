package model

import (
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"merchant2/contrib/helper"
)

// MemberTransferSubCheck 检查当前会员是否有下级
func MemberTransferSubCheck(username string) bool {

	var num int
	ex := g.Ex{
		"parent_name": username,
	}
	query, _, _ := dialect.From("tbl_members").Select(g.COUNT("uid").As("num")).Where(ex).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Get(&num, query)
	if err == nil && num == 0 {
		return false
	}

	return true
}

//MemberTransferAg 跳线转代
func MemberTransferAg(mb, destMb Member) error {

	tx, err := meta.MerchantDB.Begin() // 开启事务
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	ex := g.Ex{
		"uid":    mb.UID,
		"prefix": meta.Prefix,
	}
	record := g.Record{
		"parent_uid":  destMb.UID,
		"parent_name": destMb.Username,
		"top_uid":     destMb.TopUid,
		"top_name":    destMb.TopName,
	}
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	query = fmt.Sprintf("delete from tbl_members_tree where descendant = %s and prefix = '%s'", mb.UID, meta.Prefix)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	treeNode := MemberClosureInsert(mb.UID, destMb.UID)
	_, err = tx.Exec(treeNode)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	_ = tx.Commit()

	param := map[string]interface{}{
		"uid":         mb.UID,
		"username":    mb.Username,
		"parent_uid":  destMb.UID,
		"parent_name": destMb.Username,
		"top_uid":     destMb.TopUid,
		"top_name":    destMb.TopName,
		"prefix":      meta.Prefix,
	}
	_, _ = BeanBetPut("transfer_ag", param, 0)

	// todo 记录转代日志

	return nil
}

//MemberTransferInsert 团队转代
func MemberTransferInsert(mb, destMb Member) error {
	return nil
}
