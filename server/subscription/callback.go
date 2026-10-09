package subscription

import (
	"fmt"
	"log"
	"strings"

	"bilidown/task"
	"bilidown/util"
)

func WriteTaskNFO(t *task.Task) {
	if t == nil || t.ID == 0 {
		return
	}
	mediaPath := t.FilePath()
	if strings.TrimSpace(t.RelDir) != "" {
		show, season, episode, epTitle := ParseEmbyLayout(t.RelDir, t.Title)
		title := epTitle
		if title == "" && episode > 0 {
			title = fmt.Sprintf("第 %d 集", episode)
		}
		if title == "" {
			title = t.Title
		}
		data := NFOData{
			Title:     title,
			ShowTitle: show,
			Owner:     t.Owner,
			BVID:      t.Bvid,
			Cover:     t.Cover,
			Season:    season,
			Episode:   episode,
		}
		if item := itemByTask(t.ID); item != nil {
			data.PublishedAt = item.PublishedAt
		}
		if err := WriteEpisodeNFO(mediaPath, data); err != nil {
			recordNfoError(t.ID, err.Error())
			return
		}
		tv := data
		tv.Title = show
		tv.BVID = ""
		if err := WriteTVShowNFO(TVShowNFOPath(mediaPath), tv); err != nil {
			recordNfoError(t.ID, err.Error())
		}
		return
	}
	item := itemByTask(t.ID)
	if item == nil {
		return
	}
	data := NFOData{
		Title:       t.Title,
		Owner:       t.Owner,
		BVID:        t.Bvid,
		Cover:       t.Cover,
		PublishedAt: item.PublishedAt,
	}
	if err := WriteNFO(mediaPath, data); err != nil {
		recordNfoError(t.ID, err.Error())
	}
}

func itemByTask(taskID int64) *Item {
	db := util.MustGetDB()
	defer db.Close()
	item, err := GetItemByTaskID(db, taskID)
	if err != nil {
		return nil
	}
	return item
}

func recordNfoError(taskID int64, msg string) {
	db := util.MustGetDB()
	defer db.Close()
	item, err := GetItemByTaskID(db, taskID)
	if err != nil {
		return
	}
	if updateErr := UpdateItemNfoError(db, item.ID, msg); updateErr != nil {
		log.Printf("subscription nfo: %v", updateErr)
	}
}
