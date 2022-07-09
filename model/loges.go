package model

import (
	"database/sql"
	"errors"
	"fmt"
	"merchant/contrib/helper"

	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

func MemberRemarkLogList(uid, adminName, startTime, endTime string, page, pageSize int) (MemberRemarkLogData, error) {

	ex := g.Ex{
		"is_delete": g.Op{"neq": 1},
	}

	if uid != "" {
		ex["uid"] = uid
	}

	if adminName != "" {
		ex["created_name"] = adminName
	}

	data := MemberRemarkLogData{}
	if len(ex) == 0 && (startTime == "" || endTime == "") {
		return data, errors.New(helper.QueryTermsErr)
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

		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}
	ex["prefix"] = meta.Prefix

	t := dialect.From("member_remarks_log")

	if page == 1 {
		query, _, _ := t.Select(g.COUNT("*")).Where(ex).ToSQL()

		fmt.Println(query)

		err := meta.MerchantTD.Get(&data.T, query)
		if err == sql.ErrNoRows {
			return data, nil
		}

		if err != nil {
			fmt.Println("Member Remarks Log err = ", err.Error())
			fmt.Println("Member Remarks Log query = ", query)
			body := fmt.Errorf("%s,[%s]", err.Error(), query)
			return data, pushLog(body, helper.DBErr)
		}
		if data.T == 0 {
			return data, nil
		}
	}

	offset := (page - 1) * pageSize
	query, _, _ := t.Select(colsMemberRemarksLog...).Where(ex).Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("ts").Desc()).ToSQL()
	fmt.Println("Member Remarks Log query = ", query)
	err := meta.MerchantTD.Select(&data.D, query)
	if err != nil {
		fmt.Println("Member Remarks Log err = ", err.Error())
		fmt.Println("Member Remarks Log query = ", query)
		body := fmt.Errorf("%s,[%s]", err.Error(), query)
		return data, pushLog(body, helper.DBErr)
	}

	data.S = pageSize

	return data, nil
}

func MemberLoginLogList(startTime, endTime string, page, pageSize int, ex g.Ex) (MemberLoginLogData, error) {

	data := MemberLoginLogData{}
	if len(ex) == 0 && (startTime == "" || endTime == "") {
		return data, errors.New(helper.QueryTermsErr)
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

		ex["create_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}
	ex["prefix"] = meta.Prefix

	t := dialect.From("member_login_log")
	fmt.Println(ex)
	if page == 1 {
		query, _, _ := t.Select(g.COUNT("*")).Where(ex).ToSQL()
		err := meta.MerchantTD.Get(&data.T, query)
		if err == sql.ErrNoRows {
			return data, nil
		}

		if err != nil {
			fmt.Println("Member Login Log err = ", err.Error())
			fmt.Println("Member Login Log query = ", query)
			body := fmt.Errorf("%s,[%s]", err.Error(), query)
			return data, pushLog(body, helper.DBErr)
		}
		if data.T == 0 {
			return data, nil
		}
	}
	offset := (page - 1) * pageSize
	query, _, _ := t.Select("username", "ip", "device", "device_no", "top_name", "parent_name", "create_at").Where(ex).Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("ts").Desc()).ToSQL()
	fmt.Println("Member Login Log query = ", query)

	err := meta.MerchantTD.Select(&data.D, query)
	if err != nil {
		fmt.Println("Member Login Log err = ", err.Error())
		fmt.Println("Member Login Log query = ", query)
		body := fmt.Errorf("%s,[%s]", err.Error(), query)
		return data, pushLog(body, helper.DBErr)
	}

	data.S = pageSize
	return data, nil
}
