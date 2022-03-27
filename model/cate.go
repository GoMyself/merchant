package model

import (
	"bytes"
	"fmt"
	"merchant2/contrib/helper"

	g "github.com/doug-martin/goqu/v9"
	"github.com/minio/minio-go/v7"
)

type Cate struct {
	ID     int64  `db:"id" json:"id"`
	Level  string `db:"level" json:"level"`
	Name   string `db:"name" json:"name"`
	Sort   int    `db:"sort" json:"sort"`
	Prefix string `db:"prefix" json:"prefix"`
}

func CateInit() {

	var top []Cate

	//初始化 一次性加载, 所以不用太注意sql性能
	err := meta.MerchantDB.Select(&top, "SELECT * FROM `tbl_tree` WHERE LENGTH(`level`) = 3 and prefix = '?'", meta.Prefix)
	if err != nil {
		return
	}

	for _, value := range top {

		var c []Cate
		query, _, _ := dialect.From("tbl_tree").
			Where(g.C("level").ILike(value.Level+"%"), g.C("prefix").Eq(meta.Prefix)).Order(g.C("sort").Asc()).ToSQL()
		err = meta.MerchantDB.Select(&c, query)
		if err != nil {
			continue
		}

		b, err := helper.JsonMarshal(c)
		if err != nil {
			continue
		}

		reader := bytes.NewReader(b)
		userMetaData := map[string]string{"x-amz-acl": "public-read"}
		name := fmt.Sprintf("T%s.json", value.Level)

		_, err = meta.MinioClient.PutObject(ctx, meta.MinioJsonBucket, name, reader, reader.Size(),
			minio.PutObjectOptions{ContentType: "application/json", UserMetadata: userMetaData})
		if err != nil {
			fmt.Println(err)
		}
	}
}

func CateDirect(level string) (string, error) {

	return meta.MerchantRedis.HGet(ctx, "tree", level).Result()

}
