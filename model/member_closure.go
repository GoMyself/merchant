package model

import (
	"fmt"
	"merchant2/contrib/helper"
)

func MemberClosureInsert(nodeID, targetID string) string {

	t := "SELECT ancestor, " + nodeID + ",prefix, lvl+1 FROM tbl_members_tree WHERE prefix='" + meta.Prefix + "' and descendant = " + targetID + " UNION SELECT " + nodeID + "," + nodeID + "," + "'" + meta.Prefix + "'" + ",0"
	query := "INSERT INTO tbl_members_tree (ancestor, descendant,prefix,lvl) (" + t + ")"

	return query
}

func MemberClosureGetParent(uid string) ([]string, error) {

	uids := []string{}
	query := fmt.Sprintf("SELECT ancestor FROM tbl_member_tree WHERE prefix = '%s' and descendant='%s' ORDER BY lvl ASC", meta.Prefix, uid)

	err := meta.MerchantDB.Select(&uids, query)
	if err != nil {
		body := fmt.Errorf("%s,[%s]", err.Error(), query)
		return uids, pushLog(body, helper.DBErr)
	}

	return uids, nil
}

func MemberClosureMove(node_id, target_id string) error {

	query1 := "DELETE a FROM tbl_member_tree AS a "
	query1 += "JOIN tbl_member_tree AS d ON a.descendant = d.descendant "
	query1 += "LEFT JOIN tbl_member_tree AS x "
	query1 += "ON x.ancestor = d.ancestor AND x.descendant = a.ancestor "
	query1 += "WHERE d.ancestor = " + node_id + "  AND x.ancestor IS NULL"

	query2 := "INSERT INTO tbl_member_tree (ancestor, descendant, lvl) "
	query2 += "SELECT a.ancestor, b.descendant, a.lvl+b.lvl+1 "
	query2 += "FROM tbl_member_tree AS a JOIN tbl_member_tree AS b "
	query2 += "WHERE b.ancestor = " + node_id + " AND a.descendant = " + target_id

	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(query1)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	_, err = tx.Exec(query2)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
