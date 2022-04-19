package model

import (
	"bytes"
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/minio/minio-go/v7"
	"merchant2/contrib/helper"
	"strconv"
)

func MemberLevelList() ([]MemberLevel, error) {

	var vip []MemberLevel
	vipQuery, _, _ := dialect.From("tbl_member_level").Select(colsMemberLevel...).Order(g.C("level").Asc()).ToSQL()
	err := meta.MerchantDB.Select(&vip, vipQuery)
	if err != nil {
		return vip, pushLog(fmt.Errorf("%s,[%s]", err.Error(), vipQuery), helper.DBErr)
	}

	return vip, nil
}

func MemberLevelFindOne(ex g.Ex) (MemberLevel, error) {

	var vip MemberLevel
	vipQuery, _, _ := dialect.From("tbl_member_level").Select(colsMemberLevel...).Where(ex).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&vip, vipQuery)
	if err != nil {
		return vip, pushLog(fmt.Errorf("%s,[%s]", err.Error(), vipQuery), helper.DBErr)
	}

	return vip, err
}

func MemberLevelToMinio() {

	vip, err := MemberLevelList()
	if err != nil || len(vip) < 1 {
		return
	}

	b, err := helper.JsonMarshal(vip)
	if err == nil {

		reader := bytes.NewReader(b)
		userMetaData := map[string]string{"x-amz-acl": "public-read"}
		name := "VIP.json"
		_, err := meta.MinioClient.PutObject(ctx, meta.MinioJsonBucket, name, reader, reader.Size(), minio.PutObjectOptions{ContentType: "application/json", UserMetadata: userMetaData})
		if err != nil {
			fmt.Println(err)
		}
	}

	MemberLevelToCache(vip)
}

func MemberLevelToCache(vip []MemberLevel) {

	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()
	pipe.Unlink(ctx, "vip")
	for _, v := range vip {
		mp := vipToMap(v)
		value, _ := helper.JsonMarshal(mp)
		pipe.HSet(ctx, "vip", v.Level, value)
	}

	pipe.Persist(ctx, "vip")
	_, _ = pipe.Exec(ctx)

	return
}

func VIPUpdate(vid string, record g.Record) error {

	vip, err := MemberLevelFindOne(g.Ex{"id": vid})
	if err != nil {
		return err
	}

	query, _, _ := dialect.Update("tbl_member_level").Set(record).Where(g.Ex{"id": vid}).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	_ = vipRefreshCache(vip.Level)
	MemberLevelToMinio()

	return nil
}

func VIPInsert(data MemberLevel) error {

	data.ID = helper.GenId()
	query, _, _ := dialect.Insert("tbl_member_level").Rows(data).ToSQL()
	_, err := meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	_ = vipRefreshCache(data.Level)
	MemberLevelToMinio()

	return nil
}

func VipRecord(page, pageSize uint, startTime, endTime string, ex g.Ex) (MemberLevelRecordData, error) {

	data := MemberLevelRecordData{}

	// 没有查询条件  startTime endTime 必填
	if len(ex) == 0 && (startTime == "" || endTime == "") {
		return data, errors.New(helper.DateTimeErr)
	}

	if _, ok := ex["username"]; !ok {

		if startTime != "" && endTime != "" {

			startAt, err := helper.TimeToLocMs(startTime, loc)
			if err != nil {
				return data, errors.New(helper.TimeTypeErr)
			}

			endAt, err := helper.TimeToLocMs(endTime, loc)
			if err != nil {
				return data, errors.New(helper.TimeTypeErr)
			}

			if startAt >= endAt {
				return data, errors.New(helper.QueryTimeRangeErr)
			}

			ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
		}
	}

	t := dialect.From("tbl_member_level_record")
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
	query, _, _ := t.Select(colLevelRecord...).Where(ex).
		Offset(offset).Limit(pageSize).Order(g.C("created_at").Desc(), g.C("before_level").Desc()).ToSQL()
	err := meta.MerchantDB.Select(&data.D, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

// 刷新vip redis缓存
func vipRefreshCache(level int) error {

	// 因为没有删除的操作，所以这边不处理error=sql.ErrNoRows情况
	vip, err := MemberLevelFindOne(g.Ex{"level": level})
	if err != nil {
		return err
	}

	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()

	mp := vipToMap(vip)
	value, _ := helper.JsonMarshal(mp)

	pipe.HDel(ctx, "vip", fmt.Sprintf("%d", vip.Level))
	pipe.HSet(ctx, "vip", vip.Level, value)
	pipe.Persist(ctx, "vip")

	_, err = pipe.Exec(ctx)
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	return nil
}

func vipToMap(vip MemberLevel) map[string]string {

	return map[string]string{
		"id":                 vip.ID,
		"level":              fmt.Sprintf("%d", vip.Level),
		"level_name":         vip.LevelName,
		"recharge_num":       fmt.Sprintf("%d", vip.RechargeNum),
		"upgrade_deposit":    fmt.Sprintf("%d", vip.UpgradeDeposit),
		"upgrade_record":     fmt.Sprintf("%d", vip.UpgradeRecord),
		"relegation_flowing": fmt.Sprintf("%d", vip.RelegationFlowing),
		"upgrade_gift":       fmt.Sprintf("%d", vip.UpgradeGift),
		"birth_gift":         fmt.Sprintf("%d", vip.BirthGift),
		"withdraw_count":     fmt.Sprintf("%d", vip.WithdrawCount),
		"withdraw_max":       strconv.FormatFloat(vip.WithdrawMax, 'f', -1, 64),
		"early_month_packet": fmt.Sprintf("%d", vip.EarlyMonthPacket),
		"late_month_packet":  fmt.Sprintf("%d", vip.LateMonthPacket),
		"created_at":         fmt.Sprintf("%d", vip.CreateAt),
		"updated_at":         fmt.Sprintf("%d", vip.UpdatedAt),
	}
}