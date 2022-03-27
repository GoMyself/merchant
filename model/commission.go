package model

import (
	"database/sql"
	"errors"
	"fmt"
	"merchant2/contrib/helper"
	"time"

	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/shopspring/decimal"
	"github.com/valyala/fasthttp"
)

func CommissionRecordList(page, pageSize int, startTime, endTime, reviewStartTime, reviewEndTime string, ex g.Ex) (CommissionTransferData, error) {

	data := CommissionTransferData{}
	if startTime != "" && endTime != "" {
		//判断日期
		startAt, err := helper.TimeToLoc(startTime, loc) // 毫秒级时间戳
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}
		endAt, err := helper.TimeToLoc(endTime, loc) // 毫秒级时间戳
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}

	if reviewStartTime != "" && reviewEndTime != "" {
		//判断日期
		startAt, err := helper.TimeToLoc(reviewStartTime, loc) // 毫秒级时间戳
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}
		endAt, err := helper.TimeToLoc(reviewEndTime, loc) // 毫秒级时间戳
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		ex["review_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}
	ex["prefix"] = meta.Prefix
	t := dialect.From("tbl_commission_transfer")
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
	query, _, _ := t.Select(colsCommissionTransfer...).Where(ex).
		Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("created_at").Desc()).ToSQL()
	err := meta.MerchantDB.Select(&data.D, query)
	if err != nil && err != sql.ErrNoRows {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func CommissionRecordReview(state int, ts int64, adminID, adminName, reviewRemark string, ids []string) error {

	var (
		data []CommissionTransfer
	)

	ex := g.Ex{
		"id":     ids,
		"prefix": meta.Prefix,
	}
	query, _, _ := dialect.From("tbl_commission_transfer").Select(colsCommissionTransfer...).Where(ex).ToSQL()
	err := meta.MerchantDB.Select(&data, query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	record := g.Record{
		"state":         state,
		"review_at":     ts,
		"review_uid":    adminID,
		"review_name":   adminName,
		"review_remark": reviewRemark,
	}

	// 审核通过
	if state == 2 {
		for _, v := range data {
			if v.State == state {
				continue
			}

			switch v.TransferType {
			case 1: //佣金发放
			case 2: //佣金提取
				_ = commissionDrawPass(v, record)
			case 3: //佣金下发
				_ = commissionRationPass(v, record)
			}
		}

		return nil
	}

	// 审核拒绝
	for _, v := range data {
		if v.State == state {
			continue
		}

		switch v.TransferType {
		case 1: //佣金发放
		case 2: //佣金提取
			_ = commissionDrawReject(v, record)
		case 3: //佣金下发
			_ = commissionRationReject(v, record)
		}
	}

	return nil

}

// 佣金下发审核通过
func commissionRationPass(data CommissionTransfer, reviewRecord g.Record) error {

	id := helper.GenId()
	mb, err := MemberFindOne(data.Username)
	if err != nil {
		return err
	}

	rmb, err := MemberFindOne(data.ReceiveName)
	if err != nil {
		return err
	}

	commission, _ := decimal.NewFromString(rmb.Commission)
	amount, _ := decimal.NewFromString(data.Amount)
	lockAmount, _ := decimal.NewFromString(mb.LockAmount)
	if amount.GreaterThan(lockAmount) {
		return errors.New(helper.LackOfBalance)
	}

	trans := MemberTransaction{
		AfterAmount:  commission.Add(amount).String(),
		Amount:       amount.String(),
		BeforeAmount: commission.String(),
		BillNo:       id,
		CreatedAt:    time.Now().UnixMilli(),
		ID:           id,
		CashType:     COTransactionReceive,
		UID:          data.ReceiveUID,
		Username:     data.ReceiveName,
		Prefix:       meta.Prefix,
	}

	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	query, _, _ := dialect.Update("tbl_commission_transfer").Set(reviewRecord).Where(g.Ex{"id": data.ID, "prefix": meta.Prefix}).ToSQL()
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	query, _, _ = dialect.Insert("tbl_commission_transaction").Rows(trans).ToSQL()
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	record := g.Record{
		"lock_amount": g.L(fmt.Sprintf("lock_amount-%s", amount.String())),
	}
	query, _, _ = dialect.Update("tbl_members").Set(record).Where(g.Ex{"uid": data.UID}).ToSQL()
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	record = g.Record{
		"commission": g.L(fmt.Sprintf("commission+%s", amount.String())),
	}
	query, _, _ = dialect.Update("tbl_members").Set(record).Where(g.Ex{"uid": data.ReceiveUID}).ToSQL()
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	return tx.Commit()
}

// 佣金下发审核拒绝
func commissionRationReject(data CommissionTransfer, reviewRecord g.Record) error {

	id := helper.GenId()
	mb, err := MemberFindOne(data.Username)
	if err != nil {
		return err
	}

	commission, _ := decimal.NewFromString(mb.Commission)
	amount, _ := decimal.NewFromString(data.Amount)
	lockAmount, _ := decimal.NewFromString(mb.LockAmount)
	if amount.GreaterThan(lockAmount) {
		return errors.New(helper.LackOfBalance)
	}

	trans := MemberTransaction{
		AfterAmount:  commission.Add(amount).String(),
		Amount:       amount.String(),
		BeforeAmount: commission.String(),
		BillNo:       id,
		CreatedAt:    time.Now().UnixMilli(),
		ID:           id,
		CashType:     COTransactionRationBack,
		UID:          data.UID,
		Username:     data.Username,
		Prefix:       meta.Prefix,
	}

	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	query, _, _ := dialect.Update("tbl_commission_transfer").Set(reviewRecord).Where(g.Ex{"id": data.ID, "prefix": meta.Prefix}).ToSQL()
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	query, _, _ = dialect.Insert("tbl_commission_transaction").Rows(trans).ToSQL()
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	record := g.Record{
		"commission":  g.L(fmt.Sprintf("commission+%s", amount.String())),
		"lock_amount": g.L(fmt.Sprintf("lock_amount-%s", amount.String())),
	}
	ex := g.Ex{
		"uid": data.UID,
	}
	query, _, _ = dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	return tx.Commit()
}

// 佣金提取
func commissionDrawPass(data CommissionTransfer, reviewRecord g.Record) error {

	id := helper.GenId()
	mb, err := MemberFindOne(data.ReceiveName)
	if err != nil {
		return err
	}

	balance, _ := decimal.NewFromString(mb.Balance)
	amount, _ := decimal.NewFromString(data.Amount)
	lockAmount, _ := decimal.NewFromString(mb.LockAmount)
	if amount.GreaterThan(lockAmount) {
		return errors.New(helper.LackOfBalance)
	}

	trans := MemberTransaction{
		AfterAmount:  balance.Add(amount).String(),
		Amount:       amount.String(),
		BeforeAmount: balance.String(),
		BillNo:       id,
		CreatedAt:    time.Now().UnixMilli(),
		ID:           id,
		CashType:     TransactionCommissionDraw,
		UID:          data.ReceiveUID,
		Username:     data.ReceiveName,
		Prefix:       meta.Prefix,
	}

	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	query, _, _ := dialect.Update("tbl_commission_transfer").Set(reviewRecord).Where(g.Ex{"id": data.ID, "prefix": meta.Prefix}).ToSQL()
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	query, _, _ = dialect.Insert("tbl_balance_transaction").Rows(trans).ToSQL()
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	record := g.Record{
		"balance":     g.L(fmt.Sprintf("balance+%s", amount.String())),
		"lock_amount": g.L(fmt.Sprintf("lock_amount-%s", amount.String())),
	}
	ex := g.Ex{
		"uid": mb.UID,
	}
	query, _, _ = dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	return tx.Commit()
}

// 佣金提取
func commissionDrawReject(data CommissionTransfer, reviewRecord g.Record) error {

	id := helper.GenId()
	mb, err := MemberFindOne(data.ReceiveName)
	if err != nil {
		return err
	}

	commission, _ := decimal.NewFromString(mb.Commission)
	amount, _ := decimal.NewFromString(data.Amount)
	lockAmount, _ := decimal.NewFromString(mb.LockAmount)
	if amount.GreaterThan(lockAmount) {
		return errors.New(helper.LackOfBalance)
	}

	trans := MemberTransaction{
		AfterAmount:  commission.Add(amount).String(),
		Amount:       amount.String(),
		BeforeAmount: commission.String(),
		BillNo:       id,
		CreatedAt:    time.Now().UnixMilli(),
		ID:           id,
		CashType:     COTransactionDrawBack,
		UID:          data.UID,
		Username:     data.Username,
		Prefix:       meta.Prefix,
	}

	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	query, _, _ := dialect.Update("tbl_commission_transfer").Set(reviewRecord).Where(g.Ex{"id": data.ID, "prefix": meta.Prefix}).ToSQL()
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	query, _, _ = dialect.Insert("tbl_commission_transaction").Rows(trans).ToSQL()
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	record := g.Record{
		"commission":  g.L(fmt.Sprintf("commission+%s", amount.String())),
		"lock_amount": g.L(fmt.Sprintf("lock_amount-%s", amount.String())),
	}
	ex := g.Ex{
		"uid": mb.UID,
	}
	query, _, _ = dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	return tx.Commit()
}

// CommissionPlanInsert 添加佣金方案
func CommissionPlanInsert(ctx *fasthttp.RequestCtx, name, month string, detailsData []byte) error {

	monthStart := helper.MonthSET(month, loc).Unix()

	var details []CommissionDetail
	err := helper.JsonUnmarshal(detailsData, &details)
	if err != nil {
		return errors.New(helper.FormatErr)
	}

	admin, err := AdminToken(ctx)
	if err != nil {
		return err
	}

	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	id := helper.GenId()
	plan := CommissionPlan{
		ID:              id,
		Name:            name,
		CommissionMonth: monthStart,
		CreatedAt:       ctx.Time().Unix(),
		UpdatedUID:      admin["id"],
		UpdatedName:     admin["name"],
		UpdatedAt:       ctx.Time().Unix(),
		Prefix:          meta.Prefix,
	}

	query, _, _ := dialect.Insert("tbl_commission_plan").Rows(plan).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	for k := range details {
		details[k].ID = helper.GenId()
		details[k].PlanID = id
		details[k].Prefix = meta.Prefix
	}

	query, _, _ = dialect.Insert("tbl_commission_detail").Rows(details).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	return nil
}

func CommissionPlanUpdate(ctx *fasthttp.RequestCtx, id, name, month string, detailsData []byte) error {

	// 检测佣金方案是否在被使用
	var number int
	ex := g.Ex{
		"plan_id": id,
		"prefix":  meta.Prefix,
	}
	query, _, _ := dialect.From("tbl_commission_conf").Select(g.COUNT(1)).Where(ex).ToSQL()
	err := meta.MerchantDB.Get(&number, query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	if number > 0 {
		return errors.New(helper.UsedCoPlanEditNotAllow)
	}

	monthStart := helper.MonthSET(month, loc).Unix()

	var details []CommissionDetail
	err = helper.JsonUnmarshal(detailsData, &details)
	if err != nil {
		return errors.New(helper.FormatErr)
	}

	admin, err := AdminToken(ctx)
	if err != nil {
		return err
	}

	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	plan := g.Record{
		"name":             name,
		"commission_month": monthStart,
		"updated_uid":      admin["id"],
		"updated_name":     admin["name"],
		"updated_at":       ctx.Time().Unix(),
	}

	ex = g.Ex{
		"id": id,
	}
	query, _, _ = dialect.Update("tbl_commission_plan").Set(plan).Where(ex).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	ex = g.Ex{
		"plan_id": id,
		"prefix":  meta.Prefix,
	}
	query, _, _ = dialect.Delete("tbl_commission_detail").Where(ex).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	for k := range details {
		details[k].ID = helper.GenId()
		details[k].PlanID = id
		details[k].Prefix = meta.Prefix
	}

	query, _, _ = dialect.Insert("tbl_commission_detail").Rows(details).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	return nil
}

func CommissionPlanList(ex g.Ex, startTime, endTime, updateStartTime, updateEndTime string, page, pageSize int) (CommPlanPageData, error) {

	data := CommPlanPageData{}

	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
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

	if updateStartTime != "" && updateEndTime != "" {

		startAt, err := helper.TimeToLoc(updateStartTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.TimeToLoc(updateEndTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}
		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}
		ex["updated_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}
	ex["prefix"] = meta.Prefix
	t := dialect.From("tbl_commission_plan")
	if page == 1 {
		totalQuery, _, _ := t.Select(g.COUNT(1)).Where(ex).ToSQL()
		err := meta.MerchantDB.Get(&data.T, totalQuery)
		if err != nil && err != sql.ErrNoRows {
			return data, pushLog(err, helper.DBErr)
		}

		if data.T == 0 {
			return data, nil
		}
	}
	offset := (page - 1) * pageSize

	query, _, _ := t.Select(colsCommPlan...).Where(ex).Order(g.I("updated_at").Desc()).Offset(uint(offset)).Limit(uint(pageSize)).ToSQL()
	err := meta.MerchantDB.Select(&data.D, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	data.S = uint(pageSize)

	if len(data.D) == 0 {
		return data, nil
	}

	var ids []string
	for _, plan := range data.D {
		ids = append(ids, plan.ID)
	}

	var details []CommissionDetail
	ex = g.Ex{
		"plan_id": ids,
		"prefix":  meta.Prefix,
	}
	query, _, _ = dialect.From("tbl_commission_detail").Where(ex).ToSQL()
	err = meta.MerchantDB.Select(&details, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	var detailsData = make(map[string][]CommissionDetail)
	for _, detail := range details {
		detailsData[detail.PlanID] = append(detailsData[detail.PlanID], detail)
	}

	data.Details = detailsData

	return data, nil
}

func CommissionPlanDetail(ex g.Ex) ([]CommissionDetail, error) {

	var data []CommissionDetail
	ex["prefix"] = meta.Prefix
	query, _, _ := dialect.From("tbl_commission_detail").Select(colsCommPlanDetail...).Where(ex).ToSQL()
	err := meta.MerchantDB.Select(&data, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func CommissionPlanFind(ex g.Ex) (CommissionPlan, error) {

	var data CommissionPlan
	ex["prefix"] = meta.Prefix
	query, _, _ := dialect.From("tbl_commission_plan").Select(colsCommPlan...).Where(ex).ToSQL()
	err := meta.MerchantDB.Get(&data, query)
	if err != nil && err != sql.ErrNoRows {
		return data, pushLog(err, helper.DBErr)
	}

	if err == sql.ErrNoRows {
		return data, errors.New(helper.RecordNotExistErr)
	}

	return data, nil
}

//总代佣金
func TopCommissionList(sortField string, isAsc, page, pageSize int, day string, ex g.Ex) (CommissionsData, error) {

	data := CommissionsData{}
	if day != "" {
		//判断日期
		startAt := helper.MonthSST(day, loc).Unix() // 毫秒级时间戳

		ex["commission_month"] = startAt
	}
	ex["prefix"] = meta.Prefix
	t := dialect.From("tbl_commissions")
	if page == 1 {
		query, _, _ := t.Select(g.COUNT("id")).Where(ex).ToSQL()
		fmt.Println("总代佣金:sql:", query)
		err := meta.MerchantDB.Get(&data.T, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}

		if data.T == 0 {
			return data, nil
		}
	}

	orderField := g.L("team_num")
	if sortField != "" {
		orderField = g.L(sortField)
	}

	orderBy := orderField.Desc()
	if isAsc == 1 {
		orderBy = orderField.Asc()
	}

	offset := pageSize * (page - 1)
	query, _, _ := t.Select(colosTblCommissions...).Where(ex).
		Offset(uint(offset)).Limit(uint(pageSize)).Order(orderBy).ToSQL()
	fmt.Println("总代佣金:sql:", query)
	err := meta.MerchantDB.Select(&data.D, query)
	if err != nil && err != sql.ErrNoRows {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func CommissionRation(ts int64, adminID, adminName string, ids []string) error {

	var (
		data []Commissions
	)

	ex := g.Ex{
		"id":     ids,
		"prefix": meta.Prefix,
	}
	query, _, _ := dialect.From("tbl_commissions").Select(colosTblCommissions...).Where(ex).ToSQL()
	err := meta.MerchantDB.Select(&data, query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	// 发放佣金
	for _, v := range data {
		if v.State == 1 {
			err := commissionPay(v, ts, adminID, adminName)
			if err != nil {
				return pushLog(err, helper.DBErr)
			}
		}
	}

	return nil
}

func commissionPay(data Commissions, ts int64, adminID, adminName string) error {

	id := helper.GenId()
	mb, err := MemberFindOne(data.Username)
	if err != nil {
		return err
	}

	commission, _ := decimal.NewFromString(mb.Commission)
	amount := decimal.NewFromFloat(data.Amount)

	trans := CommissionTransaction{
		AfterAmount:  commission.Add(amount).String(),
		Amount:       amount.String(),
		BeforeAmount: commission.String(),
		BillNo:       id,
		CreatedAt:    time.Now().UnixMilli(),
		Id:           id,
		CashType:     COTransactionReceive,
		Uid:          data.Uid,
		Username:     data.Username,
		Prefix:       data.Prefix,
	}

	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	data.State = 2
	data.HandOutAt = ts
	data.HandOutUid = adminID
	data.HandOutName = adminName

	query, _, _ := dialect.Update("tbl_commissions").Set(data).Where(g.Ex{"id": data.Id}).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	query, _, _ = dialect.Insert("tbl_commission_transaction").Rows(trans).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	transfer := CommissionTransfer{
		ID:           helper.GenId(),
		State:        2,
		ReviewAt:     ts,
		ReceiveUID:   data.Uid,
		ReceiveName:  adminName,
		ReviewRemark: "佣金发放",
		TransferType: 1,
		UID:          data.Uid,
		Username:     data.Username,
		ReviewUid:    adminID,
		ReviewName:   data.Username,
		Amount:       amount.String(),
		CreatedAt:    time.Now().Unix(),
		Automatic:    1,
		Prefix:       meta.Prefix,
	}

	query, _, _ = dialect.Insert("tbl_commission_transfer").Rows(transfer).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	record := g.Record{
		"commission": g.L(fmt.Sprintf("commission+%s", amount.String())),
	}
	ex := g.Ex{
		"uid": data.Uid,
	}
	query, _, _ = dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	return tx.Commit()
}
