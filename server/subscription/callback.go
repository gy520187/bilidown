package subscription

import (
	"log"

	"bilidown/task"
	"bilidown/util"
)

func WriteTaskNFO(t *task.Task) {
	if t == nil || t.ID == 0 {
		return
	}
	db := util.MustGetDB()
	defer db.Close()
	item, err := GetItemByTaskID(db, t.ID)
	if err != nil {
		return
	}
	data := NFOData{
		Title:       t.Title,
		Owner:       t.Owner,
		BVID:        t.Bvid,
		Cover:       t.Cover,
		PublishedAt: item.PublishedAt,
	}
	if err := WriteNFO(t.FilePath(), data); err != nil {
		if updateErr := UpdateItemNfoError(db, item.ID, err.Error()); updateErr != nil {
			log.Printf("subscription nfo: %v", updateErr)
		}
	}
}
