package model

import (
	"fmt"
	"merchant2/contrib/helper"

	g "github.com/doug-martin/goqu/v9"
	"github.com/shopspring/decimal"
)

func MemberRebateCmp(uid string, own MemberRebateResult_t) bool {

	lower, err := MemberMaxRebateFindOne(uid)
	if err != nil {
		return false
	}

	if own.QP.Cmp(lower.QP) == -1 {
		return false
	}
	if own.ZR.Cmp(lower.ZR) == -1 {
		return false
	}
	if own.TY.Cmp(lower.TY) == -1 {
		return false
	}
	if own.DJ.Cmp(lower.DJ) == -1 {
		return false
	}
	if own.DZ.Cmp(lower.DZ) == -1 {
		return false
	}
	if own.CP.Cmp(lower.CP) == -1 {
		return false
	}
	if own.FC.Cmp(lower.FC) == -1 {
		return false
	}
	if own.BY.Cmp(lower.BY) == -1 {
		return false
	}
	if own.CGHighRebate.Cmp(lower.CGHighRebate) == -1 {
		return false
	}
	if own.CGOfficialRebate.Cmp(lower.CGOfficialRebate) == -1 {
		return false
	}
	return true

}

func MemberRebateUpdateCache1(uid string, mr MemberRebateResult_t) error {

	key := fmt.Sprintf("%s:m:rebate:%s", meta.Prefix, uid)
	vals := []interface{}{"zr", mr.ZR.Truncate(1), "qp", mr.QP.Truncate(1), "ty", mr.TY.Truncate(1), "dj", mr.DJ.Truncate(1), "dz", mr.DZ.Truncate(1), "cp", mr.CP.Truncate(1), "fc", mr.FC.Truncate(1), "by", mr.BY.Truncate(1), "cg_high_rebate", mr.CGHighRebate.Truncate(2), "cg_official_rebate", mr.CGOfficialRebate.Truncate(2)}

	pipe := meta.MerchantRedis.Pipeline()
	pipe.Del(ctx, key)
	pipe.HMSet(ctx, key, vals...)
	pipe.Persist(ctx, key)
	_, err := pipe.Exec(ctx)
	pipe.Close()

	return err
}

func MemberRebateUpdateCache2(uid string, mr MemberRebate) error {

	key := fmt.Sprintf("%s:m:rebate:%s", meta.Prefix, uid)
	vals := []interface{}{"zr", mr.ZR, "qp", mr.QP, "ty", mr.TY, "dj", mr.DJ, "dz", mr.DZ, "cp", mr.CP, "fc", mr.FC, "by", mr.BY, "cg_high_rebate", mr.CgHighRebate, "cg_official_rebate", mr.CgOfficialRebate}

	pipe := meta.MerchantRedis.Pipeline()
	pipe.Del(ctx, key)
	pipe.HMSet(ctx, key, vals...)
	pipe.Persist(ctx, key)
	_, err := pipe.Exec(ctx)
	pipe.Close()

	return err
}

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
	res.FC, _ = decimal.NewFromString(data.FC)
	res.BY, _ = decimal.NewFromString(data.BY)
	res.CGOfficialRebate, _ = decimal.NewFromString(data.CgOfficialRebate)
	res.CGHighRebate, _ = decimal.NewFromString(data.CgHighRebate)

	res.ZR = res.ZR.Truncate(1)
	res.QP = res.QP.Truncate(1)
	res.TY = res.TY.Truncate(1)
	res.DJ = res.DJ.Truncate(1)
	res.DZ = res.DZ.Truncate(1)
	res.CP = res.CP.Truncate(1)
	res.FC = res.FC.Truncate(1)
	res.BY = res.BY.Truncate(1)

	res.CGOfficialRebate = res.CGOfficialRebate.Truncate(2)
	res.CGHighRebate = res.CGHighRebate.Truncate(2)

	return res, nil
}
