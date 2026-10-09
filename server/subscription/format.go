package subscription

import (
	"strings"

	"bilidown/bilibili"
)

func SelectFormat(target int, available []int) int {
	if len(available) == 0 {
		return target
	}
	best := 0
	min := available[0]
	for _, q := range available {
		if q < min {
			min = q
		}
		if q <= target && q > best {
			best = q
		}
	}
	if best == 0 {
		return min
	}
	return best
}

func availableQualities(videos []bilibili.Media) []int {
	seen := map[int]struct{}{}
	var list []int
	for _, item := range videos {
		q := int(item.ID)
		if _, ok := seen[q]; ok {
			continue
		}
		seen[q] = struct{}{}
		list = append(list, q)
	}
	return list
}

type SelectedVideo struct {
	URL     string
	Quality int
	Width   int
	Height  int
	Codecid int
	Codecs  string
}

func SelectVideo(videos []bilibili.Media, target int, preferredCodec int) (url string, quality int, ok bool) {
	sel, ok := SelectVideoMedia(videos, target, preferredCodec)
	if !ok {
		return "", 0, false
	}
	return sel.URL, sel.Quality, true
}

func SelectVideoMedia(videos []bilibili.Media, target int, preferredCodec int) (SelectedVideo, bool) {
	quality := SelectFormat(target, availableQualities(videos))
	codecOrder := uniqueCodecs(preferredCodec, 12, 7, 13)
	for _, codec := range codecOrder {
		for _, item := range videos {
			if int(item.ID) == quality && item.Codecid == codec {
				return mediaToSelected(item, quality), true
			}
		}
	}
	for _, item := range videos {
		if int(item.ID) == quality {
			return mediaToSelected(item, quality), true
		}
	}
	return SelectedVideo{}, false
}

func mediaToSelected(item bilibili.Media, quality int) SelectedVideo {
	return SelectedVideo{
		URL:     item.BaseURL,
		Quality: quality,
		Width:   item.Width,
		Height:  item.Height,
		Codecid: item.Codecid,
		Codecs:  item.Codecs,
	}
}

func uniqueCodecs(values ...int) []int {
	seen := map[int]struct{}{}
	var list []int
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		list = append(list, v)
	}
	return list
}

func SelectAudio(dash *bilibili.Dash, preferHiRes bool) string {
	item := SelectAudioMedia(dash, preferHiRes)
	if item == nil {
		return ""
	}
	return item.BaseURL
}

func SelectAudioMedia(dash *bilibili.Dash, preferHiRes bool) *bilibili.Media {
	if dash == nil {
		return nil
	}
	if preferHiRes && dash.Flac != nil {
		audio := dash.Flac.Audio
		return &audio
	}
	maxID := -1
	var best *bilibili.Media
	for i := range dash.Audio {
		item := &dash.Audio[i]
		if int(item.ID) > maxID {
			maxID = int(item.ID)
			best = item
		}
	}
	return best
}

func VideoTag(quality, width, height, codecid int, codecs string) string {
	if tag := resolutionTag(width, height); tag != "" {
		return tag
	}
	return qualityTag(quality)
}

func AudioTag(preferHiRes bool, codecs string, flac bool) string {
	c := strings.ToLower(codecs)
	if flac || preferHiRes && strings.Contains(c, "flac") {
		return "FLAC"
	}
	if strings.Contains(c, "ec-3") || strings.Contains(c, "eac3") || strings.Contains(c, "e-ac-3") {
		return "EAC3"
	}
	return "AAC"
}

func CodecTag(codecid int, codecs string) string {
	c := strings.ToLower(codecs)
	switch codecid {
	case 12:
		return "H265"
	case 13:
		return "AV1"
	case 7:
		return "H264"
	}
	if strings.Contains(c, "hev") || strings.Contains(c, "hvc") || strings.Contains(c, "hevc") {
		return "H265"
	}
	if strings.Contains(c, "av01") || strings.Contains(c, "av1") {
		return "AV1"
	}
	return "H264"
}

func resolutionTag(width, height int) string {
	h := height
	if width > 0 && width < height {
		h = width
	}
	switch {
	case h >= 2160:
		return "2160p"
	case h >= 1440:
		return "1440p"
	case h >= 1080:
		return "1080p"
	case h >= 720:
		return "720p"
	case h >= 480:
		return "480p"
	case h >= 360:
		return "360p"
	case h > 0:
		return "240p"
	default:
		return ""
	}
}

func qualityTag(quality int) string {
	switch quality {
	case 127:
		return "4320p"
	case 120, 125, 126:
		return "2160p"
	case 112, 116, 80:
		return "1080p"
	case 74, 64:
		return "720p"
	case 32:
		return "480p"
	case 16:
		return "360p"
	case 6:
		return "240p"
	default:
		return "1080p"
	}
}
