package controller

import (
	"github.com/valyala/fasthttp"
	"merchant/contrib/helper"
	"merchant/model"
)

type PrivController struct{}

/**
 * @Description: 权限列表
 * @Author: carl
 */
func (that *PrivController) List(ctx *fasthttp.RequestCtx) {

	admin, err := model.AdminToken(ctx)
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}
	// 获取权限列表
	data, err := model.PrivList(admin["group_id"])
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.PrintJson(ctx, true, data)

}
