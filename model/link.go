package model

import (
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"merchant2/contrib/helper"
)

func LinkLoad() {

	var total int

	t := dialect.From("tbl_members")
	query, _, _ := t.Select(g.COUNT("uid")).ToSQL()
	err := meta.MerchantDB.Get(&total, query)
	if err != nil {
		fmt.Println(query, err)
		return
	}

	if total > 0 {

		p := total / LINK_PAGE
		if total%LINK_PAGE > 0 {
			p += 1
		}
		for i := 0; i < p; i++ {

			var (
				uids []string
				data []Link_t
			)
			query, _, _ = t.Where(g.Ex{}).Select("uid").Offset(uint(i * LINK_PAGE)).Limit(LINK_PAGE).ToSQL()
			err := meta.MerchantDB.Select(&uids, query)
			if err != nil {
				fmt.Println(query, err)
				return
			}

			ex := g.Ex{
				"uid": uids,
			}
			query, _, _ = dialect.From("tbl_member_link").Where(ex).Select(colsLink...).ToSQL()
			err = meta.MerchantDB.Select(&data, query)
			if err != nil {
				fmt.Println(query, err)
				return
			}

			bcs := make(map[string]map[string]Link_t)
			for _, v := range data {
				key := "lk:" + v.UID
				bcs[key] = map[string]Link_t{
					"$" + v.ID: v,
				}
			}

			pipe := meta.MerchantRedis.TxPipeline()

			for k, v := range bcs {

				value, err := helper.JsonMarshal(&v)
				if err != nil {
					fmt.Println(err)
					return
				}

				pipe.Unlink(ctx, k)
				pipe.Do(ctx, "JSON.SET", k, ".", string(value))
				pipe.Persist(ctx, k)

				fmt.Println(k, string(value))
			}

			_, err = pipe.Exec(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}

			_ = pipe.Close()
		}
	}
}
