package model

import (
	"fmt"
	"github.com/wI2L/jettison"
)

func VerifyCodeList(username, phone string, endAt int64) (string, error) {

	startAt := endAt - int64(86400)
	param := map[string]interface{}{}
	if username != "" {
		param["username"] = username
	}
	if phone != "" {
		param["phone_hash"] = fmt.Sprintf("%d", MurmurHash(phone, 0))
	}

	rangeParam := map[string][]interface{}{
		"create_at": {startAt, endAt},
	}

	data, _, err := SmsESQuery(esPrefixIndex("smslog"), "create_at", 1, 1, param, rangeParam)

	if err != nil {
		return "", err
	}

	return data, nil
}

func SmsESQuery(index, sortField string, page, pageSize int,
	param map[string]interface{}, rangeParam map[string][]interface{}) (string, string, error) {

	fields := []string{"username", "ip", "create_at", "code", "phone", "phone_hash"}
	total, esData, _, err := esSearch(index, sortField, page, pageSize, fields, param, rangeParam, map[string]string{})
	if err != nil {
		return `{"t":0,"d":[]}`, "", err
	}

	data := smsData{}

	data.T = total
	for _, v := range esData {

		sms := smsLog{}
		_ = cjson.Unmarshal(v.Source, &sms)
		data.D = append(data.D, sms)
	}

	b, err := jettison.Marshal(data)
	if err != nil {
		return "", "", err
	}

	return string(b), "", nil
}
