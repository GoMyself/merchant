package controller

import (
	"fmt"
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"merchant2/model"
	"strconv"

	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"
)

type PlatformController struct{}

type PlatformParam struct {
	ID    uint64 `rule:"sDigit" min:"10" max:"35" msg:"id  not found" name:"id"`
	Flag  uint8  `rule:"digit" default:"" min:"0" max:"2" msg:"flag  not found,val in [0-2]" name:"flag"`
	State int    `rule:"none" name:"state"`
	Seq   int    `rule:"none" default:"-1" min:"-1" max:"999" msg:"seq not found" name:"seq"`
}

type PlatformListParam struct {
	ID       string `rule:"none" name:"id"`
	State    string `rule:"digit" default:"-1" min:"-1" max:"1" msg:"state error" name:"state"`
	GameType string `rule:"digit" default:"0" min:"0" max:"8" msg:"game_type error" name:"game_type"`
	PageSize uint   `rule:"digit" default:"10" min:"10" max:"200" msg:"page_size error" name:"page_size"`
	Page     uint   `rule:"digit" default:"1" min:"1" msg:"page error" name:"page"`
}

type PlatformID struct {
	ID string `rule:"none" default:"" min:"0" max:"511" msg:"noted error[0-511]" json:"id"`
}

/**
 * @Description: 更新场馆(状态,锁定钱包,排序)
 * flag=0，修改场馆排序字段seq,且seq必传
 * flag=1，维护场馆状态,传参id、flag其他无需传参
 * flag=2，锁定场馆的钱包,传参id、flag其他无需传参
 * @Author: parker
 * @Date: 2021/4/3 10:43
 * @LastEditTime: 2021/4/3 17:43
 * @LastEditors: parker
 */
func (that *PlatformController) Update(ctx *fasthttp.RequestCtx) {

	param := PlatformParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if param.Flag == model.PlatformFlagEdit && param.Seq < 0 {
		helper.Print(ctx, false, helper.PlatSeqErr)
		return
	}

	if param.Flag == model.PlatformFlagState {

		s := map[int]bool{
			0: true,
			1: true,
		}
		if _, ok := s[param.State]; !ok {
			helper.Print(ctx, false, helper.StateParamErr)
			return
		}
	}

	if param.Flag == model.PlatformFlagWallet {

		if !validator.CheckIntScope(fmt.Sprintf("%d", param.State), 0, 2) {
			helper.Print(ctx, false, helper.PlatWalletErr)
			return
		}
	}

	err = model.PlatformUpdate(param.ID, param.Flag, param.State, param.Seq, uint32(ctx.Time().Unix()))
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

/**
 * @Description: 场馆列表
 * @Author: parker
 * @Date: 2021/4/3 10:43
 * @LastEditTime: 2021/4/3 17:43
 * @LastEditors: parker
 */
func (that *PlatformController) List(ctx *fasthttp.RequestCtx) {

	params := PlatformListParam{}
	err := validator.Bind(ctx, &params)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	ex := g.Ex{}
	if params.ID != "" {

		if !validator.CheckStringDigit(params.ID) {
			helper.Print(ctx, false, helper.PlatIDErr)
			return
		}

		ex["id"] = params.ID
	}

	if params.State != "" {

		if !validator.CheckStringDigit(params.State) {
			helper.Print(ctx, false, helper.StateParamErr)
			return
		}

		state, _ := strconv.ParseInt(params.State, 10, 8)
		if state == 0 || state == 1 {
			ex["state"] = state
		}
	}

	if params.GameType != "" {

		if !validator.CheckStringDigit(params.GameType) {
			helper.Print(ctx, false, helper.GameTypeErr)
			return
		}

		gameType, _ := strconv.ParseInt(params.GameType, 10, 8)
		if gameType >= 1 && gameType <= 8 {
			ex["game_type"] = gameType
		}
	}

	data, err := model.PlatformList(ex, params.PageSize, params.Page)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

/**
 * @Description: 场馆列表
 * @Author: parker
 * @Date: 2021/4/3 10:43
 * @LastEditTime: 2021/4/3 17:43
 * @LastEditors: parker
 */
func (that *PlatformController) PlatList(ctx *fasthttp.RequestCtx) {

	data := model.PlatListRedis()
	helper.PrintJson(ctx, true, data)
}

// 场馆费率列表
func (that *PlatformController) PlatRate(ctx *fasthttp.RequestCtx) {

	data, err := model.PlatformRate()
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}
