package model

import (
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"merchant2/contrib/helper"
	"merchant2/message"
	"strconv"
	"strings"
	"time"
)

//  站内信模板添加
func MessageTemplateInsert(template message.Template) error {

	flg, err := message.TemplateInsert(meta.MerchantDB, template)
	if err != nil {
		return pushLog(err, flg)
	}

	return nil
}

// 站内信模板更新
func MessageTemplateUpdate(ex g.Ex, record g.Record) error {

	flg, err := message.TemplateUpdate(meta.MerchantDB, ex, record)
	if err != nil {
		return pushLog(err, flg)
	}

	return nil
}

// 站内信模板删除
func MessageTemplateDelete(ids []string) error {

	flg, err := message.TemplateDelete(meta.MerchantDB, ids)
	if err != nil {
		return pushLog(err, flg)
	}

	return nil
}

// 站内信模板列表
func MessageTemplateList(page, pageSize uint, startTime, endTime string, ex g.Ex) (message.TemplateData, error) {

	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return message.TemplateData{}, errors.New(helper.DateTimeErr)
		}

		endAt, err := helper.TimeToLoc(endTime, loc)
		if err != nil {
			return message.TemplateData{}, errors.New(helper.DateTimeErr)
		}

		if startAt >= endAt {
			return message.TemplateData{}, errors.New(helper.QueryTimeRangeErr)
		}

		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}

	d, flg, err := message.TemplateList(meta.MerchantDB, page, pageSize, ex)
	if err != nil {
		return d, pushLog(err, flg)
	}

	return d, nil
}

// 公告添加
func MessagePostsInsert(posts message.Posts, topStartTime, topEndTime, showStartTime, showEndTime string) error {

	if posts.Top == message.Yes {

		st, err := helper.TimeToLoc(topStartTime, loc)
		if err != nil {
			return errors.New(helper.TimeTypeErr)
		}

		et, err := helper.TimeToLoc(topEndTime, loc)
		if err != nil {
			return errors.New(helper.TimeTypeErr)
		}

		posts.TopStartTime = st
		posts.TopEndTime = et
	}

	if posts.IsShow == message.Yes {

		st, err := helper.TimeToLoc(showStartTime, loc)
		if err != nil {
			return errors.New(helper.TimeTypeErr)
		}

		et, err := helper.TimeToLoc(showEndTime, loc)
		if err != nil {
			return errors.New(helper.TimeTypeErr)
		}

		posts.ShowStartTime = st
		posts.ShowEndTime = et
	}

	f, err := message.PostsInsert(meta.MerchantDB, posts)
	if err != nil {
		return pushLog(err, f)
	}

	f, err = message.RefreshToCache(meta.MerchantDB, meta.MerchantRedis, posts.ID)
	if err != nil {
		return pushLog(err, f)
	}

	return nil
}

// 公告列表
func MessagePostsList(page, pageSize uint, startTime, endTime string, ex g.Ex) (message.PostsData, error) {

	var data message.PostsData

	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		endAt, err := helper.TimeToLoc(endTime, loc)
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}

	data, f, err := message.PostsList(meta.MerchantDB, page, pageSize, ex)
	if err != nil {
		return data, pushLog(err, f)
	}

	return data, nil
}

// 站内列表
func MessageLetterList(page, pageSize uint, startTime, endTime string, ex g.Ex) (message.LetterData, error) {

	var data message.LetterData

	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		endAt, err := helper.TimeToLoc(endTime, loc)
		if err != nil {
			return data, errors.New(helper.DateTimeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}

		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	}

	data, f, err := message.LetterList(meta.MerchantDB, page, pageSize, ex)
	if err != nil {
		return data, pushLog(err, f)
	}

	return data, nil
}

// 公告 启用 停用
func MessagePostsState(id, adminName, uid, remark string, t int64, state int) error {

	origin, f, err := message.PostsFind(meta.MerchantDB, id)
	if err != nil {
		return pushLog(err, f)
	}

	if state == message.PostsStateReject && origin.State != message.PostsStateReviewing { // 拒绝
		return errors.New(helper.CanOnlyDealReviewing)
	}

	if state == message.PostsStateBegin && origin.State != message.PostsStateReviewing && origin.State != message.PostsStateStop { // 启用
		return errors.New(helper.CanOnlyOpenReviewing)
	}

	if state == message.PostsStateStop && origin.State != message.PostsStateBegin { // 停用
		return errors.New(helper.OnlyOpenStopStatus)
	}

	ex := g.Ex{
		"id": id,
	}

	record := g.Record{
		"state":         state,
		"review_at":     t,
		"review_uid":    uid,
		"review_name":   adminName,
		"review_remark": remark,
	}

	f, err = message.PostsUpdate(meta.MerchantDB, ex, record)
	if err != nil {
		return pushLog(err, f)
	}

	if state == message.PostsStateBegin && (origin.ShowStartTime > 0 || origin.ShowEndTime > 0) { // 首次启用写入 es
		//添加脚本自动启用停用
		now := time.Now()
		st := time.Unix(origin.ShowStartTime, 0)
		et := time.Unix(origin.ShowEndTime, 0)

		if origin.ShowStartTime > 0 {
			// 公告自动启用
			sDelay := st.Sub(now).Seconds()
			_, _ = BeanPut("message", map[string]interface{}{"id": id, "state": "3", "ty": "1"}, int(sDelay)-5)
		}

		if origin.ShowEndTime > 0 {
			// 公告自动停用
			eDelay := et.Sub(now).Seconds()
			_, _ = BeanPut("message", map[string]interface{}{"id": id, "state": "4", "ty": "1"}, int(eDelay)-5)
		}
	}

	f, err = message.RefreshToCache(meta.MerchantDB, meta.MerchantRedis, id)
	if err != nil {
		return pushLog(err, f)
	}

	return nil
}

// 公告删除
func MessagePostsDelete(ids []string) error {

	for _, v := range ids {
		posts, f, err := message.PostsFind(meta.MerchantDB, v)
		if err != nil {
			return pushLog(err, f)
		}

		if posts.Ty == message.PostTySpecial {

			num := 0
			query, _, _ := dialect.From("tbl_messages_posts").Where(g.Ex{"ty": message.PostTySpecial}).Select(g.COUNT("id")).ToSQL()
			err = meta.MerchantDB.Get(&num, query)
			if err != nil {
				return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
			}

			if num == 1 {
				return errors.New(helper.MustLeaveAtLeastOneNotice)
			}
		}

		f, err = message.PostsDelete(meta.MerchantDB, []string{v})
		if err != nil {
			return pushLog(err, f)
		}
	}

	return nil
}

func MessageLoadToCache() error {

	f, err := message.LoadToCache(meta.MerchantDB, meta.MerchantRedis)
	if err != nil {
		return pushLog(err, f)
	}

	return nil
}

// 公告审核详情
func MessagePostsReviewDetail(id string) (message.PostReviewDetail, error) {

	d, f, err := message.PostsReviewDetail(meta.MerchantDB, id)
	if err != nil {
		return d, pushLog(err, f)
	}

	return d, nil
}

// 公告更新
func MessagePostsUpdate(id, title, content string) error {

	record := g.Record{
		"title":   title,
		"content": content,
	}
	ex := g.Ex{
		"id": id,
	}

	f, err := message.PostsUpdate(meta.MerchantDB, ex, record)
	if err != nil {
		return pushLog(err, f)
	}

	param := map[string]interface{}{
		"posts_id": id,
		"ty":       message.PostUpdateOperating,
		"title":    title,
		"content":  content,
	}

	_, err = BeanPut("message", param, 0)
	if err != nil {
		fmt.Println("PostUpdateOperating err:", err.Error())
	}

	f, err = message.RefreshToCache(meta.MerchantDB, meta.MerchantRedis, id)
	if err != nil {
		return pushLog(err, f)
	}

	return nil
}

// 站内信更新
func MessageLetterUpdate(id, title, content string) error {

	record := g.Record{
		"title":   title,
		"content": content,
	}
	ex := g.Ex{
		"id": id,
	}
	f, err := message.LetterUpdate(meta.MerchantDB, ex, record)
	if err != nil {
		return pushLog(err, f)
	}

	param := map[string]interface{}{
		"letter_id": id,
		"ty":        message.LetterUpdateOperating,
		"title":     title,
		"content":   content,
	}
	_, err = BeanPut("message", param, 0)
	if err != nil {
		fmt.Println("LetterUpdateOperating err:", err.Error())
	}

	return nil
}

// 站内信添加
func MessageLetterInsert(letter message.Letter, usernames string) error {

	flg, err := message.LetterInsert(meta.MerchantDB, letter)
	if err != nil {
		return pushLog(err, flg)
	}

	names := strings.Split(usernames, ",")
	size := 100
	total := (len(names) + size - 1) / size // total 批
	for i := 0; i < total; i++ {

		var tempNames []string
		if i*size+size > len(names) {
			tempNames = names[i*size:]
		} else {
			tempNames = names[i*size : (i*size + size)]
		}

		// 获取当前用户名的信息
		members, err := memberFindBatch(tempNames)
		if err != nil {
			return pushLog(err, flg)
		}

		for k, v := range members {

			param := map[string]interface{}{
				"ty":        message.LetterMembersInsertOperating,
				"letter_id": letter.ID,
				"username":  k,
				"uid":       v.UID,
			}
			_, err = BeanPut("message", param, 0)
			if err != nil {
				fmt.Println("LetterMembersInsertOperating Err: ", err.Error())
			}
		}
	}

	return nil
}

// 站内信启用 停用
func MessageLetterState(id, adminName, uid, remark string, t int64, state int) error {

	ex := g.Ex{
		"id": id,
	}
	record := g.Record{
		"state": state,
	}
	origin, f, err := message.LetterFind(meta.MerchantDB, id)
	if err != nil {
		return pushLog(err, f)
	}

	if state == message.PostsStateReject { // 拒绝

		if origin.State != message.PostsStateReviewing {
			return errors.New(helper.CanOnlyDealReviewing)
		}

		record["review_at"] = t
		record["review_uid"] = uid
		record["review_name"] = adminName
		record["review_remark"] = remark
	}

	if state == message.PostsStateBegin { // 启用

		if origin.State != message.PostsStateReviewing && origin.State != message.PostsStateStop {
			return errors.New(helper.CanOnlyOpenReviewing)
		}

		record["review_at"] = t
		record["review_uid"] = uid
		record["review_name"] = adminName
		record["review_remark"] = remark
	}

	if state == message.PostsStateStop { // 停用

		if origin.State != message.PostsStateBegin {
			return errors.New(helper.OnlyOpenStopStatus)
		}

		record["review_at"] = t
		record["review_uid"] = uid
		record["review_name"] = adminName
		record["review_remark"] = remark
	}

	f, err = message.LetterUpdate(meta.MerchantDB, ex, record)
	if err != nil {
		return pushLog(err, f)
	}

	if origin.State == message.PostsStateReviewing && state == message.PostsStateBegin { // 首次启用写入 es

		param := make(map[string]interface{})
		if origin.IsAll == message.Yes { // 全员发放

			param["letter_id"] = id
			param["ty"] = message.LetterInsertOperating
		} else {

			param["letter_id"] = id
			param["ty"] = message.LetterUpdateOperating
			param["state"] = strconv.Itoa(state)
		}

		_, err = BeanPut("message", param, 0)
		if err != nil {
			fmt.Println("LetterEsOperating Err: ", err.Error())
		}
	}

	if (origin.State == message.PostsStateStop && state == message.PostsStateBegin) || // 原来是关闭 现在开启  修改es 状态
		(origin.State == message.PostsStateBegin && state == message.PostsStateStop) { // 原来是开启 现在关闭 修改es 状态

		param := map[string]interface{}{
			"letter_id": id,
			"ty":        message.LetterUpdateOperating,
			"state":     strconv.Itoa(state),
		}
		_, err = BeanPut("message", param, 0)
		if err != nil {
			fmt.Println("LetterUpdateOperating Err: ", err.Error())
		}
	}

	return nil
}

// 站内信删除
func MessageLetterDelete(ids []string) error {

	for _, v := range ids {

		letter, f, err := message.LetterFind(meta.MerchantDB, v)
		if err != nil {
			return pushLog(err, f)
		}

		f, err = message.LetterDelete(meta.MerchantDB, []string{v})
		if err != nil {
			return pushLog(err, f)
		}

		if letter.State == message.PostsStateBegin || letter.State == message.PostsStateStop {
			// 删除Es数据
			param := map[string]interface{}{
				"letter_id": v,
				"ty":        message.LetterDeleteOperating,
			}
			_, err = BeanPut("message", param, 0)
			if err != nil {
				fmt.Println("LetterDeleteOperating Err: ", err.Error())
			}
		}
	}

	return nil
}

// 站内信审核详情
func MessageLetterReviewDetail(id string) (message.PostReviewDetail, error) {

	data, f, err := message.LetterReviewDetail(meta.MerchantDB, id)
	if err != nil {
		return data, pushLog(err, f)
	}

	return data, nil
}

// 站内信通知模板列表
func MessageSystemTplList(keyword, title, scene string, state int8, module int, page, pageSize int) (message.LetterTplData, error) {

	sceneVal := 0
	length := 3
	if len(scene) > 0 {
		sceneVal, _ = strconv.Atoi(scene)
		length = 0
	}

	d, f, err := message.EsTplList(meta.ES, meta.EsPrefix, keyword, title, state, module, page, pageSize, sceneVal, length)
	if err != nil {
		return d, pushLog(err, f)
	}

	return d, nil
}

// 站内信通知模板添加
func MessageSystemTplInsert(tpl message.LetterTpl) error {

	f, err := message.EsTplInsert(meta.ES, meta.EsPrefix, tpl)
	if err != nil {
		return pushLog(err, f)
	}

	return nil
}

//  站内信通知模板更新
func MessageSystemTplUpdate(id string, ty int8, tpl message.LetterTpl) error {

	data := map[string]string{}

	if ty == message.MessageSystemTplInfo {
		data = map[string]string{
			"content":      tpl.Content,
			"title":        tpl.Title,
			"icon":         tpl.Icon,
			"updated_at":   fmt.Sprintf("%d", tpl.UpdatedAt),
			"updated_name": tpl.UpdatedName,
			"updated_uid":  tpl.UpdatedUid,
		}
	}

	if ty == message.MessageSystemTplState {
		data = map[string]string{
			"state":        fmt.Sprintf("%d", tpl.State),
			"updated_at":   fmt.Sprintf("%d", tpl.UpdatedAt),
			"updated_name": tpl.UpdatedName,
			"updated_uid":  tpl.UpdatedUid,
		}
	}

	f, err := message.EsTplUpdate(meta.ES, ty, meta.EsPrefix, id, data)
	if err != nil {
		return pushLog(err, f)
	}

	return nil
}

//  站内信通知模板删除
func MessageSystemTplDelete(ids []string) error {

	f, err := message.EsTplDelete(meta.ES, meta.EsPrefix, ids)
	if err != nil {
		return pushLog(err, f)
	}

	return nil
}
