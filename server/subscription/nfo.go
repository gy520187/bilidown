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
	Owner       string
	BVID        string
	Cover       string
	PublishedAt int64
}

type nfoMovie struct {
	XMLName   xml.Name `xml:"movie"`
	Title     string   `xml:"title"`
	Studio    string   `xml:"studio"`
	UniqueID  nfoID    `xml:"uniqueid"`
	Thumb     string   `xml:"thumb"`
	Premiered string   `xml:"premiered"`
}

type nfoID struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

func NFOPath(mediaPath string) string {
	ext := filepath.Ext(mediaPath)
	return strings.TrimSuffix(mediaPath, ext) + ".nfo"
}

func RenderNFO(data NFOData) ([]byte, error) {
	premiered := ""
	if data.PublishedAt > 0 {
		premiered = time.Unix(data.PublishedAt, 0).UTC().Format("2006-01-02")
	}
	movie := nfoMovie{
		Title:  data.Title,
		Studio: data.Owner,
		UniqueID: nfoID{
			Type:  "bilibili",
			Value: data.BVID,
		},
		Thumb:     data.Cover,
		Premiered: premiered,
	}
	body, err := xml.MarshalIndent(movie, "", "  ")
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
