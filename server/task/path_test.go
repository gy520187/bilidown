package task

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFilePathLegacy(t *testing.T) {
	task := TaskInDB{TaskInitOption: TaskInitOption{Folder: "/dl", Title: "hello", DownloadType: "merge"}, ID: 1}
	got := task.FilePath()
	if !strings.HasPrefix(got, filepath.Join("/dl", "hello ")) || !strings.HasSuffix(got, ".mp4") {
		t.Fatalf("got %s", got)
	}
}

func TestFilePathEmby(t *testing.T) {
	task := TaskInDB{TaskInitOption: TaskInitOption{
		Folder:       "/dl",
		Title:        "间谍过家家.SPYxFAMILY.S01E01.2022.1080p.WEB-DL.H265.AAC.bili",
		DownloadType: "merge",
		RelDir:       "间谍过家家/Season 01",
	}, ID: 9}
	got := task.FilePath()
	want := filepath.Join("/dl", "间谍过家家", "Season 01", "间谍过家家.SPYxFAMILY.S01E01.2022.1080p.WEB-DL.H265.AAC.bili.mp4")
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestFilePathSplitsBakedRelDir(t *testing.T) {
	task := TaskInDB{TaskInitOption: TaskInitOption{
		Folder:       "/dl",
		Title:        "闪闪的儿科医生/Season 01/闪闪的儿科医生.The.GLorious.Pediatricians.S04E10.2023.2160p.WEB-DL.H265.AAC.bili",
		DownloadType: "merge",
	}, ID: 3}
	got := task.FilePath()
	want := filepath.Join("/dl", "闪闪的儿科医生", "Season 01", "闪闪的儿科医生.The.GLorious.Pediatricians.S04E10.2023.2160p.WEB-DL.H265.AAC.bili.mp4")
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}
