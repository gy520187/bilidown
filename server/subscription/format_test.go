package subscription

import (
	"testing"

	"bilidown/bilibili"
	"bilidown/common"
)

func TestSelectFormatExact(t *testing.T) {
	got := SelectFormat(80, []int{16, 32, 64, 80, 112})
	if got != 80 {
		t.Fatalf("got %d", got)
	}
}

func TestSelectFormatFallback(t *testing.T) {
	got := SelectFormat(80, []int{16, 32, 64})
	if got != 64 {
		t.Fatalf("got %d", got)
	}
}

func TestSelectFormatBelowAll(t *testing.T) {
	got := SelectFormat(6, []int{16, 32, 64})
	if got != 16 {
		t.Fatalf("got %d", got)
	}
}

func TestShouldCreateTaskFirstRun(t *testing.T) {
	now := int64(1_700_000_000)
	if !ShouldCreateTask(true, now-3600, now) {
		t.Fatal("recent item should create on first run")
	}
	if ShouldCreateTask(true, now-FirstRunWindowSeconds-1, now) {
		t.Fatal("old item should skip on first run")
	}
	if !ShouldCreateTask(false, now-FirstRunWindowSeconds-1, now) {
		t.Fatal("old unseen item should create on later run")
	}
}

func TestSelectVideoPrefersHEVC(t *testing.T) {
	videos := []bilibili.Media{
		{ID: common.MediaFormat(80), Codecid: 7, BaseURL: "avc"},
		{ID: common.MediaFormat(80), Codecid: 12, BaseURL: "hevc"},
		{ID: common.MediaFormat(64), Codecid: 12, BaseURL: "hevc64"},
	}
	url, quality, ok := SelectVideo(videos, 80, 12)
	if !ok || quality != 80 || url != "hevc" {
		t.Fatalf("url=%s quality=%d ok=%v", url, quality, ok)
	}
}

func TestVideoAndCodecTags(t *testing.T) {
	if got := VideoTag(120, 3840, 2160, 12, "hev1.1.6.L153.90"); got != "2160p" {
		t.Fatalf("2160p got %s", got)
	}
	if got := VideoTag(80, 0, 0, 12, ""); got != "1080p" {
		t.Fatalf("quality fallback got %s", got)
	}
	if got := CodecTag(12, "hev1"); got != "H265" {
		t.Fatalf("hevc got %s", got)
	}
	if got := CodecTag(7, "avc1"); got != "H264" {
		t.Fatalf("avc got %s", got)
	}
	if got := AudioTag(true, "flac", true); got != "FLAC" {
		t.Fatalf("flac got %s", got)
	}
	if got := AudioTag(false, "mp4a.40.2", false); got != "AAC" {
		t.Fatalf("aac got %s", got)
	}
}

func TestNFOPathAndRender(t *testing.T) {
	path := NFOPath("/tmp/demo.mp4")
	if path != "/tmp/demo.nfo" {
		t.Fatalf("unexpected path %s", path)
	}
	body, err := RenderNFO(NFOData{
		Title:       "标题",
		Owner:       "UP",
		BVID:        "BV1xx",
		Cover:       "https://example.com/cover.jpg",
		PublishedAt: 1700000000,
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !containsAll(text, "标题", "UP", "BV1xx", "https://example.com/cover.jpg", "2023-11-14") {
		t.Fatalf("missing nfo fields: %s", text)
	}
	if !stringContains(text, "<movie>") {
		t.Fatalf("movie nfo missing root: %s", text)
	}
}

func TestRenderEpisodeAndTVShowNFO(t *testing.T) {
	body, err := RenderEpisodeNFO(NFOData{
		Title:       "任务开始",
		ShowTitle:   "间谍过家家",
		Owner:       "制作委员会",
		BVID:        "BV1xx",
		Cover:       "https://example.com/cover.jpg",
		PublishedAt: 1700000000,
		Season:      1,
		Episode:     1,
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !containsAll(text, "<episodedetails>", "任务开始", "间谍过家家", "<season>1</season>", "<episode>1</episode>", "BV1xx", "2023-11-14") {
		t.Fatalf("missing episode nfo fields: %s", text)
	}
	show, err := RenderTVShowNFO(NFOData{ShowTitle: "间谍过家家", Owner: "制作委员会", Cover: "https://example.com/cover.jpg"})
	if err != nil {
		t.Fatal(err)
	}
	if !containsAll(string(show), "<tvshow>", "间谍过家家", "制作委员会") {
		t.Fatalf("missing tvshow nfo fields: %s", show)
	}
	if p := TVShowNFOPath("/tmp/show/Season 01/ep.mp4"); p != "/tmp/show/tvshow.nfo" {
		t.Fatalf("tvshow path = %s", p)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !stringContains(s, p) {
			return false
		}
	}
	return true
}

func stringContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
