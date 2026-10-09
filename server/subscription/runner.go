package subscription

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"bilidown/bilibili"
	"bilidown/common"
	"bilidown/task"
	"bilidown/util"
)

func CheckOne(db *sql.DB, src *Source, firstRun bool) error {
	client, err := loggedClient(db)
	if err != nil {
		_ = UpdateCheckResult(db, src.ID, "need_login", err.Error())
		return err
	}
	items, meta, err := listSourceItems(client, src)
	if err != nil {
		_ = UpdateCheckResult(db, src.ID, "error", err.Error())
		return err
	}
	if meta != nil {
		_ = applyMeta(db, src, meta)
	}
	now := time.Now().Unix()
	var lastErr error
	for _, item := range items {
		exists, err := ItemExists(db, src.ID, item.ItemKey)
		if err != nil {
			lastErr = err
			continue
		}
		if exists {
			continue
		}
		record := &Item{
			SubscriptionID: src.ID,
			ItemKey:        item.ItemKey,
			Bvid:           item.Bvid,
			Cid:            item.Cid,
			Title:          item.Title,
			PublishedAt:    item.PubTime,
		}
		if err := InsertItem(db, record); err != nil {
			lastErr = err
			continue
		}
		if !ShouldCreateTask(firstRun, item.PubTime, now) {
			continue
		}
		if err := createDownloadTask(db, client, src, record, item); err != nil {
			record.NfoError = err.Error()
			_ = UpdateItemNfoError(db, record.ID, err.Error())
			lastErr = err
			continue
		}
	}
	status := "success"
	msg := ""
	if lastErr != nil {
		status = "error"
		msg = lastErr.Error()
	}
	_ = UpdateCheckResult(db, src.ID, status, msg)
	return lastErr
}

func CheckAll(db *sql.DB, onlyID int64) error {
	if _, err := loggedClient(db); err != nil {
		sources, listErr := GetEnabledSources(db)
		if listErr == nil {
			for _, src := range sources {
				if onlyID != 0 && src.ID != onlyID {
					continue
				}
				_ = UpdateCheckResult(db, src.ID, "need_login", err.Error())
			}
		}
		return err
	}
	sources, err := GetEnabledSources(db)
	if err != nil {
		return err
	}
	var lastErr error
	for _, src := range sources {
		if onlyID != 0 && src.ID != onlyID {
			continue
		}
		copySrc := src
		firstRun := src.LastCheckAt == ""
		if err := CheckOne(db, &copySrc, firstRun); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

type sourceMeta struct {
	Title string
	Cover string
	Owner string
}

func listSourceItems(client *bilibili.BiliClient, src *Source) ([]bilibili.ArchiveItem, *sourceMeta, error) {
	switch src.Type {
	case TypeSeason:
		mid, _ := strconv.Atoi(src.Mid)
		sid, _ := strconv.Atoi(src.ResourceID)
		items, err := client.ListSeasonArchives(mid, sid)
		return items, nil, err
	case TypeSeries:
		mid, _ := strconv.Atoi(src.Mid)
		sid, _ := strconv.Atoi(src.ResourceID)
		items, err := client.ListSeriesArchives(mid, sid)
		return items, nil, err
	case TypeFav:
		mediaId, _ := strconv.Atoi(src.ResourceID)
		items, err := client.ListFavArchives(mediaId)
		return items, nil, err
	case TypeSpace:
		mid, _ := strconv.Atoi(src.ResourceID)
		items, err := client.ListSpaceArchives(mid)
		return items, nil, err
	case TypeBangumi:
		ssid, _ := strconv.Atoi(src.ResourceID)
		sections, info, err := client.ListBangumiSections(0, ssid)
		if err != nil {
			return nil, nil, err
		}
		sectionIDs := src.SectionIDs
		if len(sectionIDs) == 0 {
			sectionIDs = []string{"main"}
		}
		items := bilibili.FilterBangumiSections(sections, sectionIDs)
		meta := &sourceMeta{Title: info.Title, Cover: info.Cover, Owner: info.Actors}
		return items, meta, nil
	default:
		return nil, nil, fmt.Errorf("未知订阅类型: %s", src.Type)
	}
}

func applyMeta(db *sql.DB, src *Source, meta *sourceMeta) error {
	if meta.Title != "" {
		src.Title = meta.Title
	}
	if meta.Cover != "" {
		src.Cover = meta.Cover
	}
	if meta.Owner != "" {
		src.Owner = meta.Owner
	}
	util.SqliteLock.Lock()
	_, err := db.Exec(`UPDATE "subscription" SET "title" = ?, "cover" = ?, "owner" = ? WHERE "id" = ?`,
		src.Title, src.Cover, src.Owner, src.ID)
	util.SqliteLock.Unlock()
	return err
}

type PreviewSection struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Count int    `json:"count"`
}

type Preview struct {
	Source   *Source          `json:"source"`
	Sections []PreviewSection `json:"sections"`
	Total    int              `json:"total"`
}

func ResolvePreview(client *bilibili.BiliClient, parsed ParsedSource) (*Preview, error) {
	src := &Source{
		Type:         parsed.Type,
		Mid:          parsed.Mid,
		ResourceID:   parsed.ResourceID,
		Enabled:      true,
		DownloadType: "merge",
		Format:       80,
		Cron:         DefaultCron,
	}
	preview := &Preview{Source: src, Sections: []PreviewSection{}}
	switch parsed.Type {
	case TypeBangumi:
		epid := 0
		ssid := 0
		if parsed.FromEP {
			epid = parsed.EPID
		} else {
			ssid, _ = strconv.Atoi(parsed.ResourceID)
		}
		sections, info, err := client.ListBangumiSections(epid, ssid)
		if err != nil {
			return nil, err
		}
		src.ResourceID = strconv.Itoa(info.SeasonID)
		src.Title = info.Title
		src.Cover = info.Cover
		src.Owner = info.Actors
		total := 0
		for _, section := range sections {
			preview.Sections = append(preview.Sections, PreviewSection{ID: section.ID, Title: section.Title, Count: section.Count})
			total += section.Count
			if src.Title == "" && len(section.Items) > 0 {
				src.Title = section.Items[0].Title
			}
		}
		preview.Total = total
	case TypeSeason:
		mid, _ := strconv.Atoi(parsed.Mid)
		sid, _ := strconv.Atoi(parsed.ResourceID)
		items, err := client.ListSeasonArchives(mid, sid)
		if err != nil {
			return nil, err
		}
		if err := fillFromFirst(client, src, items); err != nil {
			return nil, err
		}
		preview.Total = len(items)
	case TypeSeries:
		mid, _ := strconv.Atoi(parsed.Mid)
		sid, _ := strconv.Atoi(parsed.ResourceID)
		items, err := client.ListSeriesArchives(mid, sid)
		if err != nil {
			return nil, err
		}
		if err := fillFromFirst(client, src, items); err != nil {
			return nil, err
		}
		preview.Total = len(items)
	case TypeFav:
		mediaId, _ := strconv.Atoi(parsed.ResourceID)
		items, err := client.ListFavArchives(mediaId)
		if err != nil {
			return nil, err
		}
		if err := fillFromFirst(client, src, items); err != nil {
			return nil, err
		}
		preview.Total = len(items)
	case TypeSpace:
		mid, _ := strconv.Atoi(parsed.ResourceID)
		items, err := client.ListSpaceArchives(mid)
		if err != nil {
			return nil, err
		}
		if err := fillFromFirst(client, src, items); err != nil {
			return nil, err
		}
		if src.Owner == "" && len(items) > 0 {
			src.Owner = items[0].Owner
		}
		if src.Title == "" {
			src.Title = "UP " + parsed.ResourceID
		}
		preview.Total = len(items)
	case TypeVideo:
		season, err := sourceFromVideo(client, parsed.ResourceID)
		if err != nil {
			return nil, err
		}
		return ResolvePreview(client, season)
	default:
		return nil, errors.New("未知订阅类型")
	}
	if src.Title == "" {
		src.Title = parsed.Type + " " + parsed.ResourceID
	}
	if len(preview.Sections) == 0 {
		preview.Sections = []PreviewSection{{ID: "all", Title: "全部稿件", Count: preview.Total}}
	}
	return preview, nil
}

func sourceFromVideo(client *bilibili.BiliClient, bvid string) (ParsedSource, error) {
	info, err := client.GetVideoInfo(bvid)
	if err != nil {
		return ParsedSource{}, err
	}
	if info.UgcSeason.ID == 0 {
		return ParsedSource{}, errors.New("单个视频不能作为订阅源")
	}
	mid := info.UgcSeason.Mid
	if mid == 0 {
		mid = int64(info.Owner.Mid)
	}
	return ParsedSource{
		Type:       TypeSeason,
		Mid:        strconv.FormatInt(mid, 10),
		ResourceID: strconv.FormatInt(info.UgcSeason.ID, 10),
	}, nil
}

func fillFromFirst(client *bilibili.BiliClient, src *Source, items []bilibili.ArchiveItem) error {
	if len(items) == 0 {
		return errors.New("订阅源内容为空")
	}
	first := items[0]
	if err := client.FillArchiveDetail(&first); err != nil {
		src.Title = first.Title
		src.Cover = first.Cover
		src.Owner = first.Owner
		return nil
	}
	src.Title = first.Title
	src.Cover = first.Cover
	src.Owner = first.Owner
	return nil
}

func createDownloadTask(db *sql.DB, client *bilibili.BiliClient, src *Source, record *Item, item bilibili.ArchiveItem) error {
	if err := client.FillArchiveDetail(&item); err != nil {
		return err
	}
	playInfo, err := client.GetPlayInfo(item.Bvid, item.Cid)
	if err != nil {
		return err
	}
	if playInfo.Dash == nil {
		return errors.New("播放地址缺少 dash")
	}
	videoURL, quality, ok := SelectVideo(playInfo.Dash.Video, src.Format, 12)
	if !ok {
		return errors.New("未找到对应视频分辨率格式")
	}
	audioURL := SelectAudio(playInfo.Dash, true)
	if src.DownloadType != "video" && audioURL == "" {
		return errors.New("未找到音频地址")
	}
	if src.DownloadType == "audio" {
		videoURL = audioURL
	}
	folder, err := util.GetCurrentFolder(db)
	if err != nil {
		return err
	}
	owner := item.Owner
	if owner == "" {
		owner = src.Owner
	}
	_task := task.Task{
		TaskInDB: task.TaskInDB{
			TaskInitOption: task.TaskInitOption{
				Bvid:         item.Bvid,
				Cid:          item.Cid,
				Format:       common.MediaFormat(quality),
				Title:        util.FilterFileName(item.Title),
				Owner:        owner,
				Cover:        item.Cover,
				Status:       "waiting",
				Folder:       folder,
				Audio:        audioURL,
				Video:        videoURL,
				Duration:     item.Duration,
				DownloadType: src.DownloadType,
			},
		},
	}
	if err := _task.Create(db); err != nil {
		return err
	}
	if err := UpdateItemTask(db, record.ID, _task.ID); err != nil {
		return err
	}
	go _task.Start()
	return nil
}

func loggedClient(db *sql.DB) (*bilibili.BiliClient, error) {
	sessdata, err := bilibili.GetSessdata(db)
	if err != nil || sessdata == "" {
		return nil, errors.New("未登录")
	}
	client := &bilibili.BiliClient{SESSDATA: sessdata}
	ok, err := client.CheckLogin()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("未登录")
	}
	return client, nil
}
