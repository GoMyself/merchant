package model

import (
	"fmt"
	"merchant2/contrib/helper"
	"time"
)

type tree_t struct {
	ID     int    `db:"id" json:"id"`         //
	Level  string `db:"level" json:"level"`   //分类等级
	Name   string `db:"name" json:"name"`     //分类名字
	Sort   int    `db:"sort" json:"sort"`     //排序
	Prefix string `db:"prefix" json:"prefix"` //排序
}

func TreeLoadToRedis() error {

	var parent []tree_t

	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()

	query := fmt.Sprintf("SELECT * FROM `tbl_tree` WHERE LENGTH(`level`) = 3 and prefix= '%s';", meta.Prefix)
	err := meta.MerchantDB.Select(&parent, query)
	if err != nil {
		fmt.Println("TreeLoadToRedis Select = ", err)
		return err
	}

	for _, val := range parent {

		var data []tree_t

		key := fmt.Sprintf("T:%s", val.Level)
		query := fmt.Sprintf("SELECT * FROM `tbl_tree`  WHERE prefix='%s' and `level` LIKE '%s%%' ORDER BY `level` ASC;", meta.Prefix, val.Level)
		err := meta.MerchantDB.Select(&data, query)
		if err != nil {
			fmt.Println("TreeLoadToRedis Select 2 = ", err)
			return err
		}

		data = data[1:]
		b, _ := helper.JsonMarshal(data)

		//fmt.Println("TreeLoadToRedis Select 2 = ", string(b))

		pipe.Unlink(ctx, key)
		pipe.Set(ctx, key, string(b), time.Duration(100)*time.Hour)
		pipe.Persist(ctx, key)

	}

	_, err = pipe.Exec(ctx)
	return err
}

func TreeList(level string) (string, error) {

	data, err := meta.MerchantRedis.Get(ctx, "T:"+level).Result()
	if err != nil {
		return "", err
	}

	return data, nil
}
