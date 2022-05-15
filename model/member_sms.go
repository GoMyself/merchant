package model

import (
	"database/sql"
	"fmt"
	"merchant2/contrib/helper"

	g "github.com/doug-martin/goqu/v9"
)

type Sms_t struct {
	Username string `json:"username" db:"username"`
	Phone    string `json:"phone" db:"phone"`
	Code     string `json:"code" db:"code"`
	IP       string `json:"ip" db:"ip"`
	CreateAt string `json:"create_at" db:"create_at"`
	Flags    string `json:"flags" db:"flags"`
	Source   string `json:"source" db:"source"`
	ID       string `json:"id" db:"id"`
}

type SmsData_t struct {
	D []Sms_t `json:"d"`
	T int64   `json:"t"`
	S uint    `json:"s"`
}

func SmsList(page, pageSize uint, username, phone string) (SmsData_t, error) {

	ex := g.Ex{}
	data := SmsData_t{}

	if username != "" {
		ex["username"] = username

	}
	if phone != "" {
		ex["phone"] = phone
	}

	t := dialect.From("sms_log")

	if page == 1 {
		query, _, _ := t.Select(g.COUNT("*")).Where(ex).ToSQL()
		err := meta.MerchantTD.Get(&data.T, query)
		if err == sql.ErrNoRows {
			return data, nil
		}

		if err != nil {
			fmt.Println("SmsList COUNT err = ", err.Error())
			fmt.Println("SmsList COUNT query = ", query)
			body := fmt.Errorf("%s,[%s]", err.Error(), query)
			return data, pushLog(body, helper.DBErr)
		}

		if data.T == 0 {
			return data, nil
		}
	}
	//.Order(g.C("ts").Desc())

	offset := (page - 1) * pageSize
	query, _, _ := t.Select("id", "username", "ip", "code", "flags", "source", "phone", "create_at").Where(ex).Offset(offset).Limit(pageSize).ToSQL()
	fmt.Println("SmsList query = ", query)

	err := meta.MerchantTD.Select(&data.D, query)
	if err != nil {
		fmt.Println("SmsList err = ", err.Error())
		fmt.Println("SmsList query = ", query)
		body := fmt.Errorf("%s,[%s]", err.Error(), query)
		return data, pushLog(body, helper.DBErr)
	}

	data.S = pageSize

	return data, nil
}
