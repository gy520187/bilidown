package util

import "testing"

func TestSanitizeRelDir(t *testing.T) {
	if got := SanitizeRelDir("间谍过家家/Season 01"); got != "间谍过家家/Season 01" {
		t.Fatalf("got %s", got)
	}
	if got := SanitizeRelDir("../a/./b/c"); got != "a/b/c" {
		t.Fatalf("got %s", got)
	}
	if got := SanitizeRelDir("a/../b"); got != "a/b" {
		t.Fatalf("dotdot dropped, got %s", got)
	}
}

func TestSplitOutRelDir(t *testing.T) {
	title, rel := SplitOutRelDir("闪闪的儿科医生.S01E01.2023.1080p.WEB-DL.H265.AAC.bili", "闪闪的儿科医生/Season 01")
	if title != "闪闪的儿科医生.S01E01.2023.1080p.WEB-DL.H265.AAC.bili" || rel != "闪闪的儿科医生/Season 01" {
		t.Fatalf("title=%s rel=%s", title, rel)
	}
	title, rel = SplitOutRelDir("闪闪的儿科医生/Season 01/闪闪的儿科医生.S01E01.2023.1080p.WEB-DL.H265.AAC.bili", "")
	if title != "闪闪的儿科医生.S01E01.2023.1080p.WEB-DL.H265.AAC.bili" || rel != "闪闪的儿科医生/Season 01" {
		t.Fatalf("baked path title=%s rel=%s", title, rel)
	}
}
