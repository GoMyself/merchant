package model

import (
	"errors"
	"merchant2/contrib/helper"
)

func SmsList(username, phone string) (string, error) {

	if username != "" {
		res, err := meta.MerchantRedis.Get(ctx, "code:"+username).Result()
		if err != nil {
			return "", errors.New(helper.RecordNotExistErr)
		}

		return res, nil
	}
	if phone != "" {
		res, err := meta.MerchantRedis.Get(ctx, "code:"+phone).Result()
		if err != nil {
			return "", errors.New(helper.RecordNotExistErr)
		}

		return res, nil
	}

	return "", errors.New(helper.RecordNotExistErr)
}
