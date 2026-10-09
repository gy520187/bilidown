package router

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bilidown/bilibili"
	"bilidown/subscription"
	"bilidown/util"
	"bilidown/util/res_error"
)

func requireLoginClient(w http.ResponseWriter) (*bilibili.BiliClient, bool) {
	db := util.MustGetDB()
	defer db.Close()
	sessdata, err := bilibili.GetSessdata(db)
	if err != nil || sessdata == "" {
		res_error.Send(w, res_error.NotLogin)
		return nil, false
	}
	client := &bilibili.BiliClient{SESSDATA: sessdata}
	ok, err := client.CheckLogin()
	if err != nil || !ok {
		res_error.Send(w, res_error.NotLogin)
		return nil, false
	}
	return client, true
}

func getSubscriptions(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireLoginClient(w); !ok {
		return
	}
	db := util.MustGetDB()
	defer db.Close()
	list, err := subscription.ListSources(db)
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	util.Res{Success: true, Message: "获取成功", Data: list}.Write(w)
}

type createSubscriptionBody struct {
	URL          string   `json:"url"`
	DownloadType string   `json:"downloadType"`
	Format       int      `json:"format"`
	Cron         string   `json:"cron"`
	SectionIDs   []string `json:"sectionIds"`
}

func previewSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		res_error.Send(w, res_error.MethodNotAllowError)
		return
	}
	client, ok := requireLoginClient(w)
	if !ok {
		return
	}
	defer r.Body.Close()
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.URL == "" {
		res_error.Send(w, res_error.ParamError)
		return
	}
	parsed, err := subscription.ParseSubscriptionURL(body.URL)
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	preview, err := subscription.ResolvePreview(client, parsed)
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	util.Res{Success: true, Message: "解析成功", Data: preview}.Write(w)
}

func createSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		res_error.Send(w, res_error.MethodNotAllowError)
		return
	}
	client, ok := requireLoginClient(w)
	if !ok {
		return
	}
	defer r.Body.Close()
	var body createSubscriptionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		res_error.Send(w, res_error.ParamError)
		return
	}
	parsed, err := subscription.ParseSubscriptionURL(body.URL)
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	preview, err := subscription.ResolvePreview(client, parsed)
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	src := preview.Source
	if parsed.Type == subscription.TypeBangumi {
		src.SectionIDs = body.SectionIDs
		if len(src.SectionIDs) == 0 {
			src.SectionIDs = []string{"main"}
		}
	}
	if body.DownloadType != "" {
		src.DownloadType = body.DownloadType
	}
	if body.Format != 0 {
		src.Format = body.Format
	}
	if body.Cron != "" {
		if err := subscription.ValidateCron(body.Cron); err != nil {
			util.Res{Success: false, Message: "cron 表达式格式错误"}.Write(w)
			return
		}
		src.Cron = body.Cron
	}
	db := util.MustGetDB()
	defer db.Close()
	if err := subscription.CreateSource(db, src); err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	subscription.SyncSource(*src)
	go subscription.RunCheck(src.ID)
	util.Res{Success: true, Message: "创建成功", Data: src}.Write(w)
}

type updateSubscriptionBody struct {
	ID           int64  `json:"id"`
	Enabled      *bool  `json:"enabled"`
	DownloadType string `json:"downloadType"`
	Format       int    `json:"format"`
	Cron         string `json:"cron"`
}

func updateSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		res_error.Send(w, res_error.MethodNotAllowError)
		return
	}
	if _, ok := requireLoginClient(w); !ok {
		return
	}
	defer r.Body.Close()
	var body updateSubscriptionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == 0 {
		res_error.Send(w, res_error.ParamError)
		return
	}
	if body.Cron != "" {
		if err := subscription.ValidateCron(body.Cron); err != nil {
			util.Res{Success: false, Message: "cron 表达式格式错误"}.Write(w)
			return
		}
	}
	db := util.MustGetDB()
	defer db.Close()
	if err := subscription.UpdateSource(db, body.ID, body.Enabled, body.DownloadType, body.Format, body.Cron); err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	src, err := subscription.GetSource(db, body.ID)
	if err == nil {
		subscription.SyncSource(*src)
	}
	util.Res{Success: true, Message: "保存成功"}.Write(w)
}

func deleteSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		res_error.Send(w, res_error.MethodNotAllowError)
		return
	}
	if _, ok := requireLoginClient(w); !ok {
		return
	}
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil || id == 0 {
		res_error.Send(w, res_error.ParamError)
		return
	}
	db := util.MustGetDB()
	defer db.Close()
	if err := subscription.DeleteSource(db, id); err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	subscription.RemoveScheduledSource(id)
	util.Res{Success: true, Message: "删除成功"}.Write(w)
}

func runSubscriptionCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		res_error.Send(w, res_error.MethodNotAllowError)
		return
	}
	if _, ok := requireLoginClient(w); !ok {
		return
	}
	var onlyID int64
	if raw := r.URL.Query().Get("id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			res_error.Send(w, res_error.ParamError)
			return
		}
		onlyID = id
	}
	go subscription.RunCheck(onlyID)
	util.Res{Success: true, Message: "已开始检查"}.Write(w)
}
