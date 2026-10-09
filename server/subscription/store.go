package subscription

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"bilidown/util"
)

const DefaultCron = "0 8 * * *"

type Source struct {
	ID              int64  `json:"id"`
	Type            string `json:"type"`
	Mid             string `json:"mid"`
	ResourceID      string `json:"resourceId"`
	Title           string `json:"title"`
	Cover           string `json:"cover"`
	Owner           string `json:"owner"`
	Enabled         bool   `json:"enabled"`
	DownloadType    string `json:"downloadType"`
	Format          int    `json:"format"`
	Cron            string   `json:"cron"`
	SectionIDs      []string `json:"sectionIds"`
	LastCheckAt     string   `json:"lastCheckAt"`
	LastCheckStatus string `json:"lastCheckStatus"`
	LastError       string `json:"lastError"`
	CreatedAt       string `json:"createdAt"`
	ItemCount       int    `json:"itemCount"`
}

type Item struct {
	ID             int64
	SubscriptionID int64
	ItemKey        string
	Bvid           string
	Cid            int
	Title          string
	PublishedAt    int64
	TaskID         int64
	NfoError       string
}

func InitTables(db *sql.DB) error {
	util.SqliteLock.Lock()
	defer util.SqliteLock.Unlock()
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS "subscription" (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"type" TEXT NOT NULL,
		"mid" TEXT NOT NULL DEFAULT '',
		"resource_id" TEXT NOT NULL,
		"title" TEXT NOT NULL,
		"cover" TEXT NOT NULL DEFAULT '',
		"owner" TEXT NOT NULL DEFAULT '',
		"enabled" INTEGER NOT NULL DEFAULT 1,
		"download_type" TEXT NOT NULL DEFAULT 'merge',
		"format" INTEGER NOT NULL DEFAULT 80,
		"cron" TEXT NOT NULL DEFAULT '0 8 * * *',
		"section_ids" TEXT NOT NULL DEFAULT '',
		"last_check_at" TEXT NOT NULL DEFAULT '',
		"last_check_status" TEXT NOT NULL DEFAULT '',
		"last_error" TEXT NOT NULL DEFAULT '',
		"created_at" TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE("type", "resource_id")
	)`); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS "subscription_item" (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"subscription_id" INTEGER NOT NULL,
		"item_key" TEXT NOT NULL,
		"bvid" TEXT NOT NULL DEFAULT '',
		"cid" INTEGER NOT NULL DEFAULT 0,
		"title" TEXT NOT NULL DEFAULT '',
		"published_at" INTEGER NOT NULL DEFAULT 0,
		"task_id" INTEGER NOT NULL DEFAULT 0,
		"nfo_error" TEXT NOT NULL DEFAULT '',
		"created_at" TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE("subscription_id", "item_key"),
		FOREIGN KEY("subscription_id") REFERENCES "subscription"("id") ON DELETE CASCADE
	)`); err != nil {
		return err
	}
	_, _ = db.Exec(`ALTER TABLE "subscription" ADD COLUMN "cron" TEXT NOT NULL DEFAULT '0 8 * * *'`)
	_, _ = db.Exec(`ALTER TABLE "subscription" ADD COLUMN "section_ids" TEXT NOT NULL DEFAULT ''`)
	return nil
}

func encodeSectionIDs(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	raw, err := json.Marshal(ids)
	if err != nil {
		return ""
	}
	return string(raw)
}

func decodeSectionIDs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var ids []string
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return nil
	}
	return ids
}

func SourceCron(expr string) string {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return DefaultCron
	}
	return expr
}

func ListSources(db *sql.DB) ([]Source, error) {
	util.SqliteLock.Lock()
	rows, err := db.Query(`SELECT
		s."id", s."type", s."mid", s."resource_id", s."title", s."cover", s."owner",
		s."enabled", s."download_type", s."format", s."cron", s."section_ids", s."last_check_at", s."last_check_status",
		s."last_error", s."created_at",
		(SELECT COUNT(*) FROM "subscription_item" i WHERE i."subscription_id" = s."id")
	FROM "subscription" s ORDER BY s."id" DESC`)
	util.SqliteLock.Unlock()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []Source{}
	for rows.Next() {
		var src Source
		var enabled int
		var sectionIDs string
		if err := rows.Scan(
			&src.ID, &src.Type, &src.Mid, &src.ResourceID, &src.Title, &src.Cover, &src.Owner,
			&enabled, &src.DownloadType, &src.Format, &src.Cron, &sectionIDs, &src.LastCheckAt, &src.LastCheckStatus,
			&src.LastError, &src.CreatedAt, &src.ItemCount,
		); err != nil {
			return nil, err
		}
		src.Enabled = enabled == 1
		src.SectionIDs = decodeSectionIDs(sectionIDs)
		list = append(list, src)
	}
	return list, nil
}

func GetEnabledSources(db *sql.DB) ([]Source, error) {
	all, err := ListSources(db)
	if err != nil {
		return nil, err
	}
	list := []Source{}
	for _, src := range all {
		if src.Enabled {
			list = append(list, src)
		}
	}
	return list, nil
}

func CreateSource(db *sql.DB, src *Source) error {
	if src.DownloadType == "" {
		src.DownloadType = "merge"
	}
	if src.Format == 0 {
		src.Format = 80
	}
	src.Cron = SourceCron(src.Cron)
	enabled := 0
	if src.Enabled {
		enabled = 1
	}
	util.SqliteLock.Lock()
	result, err := db.Exec(`INSERT INTO "subscription" (
		"type", "mid", "resource_id", "title", "cover", "owner", "enabled", "download_type", "format", "cron", "section_ids"
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		src.Type, src.Mid, src.ResourceID, src.Title, src.Cover, src.Owner, enabled, src.DownloadType, src.Format, src.Cron, encodeSectionIDs(src.SectionIDs),
	)
	util.SqliteLock.Unlock()
	if err != nil {
		if isUniqueError(err) {
			return errors.New("该订阅源已存在")
		}
		return err
	}
	src.ID, err = result.LastInsertId()
	return err
}

func UpdateSource(db *sql.DB, id int64, enabled *bool, downloadType string, format int, cron string) error {
	src, err := GetSource(db, id)
	if err != nil {
		return err
	}
	if enabled != nil {
		src.Enabled = *enabled
	}
	if downloadType != "" {
		src.DownloadType = downloadType
	}
	if format != 0 {
		src.Format = format
	}
	if cron != "" {
		src.Cron = SourceCron(cron)
	}
	flag := 0
	if src.Enabled {
		flag = 1
	}
	util.SqliteLock.Lock()
	_, err = db.Exec(`UPDATE "subscription" SET "enabled" = ?, "download_type" = ?, "format" = ?, "cron" = ? WHERE "id" = ?`,
		flag, src.DownloadType, src.Format, src.Cron, id)
	util.SqliteLock.Unlock()
	return err
}

func DeleteSource(db *sql.DB, id int64) error {
	util.SqliteLock.Lock()
	defer util.SqliteLock.Unlock()
	if _, err := db.Exec(`DELETE FROM "subscription_item" WHERE "subscription_id" = ?`, id); err != nil {
		return err
	}
	_, err := db.Exec(`DELETE FROM "subscription" WHERE "id" = ?`, id)
	return err
}

func GetSource(db *sql.DB, id int64) (*Source, error) {
	util.SqliteLock.Lock()
	row := db.QueryRow(`SELECT
		"id", "type", "mid", "resource_id", "title", "cover", "owner",
		"enabled", "download_type", "format", "cron", "section_ids", "last_check_at", "last_check_status",
		"last_error", "created_at"
	FROM "subscription" WHERE "id" = ?`, id)
	util.SqliteLock.Unlock()
	var src Source
	var enabled int
	var sectionIDs string
	err := row.Scan(
		&src.ID, &src.Type, &src.Mid, &src.ResourceID, &src.Title, &src.Cover, &src.Owner,
		&enabled, &src.DownloadType, &src.Format, &src.Cron, &sectionIDs, &src.LastCheckAt, &src.LastCheckStatus,
		&src.LastError, &src.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	src.Enabled = enabled == 1
	src.SectionIDs = decodeSectionIDs(sectionIDs)
	return &src, nil
}

func UpdateCheckResult(db *sql.DB, id int64, status, lastError string) error {
	util.SqliteLock.Lock()
	_, err := db.Exec(`UPDATE "subscription" SET "last_check_at" = ?, "last_check_status" = ?, "last_error" = ? WHERE "id" = ?`,
		FormatNow(), status, lastError, id)
	util.SqliteLock.Unlock()
	return err
}

func ItemExists(db *sql.DB, subscriptionID int64, itemKey string) (bool, error) {
	var count int
	util.SqliteLock.Lock()
	err := db.QueryRow(`SELECT COUNT(*) FROM "subscription_item" WHERE "subscription_id" = ? AND "item_key" = ?`,
		subscriptionID, itemKey).Scan(&count)
	util.SqliteLock.Unlock()
	return count > 0, err
}

func InsertItem(db *sql.DB, item *Item) error {
	util.SqliteLock.Lock()
	result, err := db.Exec(`INSERT INTO "subscription_item" (
		"subscription_id", "item_key", "bvid", "cid", "title", "published_at", "task_id", "nfo_error"
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		item.SubscriptionID, item.ItemKey, item.Bvid, item.Cid, item.Title, item.PublishedAt, item.TaskID, item.NfoError,
	)
	util.SqliteLock.Unlock()
	if err != nil {
		return err
	}
	item.ID, err = result.LastInsertId()
	return err
}

func UpdateItemTask(db *sql.DB, id, taskID int64) error {
	util.SqliteLock.Lock()
	_, err := db.Exec(`UPDATE "subscription_item" SET "task_id" = ? WHERE "id" = ?`, taskID, id)
	util.SqliteLock.Unlock()
	return err
}

func UpdateItemNfoError(db *sql.DB, id int64, nfoError string) error {
	util.SqliteLock.Lock()
	_, err := db.Exec(`UPDATE "subscription_item" SET "nfo_error" = ? WHERE "id" = ?`, nfoError, id)
	util.SqliteLock.Unlock()
	return err
}

func GetItemByTaskID(db *sql.DB, taskID int64) (*Item, error) {
	util.SqliteLock.Lock()
	row := db.QueryRow(`SELECT "id", "subscription_id", "item_key", "bvid", "cid", "title", "published_at", "task_id", "nfo_error"
		FROM "subscription_item" WHERE "task_id" = ?`, taskID)
	util.SqliteLock.Unlock()
	var item Item
	err := row.Scan(&item.ID, &item.SubscriptionID, &item.ItemKey, &item.Bvid, &item.Cid, &item.Title, &item.PublishedAt, &item.TaskID, &item.NfoError)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func isUniqueError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE") || strings.Contains(msg, "unique")
}
