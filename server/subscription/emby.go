package subscription

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"bilidown/bilibili"
	"bilidown/util"
)

type EmbyLayout struct {
	Show     string
	Season   int
	Episode  int
	EpTitle  string
	RelDir   string
	FileStem string
}

type WebDLMeta struct {
	OriginName string
	Evaluate   string
	Year       int
	VideoTag   string
	VideoCodec string
	AudioCodec string
}

func BangumiLayout(showTitle, seasonTitle, episodeTitle, longTitle, sectionID string, index int, meta WebDLMeta) EmbyLayout {
	show := strings.TrimSpace(showTitle)
	if show == "" {
		show = "未命名番剧"
	}
	show = util.FilterFileName(show)
	season := ParseSeasonNumber(seasonTitle, sectionID)
	episode := ParseEpisodeNumber(episodeTitle, index)
	epTitle := episodeDisplayTitle(longTitle, episodeTitle)
	stem := WebDLFileStem(show, meta, season, episode)
	return EmbyLayout{
		Show:     show,
		Season:   season,
		Episode:  episode,
		EpTitle:  epTitle,
		RelDir:   util.SanitizeRelDir(path.Join(show, fmt.Sprintf("Season %02d", season))),
		FileStem: stem,
	}
}

func LayoutFromArchive(src *Source, item bilibili.ArchiveItem, meta WebDLMeta) EmbyLayout {
	show := item.ShowTitle
	if show == "" && src != nil {
		show = src.Title
	}
	if meta.OriginName == "" {
		meta.OriginName = item.OriginName
	}
	if meta.Evaluate == "" {
		meta.Evaluate = item.Evaluate
	}
	if meta.Year == 0 {
		meta.Year = item.Year
	}
	return BangumiLayout(show, item.SeasonTitle, item.EpisodeTitle, item.Title, item.SectionID, item.Index, meta)
}

func WebDLFileStem(show string, meta WebDLMeta, season, episode int) string {
	zh := compactZh(show)
	if zh == "" {
		zh = "未命名番剧"
	}
	en := DottedEnglish(ExtractEnglish(meta.OriginName, show, meta.Evaluate))
	parts := []string{zh}
	if en != "" && !strings.EqualFold(en, zh) {
		parts = append(parts, en)
	}
	parts = append(parts, fmt.Sprintf("S%02dE%02d", season, episode))
	if meta.Year > 0 {
		parts = append(parts, strconv.Itoa(meta.Year))
	}
	videoTag := strings.TrimSpace(meta.VideoTag)
	if videoTag == "" {
		videoTag = "1080p"
	}
	vcodec := strings.TrimSpace(meta.VideoCodec)
	if vcodec == "" {
		vcodec = "H265"
	}
	acodec := strings.TrimSpace(meta.AudioCodec)
	if acodec == "" {
		acodec = "AAC"
	}
	parts = append(parts, videoTag, "WEB-DL", vcodec, acodec, "bili")
	return util.FilterFileName(strings.Join(parts, "."))
}

func compactZh(s string) string {
	s = util.FilterFileName(strings.TrimSpace(s))
	return strings.Join(strings.Fields(s), "")
}

func ExtractEnglish(values ...string) string {
	for _, s := range values {
		if en := englishFrom(s); en != "" {
			return en
		}
	}
	return ""
}

func DottedEnglish(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	repl := strings.NewReplacer("×", "x", "·", ".", "—", " ", "–", " ", "-", " ", "_", " ", "/", " ")
	s = repl.Replace(s)
	var b strings.Builder
	lastDot := false
	for _, r := range s {
		if (unicode.IsLetter(r) && r < 128) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDot = false
			continue
		}
		if unicode.IsSpace(r) || r == '.' {
			if b.Len() > 0 && !lastDot {
				b.WriteByte('.')
				lastDot = true
			}
		}
	}
	return strings.Trim(b.String(), ".")
}

func englishFrom(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if isMostlyASCII(s) {
		return s
	}
	reParen := regexp.MustCompile(`[\(（]([^\)）]{2,})[\)）]`)
	if m := reParen.FindStringSubmatch(s); len(m) == 2 && isMostlyASCII(m[1]) {
		return m[1]
	}
	reRun := regexp.MustCompile(`[A-Za-z][A-Za-z0-9][A-Za-z0-9 .:_'\-]{2,}`)
	best := ""
	for _, m := range reRun.FindAllString(s, -1) {
		if len(m) > len(best) && isMostlyASCII(m) {
			best = m
		}
	}
	return strings.TrimSpace(best)
}

func isReleaseTagRest(s string) bool {
	u := strings.ToUpper(s)
	return strings.Contains(u, "WEB-DL") ||
		strings.Contains(u, "BILI") ||
		regexp.MustCompile(`(?i)^\d{3,4}P`).MatchString(s)
}

func isMostlyASCII(s string) bool {
	letters := 0
	ascii := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			letters++
			if r < 128 {
				ascii++
			}
		}
	}
	return letters >= 2 && ascii*2 >= letters
}

func ParseSeasonNumber(seasonTitle, sectionID string) int {
	if sectionID != "" && sectionID != "main" {
		return 0
	}
	s := strings.TrimSpace(seasonTitle)
	if s == "" {
		return 1
	}
	if n := chineseSeasonNumber(s); n > 0 {
		return n
	}
	re := regexp.MustCompile(`(\d+)`)
	if m := re.FindString(s); m != "" {
		n, err := strconv.Atoi(m)
		if err == nil && n >= 0 {
			return n
		}
	}
	return 1
}

func ParseEpisodeNumber(episodeTitle string, index int) int {
	s := strings.TrimSpace(episodeTitle)
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	re := regexp.MustCompile(`(\d+)`)
	if m := re.FindString(s); m != "" {
		n, err := strconv.Atoi(m)
		if err == nil && n > 0 {
			return n
		}
	}
	if index >= 0 {
		return index + 1
	}
	return 1
}

func ParseEmbyLayout(relDir, fileStem string) (show string, season, episode int, epTitle string) {
	relDir = strings.ReplaceAll(relDir, "\\", "/")
	parts := strings.Split(relDir, "/")
	if len(parts) > 0 {
		show = parts[0]
	}
	reSeason := regexp.MustCompile(`(?i)season\s*(\d+)`)
	if m := reSeason.FindStringSubmatch(relDir); len(m) == 2 {
		season, _ = strconv.Atoi(m[1])
	}
	reEP := regexp.MustCompile(`(?i)S(\d+)E(\d+)`)
	if m := reEP.FindStringSubmatch(fileStem); len(m) == 3 {
		season, _ = strconv.Atoi(m[1])
		episode, _ = strconv.Atoi(m[2])
		if i := strings.Index(fileStem, m[0]); i >= 0 {
			rest := strings.TrimSpace(strings.TrimPrefix(fileStem[i+len(m[0]):], " - "))
			rest = strings.Trim(rest, ". ")
			if rest != "" && !isReleaseTagRest(rest) {
				epTitle = rest
			}
		}
	}
	return
}

func episodeDisplayTitle(longTitle, episodeTitle string) string {
	name := strings.TrimSpace(longTitle)
	if name == "" {
		name = strings.TrimSpace(episodeTitle)
	}
	if name == "" {
		return ""
	}
	if _, err := strconv.Atoi(name); err == nil {
		return ""
	}
	if episodeTitle != "" && name == episodeTitle {
		if _, err := strconv.Atoi(strings.TrimSpace(episodeTitle)); err == nil {
			return ""
		}
	}
	return util.FilterFileName(name)
}

func chineseSeasonNumber(s string) int {
	start := strings.Index(s, "第")
	end := strings.Index(s, "季")
	if start < 0 || end <= start {
		return 0
	}
	return chineseNumeral(s[start+len("第") : end])
}

func chineseNumeral(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	digits := map[rune]int{
		'零': 0, '〇': 0, '一': 1, '二': 2, '两': 2, '三': 3, '四': 4,
		'五': 5, '六': 6, '七': 7, '八': 8, '九': 9, '十': 10,
	}
	n := 0
	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}
		d, ok := digits[r]
		if !ok {
			return 0
		}
		if d == 10 {
			if n == 0 {
				n = 10
			} else {
				n = n * 10
			}
			continue
		}
		if n >= 10 {
			n += d
		} else {
			n = d
		}
	}
	return n
}
