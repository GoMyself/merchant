package model

import (
	g "github.com/doug-martin/goqu/v9"
	"github.com/shopspring/decimal"
	"merchant2/contrib/helper"
)

func MemberRebateFindOne(uid string) (MemberRebateResult_t, error) {

	data := MemberRebate{}
	res := MemberRebateResult_t{}

	t := dialect.From("tbl_member_rebate_info")
	query, _, _ := t.Select(colsMemberRebate...).Where(g.Ex{"uid": uid}).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&data, query)
	if err != nil {
		return res, pushLog(err, helper.DBErr)
	}

	res.ZR, _ = decimal.NewFromString(data.ZR)
	res.QP, _ = decimal.NewFromString(data.QP)
	res.TY, _ = decimal.NewFromString(data.TY)
	res.DJ, _ = decimal.NewFromString(data.DJ)
	res.DZ, _ = decimal.NewFromString(data.DZ)
	res.CP, _ = decimal.NewFromString(data.CP)

	res.ZR = res.ZR.Truncate(1)
	res.QP = res.QP.Truncate(1)
	res.TY = res.TY.Truncate(1)
	res.DJ = res.DJ.Truncate(1)
	res.DZ = res.DZ.Truncate(1)
	res.CP = res.CP.Truncate(1)

	return res, nil
}
