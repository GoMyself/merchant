package model

import (
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
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
func MessageList(page, pageSize int, sendStartTime, sendEndTime,
	startTime, endTime, reviewStartTime, reviewEndTime string, ex g.Ex) (MessageData, error) {

	data := MessageData{}
	// 没有查询条件  startTime endTime 必填
	if len(ex) == 0 && (startTime == "" || endTime == "") {
		return data, errors.New(helper.QueryTermsErr)
	}

	if sendStartTime != "" && sendEndTime != "" {

		startAt, err := helper.TimeToLoc(sendStartTime, loc)
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		endAt, err := helper.TimeToLoc(sendEndTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		ex["send_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}

	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		endAt, err := helper.TimeToLoc(endTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		ex["apply_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}

	if reviewStartTime != "" && reviewEndTime != "" {

		rStart, err := helper.TimeToLoc(reviewStartTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		rEnd, err := helper.TimeToLoc(reviewEndTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		if rStart >= rEnd {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		ex["review_at"] = g.Op{"between": exp.NewRangeVal(rStart, rEnd)}
	}

	t := dialect.From("tbl_agency_transfer_apply")
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
	query, _, _ := t.Select(colsAgencyTransfer...).Where(ex).
		Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("apply_at").Desc()).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Select(&data.D, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
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
