package controller

import (
	"fmt"
	"merchant2/contrib/helper"
	"merchant2/contrib/validator"
	"merchant2/model"

	"github.com/valyala/fasthttp"
)

type bankcardInsertParam struct {
	Username    string `rule:"uname" name:"username" min:"4" max:"9" msg:"1031"`
	BankID      string `rule:"digit" name:"bank_id" msg:"bank id error"`
	bankcardNo  string `rule:"digitString" name:"bankcard_no" min:"6" max:"20" msg:"bankcard no error"`
	BankAddress string `rule:"none" name:"bank_addr"`
	Realname    string `rule:"none" name:"realname"`
}

//查询银行卡列表参数
type bankcardUpdateParam struct {
	BID        string `rule:"digit" name:"bid" msg:"bid error"`
	bankcardNo string `rule:"none" name:"bankcard_no"`
	BankAddr   string `rule:"none" name:"bank_addr"`
	BankID     string `rule:"digit" name:"bank_id"`
}

type BankcardController struct{}

func (that *BankcardController) Insert(ctx *fasthttp.RequestCtx) {

	param := bankcardInsertParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		fmt.Println("bankcardInsertParam = ", err)
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	data := model.BankCard{
		ID:          helper.GenId(),
		BankID:      param.BankID,
		Username:    param.Username,
		BankBranch:  param.BankAddress,
		BankAddress: param.BankAddress,
		CreatedAt:   uint64(ctx.Time().Unix()),
	}

	// 更新权限信息
	err = model.BankcardInsert(param.Realname, param.bankcardNo, data)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *BankcardController) List(ctx *fasthttp.RequestCtx) {

	username := string(ctx.PostArgs().Peek("username"))
	bankcardNo := string(ctx.PostArgs().Peek("bank_card"))

	//if username == "" && bankcardNo == "" {
	//	helper.Print(ctx, false, helper.ParamNull)
	//	return
	//}

	if bankcardNo != "" {
		if !validator.CheckStringDigit(bankcardNo) {
			helper.Print(ctx, false, helper.BankcardIDErr)
			return
		}
	}

	if username != "" {
		if !validator.CheckUName(username, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}
	}

	// 更新权限信息
	data, err := model.BankcardList(username, bankcardNo)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *BankcardController) Update(ctx *fasthttp.RequestCtx) {

	//fmt.Println("Update = ", string(ctx.PostBody()))

	param := bankcardUpdateParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if param.BankID == "" && param.BankAddr == "" {
		helper.Print(ctx, false, helper.NoDataUpdate)
		return
	}

	// 更新权限信息
	err = model.BankcardUpdate(param.BID, param.BankID, param.BankAddr, param.bankcardNo)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *BankcardController) Delete(ctx *fasthttp.RequestCtx) {

	bid := string(ctx.QueryArgs().Peek("bid"))
	if !validator.CheckStringDigit(bid) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	// 删除银行卡
	err := model.BankcardDelete(ctx, bid)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
