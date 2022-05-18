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

	query, _, _ := t.Select("id", "name", "alias", "created_at", "state", "remark", "created_name").
		Where(ex).Order(g.C("state").Desc()).Order(g.C("created_at").Desc()).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Select(&data, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func SMSChannelUpdateState(cid string, state int) error {

	ex := g.Ex{
		"id":     cid,
		"prefix": meta.Prefix,
	}

	query, _, _ := dialect.Update("tbl_sms").Set(g.Record{"state": state}).Where(ex).ToSQL()
	fmt.Println(query)
	_, err := meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	return nil
}
