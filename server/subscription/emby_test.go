package subscription

import "testing"

func TestBangumiLayoutMainEpisode(t *testing.T) {
	got := BangumiLayout("间谍过家家", "第一季", "1", "任务开始", "main", 0, WebDLMeta{
		OriginName: "SPY×FAMILY",
		Year:       2022,
		VideoTag:   "1080p",
		VideoCodec: "H265",
		AudioCodec: "AAC",
	})
	if got.Season != 1 || got.Episode != 1 {
		t.Fatalf("season/episode = %d/%d", got.Season, got.Episode)
	}
	if got.RelDir != "间谍过家家/Season 01" {
		t.Fatalf("relDir = %s", got.RelDir)
	}
	if got.FileStem != "间谍过家家.SPYxFAMILY.S01E01.2022.1080p.WEB-DL.H265.AAC.bili" {
		t.Fatalf("fileStem = %s", got.FileStem)
	}
}

func TestBangumiLayoutNumericSeasonTitle(t *testing.T) {
	got := BangumiLayout("Demo", "第2季", "12", "", "main", 0, WebDLMeta{Year: 2024, VideoTag: "720p"})
	if got.Season != 2 || got.Episode != 12 {
		t.Fatalf("season/episode = %d/%d", got.Season, got.Episode)
	}
	if got.FileStem != "Demo.S02E12.2024.720p.WEB-DL.H265.AAC.bili" {
		t.Fatalf("fileStem = %s", got.FileStem)
	}
}

func TestBangumiLayoutExtraSection(t *testing.T) {
	got := BangumiLayout("间谍过家家", "第一季", "PV", "预告", "section_0", 0, WebDLMeta{})
	if got.Season != 0 || got.Episode != 1 {
		t.Fatalf("season/episode = %d/%d", got.Season, got.Episode)
	}
	if got.RelDir != "间谍过家家/Season 00" {
		t.Fatalf("relDir = %s", got.RelDir)
	}
}

func TestWebDLFileStemUserTemplate(t *testing.T) {
	got := WebDLFileStem("狄仁杰之狼人劫", WebDLMeta{
		OriginName: "The Frontier Under The Blood Moon",
		Year:       2026,
		VideoTag:   "2160p",
		VideoCodec: "H265",
		AudioCodec: "AAC",
	}, 1, 1)
	want := "狄仁杰之狼人劫.The.Frontier.Under.The.Blood.Moon.S01E01.2026.2160p.WEB-DL.H265.AAC.bili"
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestExtractEnglish(t *testing.T) {
	if got := ExtractEnglish("SPY×FAMILY", "间谍过家家"); got != "SPY×FAMILY" {
		t.Fatalf("got %s", got)
	}
	if got := DottedEnglish("The Frontier Under The Blood Moon"); got != "The.Frontier.Under.The.Blood.Moon" {
		t.Fatalf("got %s", got)
	}
}

func TestParseSeasonAndEpisode(t *testing.T) {
	if n := ParseSeasonNumber("第三季", "main"); n != 3 {
		t.Fatalf("第三季 got %d", n)
	}
	if n := ParseSeasonNumber("Season 4", "main"); n != 4 {
		t.Fatalf("Season 4 got %d", n)
	}
	if n := ParseEpisodeNumber("第08话", 0); n != 8 {
		t.Fatalf("第08话 got %d", n)
	}
	if n := ParseEpisodeNumber("SP", 2); n != 3 {
		t.Fatalf("SP index fallback got %d", n)
	}
}

func TestParseEmbyLayout(t *testing.T) {
	show, season, episode, epTitle := ParseEmbyLayout("间谍过家家/Season 01", "间谍过家家.SPYxFAMILY.S01E03.2022.1080p.WEB-DL.H265.AAC.bili")
	if show != "间谍过家家" || season != 1 || episode != 3 {
		t.Fatalf("got %s S%dE%d %s", show, season, episode, epTitle)
	}
}
