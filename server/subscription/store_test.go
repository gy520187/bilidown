package subscription

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestInitTablesMigratesLegacySubscription(t *testing.T) {
	db, err := sql.Open("sqlite", "file:mig?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE "subscription" ("id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT)`); err != nil {
		t.Fatal(err)
	}
	if err := InitTables(db); err != nil {
		t.Fatal(err)
	}
	src := &Source{
		Type:         TypeSeason,
		Mid:          "1",
		ResourceID:   "2",
		Title:        "test",
		Enabled:      true,
		DownloadType: "merge",
		Format:       80,
	}
	if err := CreateSource(db, src); err != nil {
		t.Fatal(err)
	}
	got, err := GetSource(db, src.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != TypeSeason || got.ResourceID != "2" {
		t.Fatalf("unexpected source: %+v", got)
	}
}
