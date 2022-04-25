package model

import (
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"merchant2/contrib/helper"
)

//MessageInsert  站内信新增
func MessageInsert(record g.Record, sendAt string) error {

	stAt, err := helper.TimeToLoc(sendAt, loc)
	if err != nil {
		return errors.New(helper.DateTimeErr)
	}

	record["send_at"] = stAt
	record["prefix"] = meta.Prefix

	query, _, _ := dialect.Insert("tbl_messages").Rows(record).ToSQL()
	fmt.Println(query)
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

//MessageList  站内信列表
func MessageList() error {
	return nil
}

//MessageUpdate  站内信更新
func MessageUpdate() error {
	return nil
}

//MessageReview  站内信审核
func MessageReview() error {
	return nil
}

//MessageSend  站内信发送
func MessageSend() error {
	return nil
}

//MessageDelete  站内信删除
func MessageDelete() error {
	return nil
}
