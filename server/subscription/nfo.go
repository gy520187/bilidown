package subscription

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type NFOData struct {
	Title       string
	ShowTitle   string
	Owner       string
	BVID        string
	Cover       string
	PublishedAt int64
	Season      int
	Episode     int
}

type nfoMovie struct {
	XMLName   xml.Name `xml:"movie"`
	Title     string   `xml:"title"`
	Studio    string   `xml:"studio"`
	UniqueID  nfoID    `xml:"uniqueid"`
	Thumb     string   `xml:"thumb"`
	Premiered string   `xml:"premiered"`
}

type nfoEpisode struct {
	XMLName   xml.Name `xml:"episodedetails"`
	Title     string   `xml:"title"`
	ShowTitle string   `xml:"showtitle"`
	Season    int      `xml:"season"`
	Episode   int      `xml:"episode"`
	Studio    string   `xml:"studio"`
	UniqueID  nfoID    `xml:"uniqueid"`
	Thumb     string   `xml:"thumb"`
	Aired     string   `xml:"aired"`
	Premiered string   `xml:"premiered"`
}

type nfoTVShow struct {
	XMLName  xml.Name `xml:"tvshow"`
	Title    string   `xml:"title"`
	Studio   string   `xml:"studio"`
	UniqueID nfoID    `xml:"uniqueid"`
	Thumb    string   `xml:"thumb"`
}

type nfoID struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

func NFOPath(mediaPath string) string {
	ext := filepath.Ext(mediaPath)
	return strings.TrimSuffix(mediaPath, ext) + ".nfo"
}

func TVShowNFOPath(mediaPath string) string {
	dir := filepath.Dir(mediaPath)
	if strings.Contains(strings.ToLower(filepath.Base(dir)), "season") {
		dir = filepath.Dir(dir)
	}
	return filepath.Join(dir, "tvshow.nfo")
}

func premieredOf(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).UTC().Format("2006-01-02")
}

func RenderNFO(data NFOData) ([]byte, error) {
	movie := nfoMovie{
		Title:  data.Title,
		Studio: data.Owner,
		UniqueID: nfoID{
			Type:  "bilibili",
			Value: data.BVID,
		},
		Thumb:     data.Cover,
		Premiered: premieredOf(data.PublishedAt),
	}
	body, err := xml.MarshalIndent(movie, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), body...), nil
}

func RenderEpisodeNFO(data NFOData) ([]byte, error) {
	title := data.Title
	if title == "" {
		title = data.ShowTitle
	}
	ep := nfoEpisode{
		Title:     title,
		ShowTitle: data.ShowTitle,
		Season:    data.Season,
		Episode:   data.Episode,
		Studio:    data.Owner,
		UniqueID: nfoID{
			Type:  "bilibili",
			Value: data.BVID,
		},
		Thumb:     data.Cover,
		Aired:     premieredOf(data.PublishedAt),
		Premiered: premieredOf(data.PublishedAt),
	}
	body, err := xml.MarshalIndent(ep, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), body...), nil
}

func RenderTVShowNFO(data NFOData) ([]byte, error) {
	title := data.ShowTitle
	if title == "" {
		title = data.Title
	}
	show := nfoTVShow{
		Title:  title,
		Studio: data.Owner,
		UniqueID: nfoID{
			Type:  "bilibili",
			Value: data.BVID,
		},
		Thumb: data.Cover,
	}
	body, err := xml.MarshalIndent(show, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), body...), nil
}

func WriteNFO(mediaPath string, data NFOData) error {
	body, err := RenderNFO(data)
	if err != nil {
		return err
	}
	return os.WriteFile(NFOPath(mediaPath), body, 0644)
}

func WriteEpisodeNFO(mediaPath string, data NFOData) error {
	body, err := RenderEpisodeNFO(data)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(mediaPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(NFOPath(mediaPath), body, 0644)
}

func WriteTVShowNFO(path string, data NFOData) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	body, err := RenderTVShowNFO(data)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0644)
}
