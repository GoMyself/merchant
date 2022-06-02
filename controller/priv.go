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

	// 获取权限列表
	data, err := model.PrivList()
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.PrintJson(ctx, true, data)

}
