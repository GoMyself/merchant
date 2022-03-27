package model

import (
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"merchant2/contrib/helper"
)

func DividendInsert(data g.Record) error {

	data["prefix"] = meta.Prefix

	query, _, _ := dialect.Insert("tbl_member_dividend").Rows(data).ToSQL()
	_, err := meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	_ = PushMerchantNotify(dividendReviewFmt, data["apply_name"].(string), data["username"].(string), data["amount"].(string))

	return nil
}

func DividendList(page, pageSize int, startTime, endTime, reviewStartTime, reviewEndTime string, ex g.Ex) (DividendData, error) {

	data := DividendData{}
	// 没有查询条件  startTime endTime 必填
	if len(ex) == 0 && (startTime == "" || endTime == "") {
		return data, errors.New(helper.QueryTermsErr)
	}

	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLocMs(startTime, loc)
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		endAt, err := helper.TimeToLocMs(endTime, loc)
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
	ex["prefix"] = meta.Prefix
	t := dialect.From("tbl_member_dividend")
	if page == 1 {
		query, _, _ := t.Select(g.COUNT("id")).Where(ex).ToSQL()
		err := meta.MerchantDB.Get(&data.T, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}

		if data.T == 0 {
			return data, nil
		}
	}

	offset := pageSize * (page - 1)
	query, _, _ := t.Select(colsDividend...).Where(ex).
		Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("apply_at").Desc()).ToSQL()
	err := meta.MerchantDB.Select(&data.D, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func DividendReview(state int, t int64, adminID, adminName, reviewRemark string, ids []string) error {

	ex := g.Ex{
		"id":    ids,
		"state": DividendReviewing,
	}
	record := g.Record{
		"state":         state,
		"review_remark": reviewRemark,
		"review_at":     t,
		"review_uid":    adminID,
		"review_name":   adminName,
	}
	err := dividendUpdate(ex, record)
	if err != nil {
		return err
	}

	// 批量/单条不通过
	if state == DividendReviewReject {
		return nil
	}

	for _, id := range ids {
		param := map[string]interface{}{
			"id":            id,                   //红利记录id，字符串
			"review_at":     fmt.Sprintf("%d", t), //字符串
			"review_uid":    adminID,              //字符串
			"review_name":   adminName,            //字符串
			"review_remark": reviewRemark,         // 审核备注
		}

		topic := "dividend"
		_, err := BeanPut(topic, param, 0)
		if err != nil {
			fmt.Printf("红利发送队列写入失败：订单号：%s, errMSg: %s", id, err.Error())
		}
	}

	return nil
}

// 更新红利
func DividendUpdate(ex g.Ex, record g.Record) error {

	return dividendUpdate(ex, record)
}

func DividendGetState(id string) (int, error) {

	var state int
	query, _, _ := dialect.From("tbl_member_dividend").Select("state").Where(g.Ex{"id": id}).ToSQL()
	err := meta.MerchantDB.Get(&state, query)
	if err != nil {
		return state, pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	return state, nil
}

// 会员红利记录更新
func dividendUpdate(ex g.Ex, record g.Record) error {

	ex["prefix"] = meta.Prefix
	t := dialect.Update("tbl_member_dividend")
	query, _, _ := t.Set(record).Where(ex).ToSQL()
	_, err := meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}
