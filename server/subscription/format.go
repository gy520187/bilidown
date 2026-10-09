package subscription

import "bilidown/bilibili"

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

func SelectVideo(videos []bilibili.Media, target int, preferredCodec int) (url string, quality int, ok bool) {
	quality = SelectFormat(target, availableQualities(videos))
	codecOrder := uniqueCodecs(preferredCodec, 12, 7, 13)
	for _, codec := range codecOrder {
		for _, item := range videos {
			if int(item.ID) == quality && item.Codecid == codec {
				return item.BaseURL, quality, true
			}
		}
	}
	for _, item := range videos {
		if int(item.ID) == quality {
			return item.BaseURL, quality, true
		}
	}
	return "", 0, false
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
	if dash == nil {
		return ""
	}
	if preferHiRes && dash.Flac != nil {
		return dash.Flac.Audio.BaseURL
	}
	maxID := -1
	url := ""
	for _, item := range dash.Audio {
		if int(item.ID) > maxID {
			maxID = int(item.ID)
			url = item.BaseURL
		}
	}
	return url
}
