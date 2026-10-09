package subscription

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"bilidown/util"
)

const (
	TypeSeason  = "season"
	TypeSeries  = "series"
	TypeFav     = "fav"
	TypeSpace   = "space"
	TypeBangumi = "bangumi"
	TypeVideo   = "video"
)

type ParsedSource struct {
	Type       string
	Mid        string
	ResourceID string
	FromEP     bool
	EPID       int
}

var (
	reBVID        = regexp.MustCompile(`(?i)(?:https?://(?:www\.)?bilibili\.com/video/)?(BV1[a-zA-Z0-9]+)`)
	reBangumi     = regexp.MustCompile(`(?i)(?:https?://(?:www\.)?bilibili\.com/bangumi/play/)?(ep|ss)(\d+)`)
	reFavML       = regexp.MustCompile(`(?i)https?://(?:www\.)?bilibili\.com/medialist/detail/ml(\d+)`)
	reFavShort    = regexp.MustCompile(`(?i)https?://(?:www\.)?bilibili\.com/list/ml(\d+)`)
	reLists       = regexp.MustCompile(`^/(\d+)/lists/(\d+)`)
	reCollection  = regexp.MustCompile(`^/(\d+)/channel/collectiondetail$`)
	reSeries      = regexp.MustCompile(`^/(\d+)/channel/seriesdetail$`)
	reFavlistPath = regexp.MustCompile(`^/(\d+)/favlist$`)
	reSpacePath   = regexp.MustCompile(`^/(\d+)/?$`)
	reSpaceVideo  = regexp.MustCompile(`^/(\d+)/video/?$`)
	reSpaceUpload = regexp.MustCompile(`^/(\d+)/upload/video/?$`)
)

func ParseSubscriptionURL(raw string) (ParsedSource, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ParsedSource{}, errors.New("链接为空")
	}

	if src, ok := parseBangumi(raw); ok {
		return src, nil
	}
	expanded, err := expandShortURL(raw)
	if err != nil {
		return ParsedSource{}, err
	}
	if expanded != raw {
		if src, ok := parseBangumi(expanded); ok {
			return src, nil
		}
		raw = expanded
	}
	if src, ok := parseHTTPSource(raw); ok {
		return src, nil
	}
	if m := reBVID.FindStringSubmatch(raw); m != nil {
		return ParsedSource{Type: TypeVideo, ResourceID: m[1]}, nil
	}
	return ParsedSource{}, errors.New("无法识别的订阅链接")
}

func expandShortURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		if !strings.Contains(raw, "://") {
			parsed, err = url.Parse("https://" + raw)
			if err != nil || parsed.Host == "" {
				return raw, nil
			}
		} else {
			return raw, nil
		}
	}
	if strings.ToLower(parsed.Hostname()) != "b23.tv" {
		return raw, nil
	}
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	location, err := util.GetRedirectedLocation(parsed.String())
	if err != nil {
		return "", fmt.Errorf("无法解析短链接")
	}
	return location, nil
}

func parseBangumi(raw string) (ParsedSource, bool) {
	match := reBangumi.FindStringSubmatch(raw)
	if match == nil {
		return ParsedSource{}, false
	}
	kind := strings.ToLower(match[1])
	id := match[2]
	if kind == "ss" {
		return ParsedSource{Type: TypeBangumi, ResourceID: id}, true
	}
	epid, _ := strconv.Atoi(id)
	return ParsedSource{Type: TypeBangumi, FromEP: true, EPID: epid}, true
}

func parseHTTPSource(raw string) (ParsedSource, bool) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		if !strings.Contains(raw, "://") {
			parsed, err = url.Parse("https://" + raw)
			if err != nil || parsed.Host == "" {
				return ParsedSource{}, false
			}
		} else {
			return ParsedSource{}, false
		}
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "www.bilibili.com" || host == "bilibili.com" {
		if m := reFavML.FindStringSubmatch(raw); m != nil {
			return ParsedSource{Type: TypeFav, ResourceID: m[1]}, true
		}
		if m := reFavShort.FindStringSubmatch(raw); m != nil {
			return ParsedSource{Type: TypeFav, ResourceID: m[1]}, true
		}
		return ParsedSource{}, false
	}
	if host != "space.bilibili.com" {
		return ParsedSource{}, false
	}

	path := parsed.Path
	if m := reLists.FindStringSubmatch(path); m != nil {
		kind := parsed.Query().Get("type")
		if kind == "series" {
			return ParsedSource{Type: TypeSeries, Mid: m[1], ResourceID: m[2]}, true
		}
		return ParsedSource{Type: TypeSeason, Mid: m[1], ResourceID: m[2]}, true
	}
	if m := reCollection.FindStringSubmatch(path); m != nil {
		sid := parsed.Query().Get("sid")
		if sid == "" {
			return ParsedSource{}, false
		}
		return ParsedSource{Type: TypeSeason, Mid: m[1], ResourceID: sid}, true
	}
	if m := reSeries.FindStringSubmatch(path); m != nil {
		sid := parsed.Query().Get("sid")
		if sid == "" {
			return ParsedSource{}, false
		}
		return ParsedSource{Type: TypeSeries, Mid: m[1], ResourceID: sid}, true
	}
	if m := reFavlistPath.FindStringSubmatch(path); m != nil {
		fid := parsed.Query().Get("fid")
		if fid == "" {
			return ParsedSource{}, false
		}
		return ParsedSource{Type: TypeFav, Mid: m[1], ResourceID: fid}, true
	}
	if m := reSpacePath.FindStringSubmatch(path); m != nil {
		return ParsedSource{Type: TypeSpace, Mid: m[1], ResourceID: m[1]}, true
	}
	if m := reSpaceVideo.FindStringSubmatch(path); m != nil {
		return ParsedSource{Type: TypeSpace, Mid: m[1], ResourceID: m[1]}, true
	}
	if m := reSpaceUpload.FindStringSubmatch(path); m != nil {
		return ParsedSource{Type: TypeSpace, Mid: m[1], ResourceID: m[1]}, true
	}
	return ParsedSource{}, false
}
