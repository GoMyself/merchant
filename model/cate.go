package model

type Cate struct {
	ID     int64  `db:"id" json:"id"`
	Level  string `db:"level" json:"level"`
	Name   string `db:"name" json:"name"`
	Sort   int    `db:"sort" json:"sort"`
	Prefix string `db:"prefix" json:"prefix"`
}

func CateDirect(level string) (string, error) {

	return meta.MerchantRedis.HGet(ctx, "tree", level).Result()

}
