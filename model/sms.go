package model

import (
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"merchant2/contrib/helper"
)

func SMSChannelList(ex g.Ex) ([]SMSChannel, error) {

	data := make([]SMSChannel, 0)

	ex["prefix"] = meta.Prefix
	t := dialect.From("tbl_sms")

	query, _, _ := t.Select("id", "name", "alias", "created_at", "txt", "voice", "remark", "created_name").
		Where(ex).Order(g.C("created_at").Desc()).Order(g.C("txt").Asc()).Order(g.C("voice").Asc()).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Select(&data, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func SMSChannelUpdateState(cid string, txtState, voiceState int) error {

	ex := g.Ex{
		"id":     cid,
		"prefix": meta.Prefix,
	}

	rc := g.Record{}

	if txtState != 0 {
		rc["txt"] = txtState
	}

	if voiceState != 0 {
		rc["voice"] = voiceState
	}

	query, _, _ := dialect.Update("tbl_sms").Set(rc).Where(ex).ToSQL()
	fmt.Println(query)
	_, err := meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	return nil
}

func SMSChannelUpdate(cid string, channelName, remark string) error {

	ex := g.Ex{
		"id":     cid,
		"prefix": meta.Prefix,
	}

	rc := g.Record{}
	if channelName != "" {
		rc["name"] = channelName
	}
	if remark != "" {
		rc["remark"] = remark
	}

	query, _, _ := dialect.Update("tbl_sms").Set(rc).Where(ex).ToSQL()
	fmt.Println(query)
	_, err := meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	return nil
}
