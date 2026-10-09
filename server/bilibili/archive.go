package bilibili

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type ArchiveItem struct {
	Bvid     string
	Cid      int
	Title    string
	Cover    string
	Owner    string
	PubTime  int64
	Duration int
	ItemKey  string
}

func (client *BiliClient) ListSeasonArchives(mid int, seasonId int) ([]ArchiveItem, error) {
	if client.SESSDATA == "" {
		return nil, errors.New("SESSDATA 不能为空")
	}
	var all []ArchiveItem
	page := 1
	for {
		params := map[string]string{
			"mid":          strconv.Itoa(mid),
			"season_id":    strconv.Itoa(seasonId),
			"page_num":     strconv.Itoa(page),
			"page_size":    "30",
			"sort_reverse": "false",
		}
		response, err := client.SignedGET("https://api.bilibili.com/x/polymer/web-space/seasons_archives_list", params)
		if err != nil {
			return nil, err
		}
		body := BaseResV2{}
		if err = json.NewDecoder(response.Body).Decode(&body); err != nil {
			response.Body.Close()
			return nil, err
		}
		response.Body.Close()
		if body.Code != 0 {
			return nil, errors.New(body.Message)
		}
		var data struct {
			Page struct {
				PageNum  int `json:"page_num"`
				PageSize int `json:"page_size"`
				Total    int `json:"total"`
			} `json:"page"`
			Archives []struct {
				Bvid     string `json:"bvid"`
				Title    string `json:"title"`
				Pic      string `json:"pic"`
				Duration int    `json:"duration"`
				Pubtime  int64  `json:"pubtime"`
			} `json:"archives"`
		}
		if err = json.Unmarshal(body.Data, &data); err != nil {
			return nil, err
		}
		for _, item := range data.Archives {
			all = append(all, ArchiveItem{
				Bvid:     item.Bvid,
				Title:    item.Title,
				Cover:    item.Pic,
				PubTime:  item.Pubtime,
				Duration: item.Duration,
				ItemKey:  item.Bvid,
			})
		}
		if len(data.Archives) == 0 || len(all) >= data.Page.Total {
			break
		}
		page++
		if page > 100 {
			break
		}
	}
	return all, nil
}

func (client *BiliClient) ListSeriesArchives(mid int, seriesId int) ([]ArchiveItem, error) {
	if client.SESSDATA == "" {
		return nil, errors.New("SESSDATA 不能为空")
	}
	var all []ArchiveItem
	page := 1
	for {
		params := map[string]string{
			"mid":       strconv.Itoa(mid),
			"series_id": strconv.Itoa(seriesId),
			"pn":        strconv.Itoa(page),
			"ps":        "30",
		}
		response, err := client.SignedGET("https://api.bilibili.com/x/series/archives", params)
		if err != nil {
			return nil, err
		}
		body := BaseResV2{}
		if err = json.NewDecoder(response.Body).Decode(&body); err != nil {
			response.Body.Close()
			return nil, err
		}
		response.Body.Close()
		if body.Code != 0 {
			return nil, errors.New(body.Message)
		}
		var data struct {
			Page struct {
				Num   int `json:"num"`
				Size  int `json:"size"`
				Total int `json:"total"`
			} `json:"page"`
			Archives []struct {
				Bvid     string `json:"bvid"`
				Title    string `json:"title"`
				Pic      string `json:"pic"`
				Duration int    `json:"duration"`
				Pubts    int64  `json:"pubts"`
			} `json:"archives"`
		}
		if err = json.Unmarshal(body.Data, &data); err != nil {
			return nil, err
		}
		for _, item := range data.Archives {
			all = append(all, ArchiveItem{
				Bvid:     item.Bvid,
				Title:    item.Title,
				Cover:    item.Pic,
				PubTime:  item.Pubts,
				Duration: item.Duration,
				ItemKey:  item.Bvid,
			})
		}
		if len(data.Archives) == 0 || len(all) >= data.Page.Total {
			break
		}
		page++
		if page > 100 {
			break
		}
	}
	return all, nil
}

func (client *BiliClient) ListSpaceArchives(mid int) ([]ArchiveItem, error) {
	if client.SESSDATA == "" {
		return nil, errors.New("SESSDATA 不能为空")
	}
	var all []ArchiveItem
	page := 1
	for {
		params := map[string]string{
			"mid":   strconv.Itoa(mid),
			"pn":    strconv.Itoa(page),
			"ps":    "30",
			"order": "pubdate",
		}
		response, err := client.SignedGET("https://api.bilibili.com/x/space/wbi/arc/search", params)
		if err != nil {
			return nil, err
		}
		body := BaseResV2{}
		if err = json.NewDecoder(response.Body).Decode(&body); err != nil {
			response.Body.Close()
			return nil, err
		}
		response.Body.Close()
		if body.Code != 0 {
			return nil, errors.New(body.Message)
		}
		var data struct {
			List struct {
				Vlist []struct {
					Bvid        string `json:"bvid"`
					Title       string `json:"title"`
					Pic         string `json:"pic"`
					Length      string `json:"length"`
					Created     int64  `json:"created"`
					Author      string `json:"author"`
					Description string `json:"description"`
				} `json:"vlist"`
			} `json:"list"`
			Page struct {
				Pn    int `json:"pn"`
				Ps    int `json:"ps"`
				Count int `json:"count"`
			} `json:"page"`
		}
		if err = json.Unmarshal(body.Data, &data); err != nil {
			return nil, err
		}
		for _, item := range data.List.Vlist {
			all = append(all, ArchiveItem{
				Bvid:    item.Bvid,
				Title:   item.Title,
				Cover:   item.Pic,
				Owner:   item.Author,
				PubTime: item.Created,
				ItemKey: item.Bvid,
			})
		}
		if len(data.List.Vlist) == 0 || len(all) >= data.Page.Count {
			break
		}
		page++
		if page > 100 {
			break
		}
	}
	return all, nil
}

func (client *BiliClient) ListFavArchives(mediaId int) ([]ArchiveItem, error) {
	favList, err := client.GetFavlist(mediaId)
	if err != nil {
		return nil, err
	}
	if favList == nil {
		return nil, nil
	}
	var all []ArchiveItem
	for _, item := range *favList {
		all = append(all, ArchiveItem{
			Bvid:     item.Bvid,
			Cid:      item.Ugc.FirstCid,
			Title:    item.Title,
			Cover:    item.Cover,
			Owner:    item.Upper.Name,
			PubTime:  int64(item.PubTime),
			Duration: item.Duration,
			ItemKey:  item.Bvid,
		})
	}
	return all, nil
}

type BangumiSection struct {
	ID    string        `json:"id"`
	Title string        `json:"title"`
	Items []ArchiveItem `json:"-"`
	Count int           `json:"count"`
}

func episodeToArchive(info *SeasonInfo, ep Episode) ArchiveItem {
	title := ep.LongTitle
	if title == "" {
		title = ep.Title
	}
	return ArchiveItem{
		Bvid:     ep.Bvid,
		Cid:      ep.Cid,
		Title:    title,
		Cover:    ep.Cover,
		Owner:    info.Actors,
		PubTime:  int64(ep.PubTime),
		Duration: ep.Duration,
		ItemKey:  fmt.Sprintf("ep_%d", ep.EPID),
	}
}

func (client *BiliClient) ListBangumiSections(epid int, ssid int) ([]BangumiSection, *SeasonInfo, error) {
	info, err := client.GetSeasonInfo(epid, ssid)
	if err != nil {
		return nil, nil, err
	}
	sections := []BangumiSection{}
	if len(info.Episodes) > 0 {
		items := make([]ArchiveItem, 0, len(info.Episodes))
		for _, ep := range info.Episodes {
			items = append(items, episodeToArchive(info, ep))
		}
		sections = append(sections, BangumiSection{ID: "main", Title: "正片", Items: items, Count: len(items)})
	}
	for i, section := range info.Section {
		title := section.Title
		if title == "" {
			title = "其他"
		}
		items := make([]ArchiveItem, 0, len(section.Episodes))
		for _, ep := range section.Episodes {
			items = append(items, episodeToArchive(info, ep))
		}
		sections = append(sections, BangumiSection{
			ID:    fmt.Sprintf("section_%d", i),
			Title: title,
			Items: items,
			Count: len(items),
		})
	}
	return sections, info, nil
}

func FilterBangumiSections(sections []BangumiSection, ids []string) []ArchiveItem {
	var all []ArchiveItem
	if len(ids) == 0 {
		for _, section := range sections {
			all = append(all, section.Items...)
		}
		return all
	}
	want := map[string]struct{}{}
	for _, id := range ids {
		want[id] = struct{}{}
	}
	for _, section := range sections {
		if _, ok := want[section.ID]; ok {
			all = append(all, section.Items...)
		}
	}
	return all
}

func (client *BiliClient) ListBangumiEpisodes(epid int, ssid int) ([]ArchiveItem, *SeasonInfo, error) {
	sections, info, err := client.ListBangumiSections(epid, ssid)
	if err != nil {
		return nil, nil, err
	}
	return FilterBangumiSections(sections, nil), info, nil
}

func (client *BiliClient) FillArchiveDetail(item *ArchiveItem) error {
	if item.Bvid == "" {
		return errors.New("bvid 为空")
	}
	if item.Cid > 0 && item.Duration > 0 && item.Cover != "" && item.Owner != "" {
		return nil
	}
	info, err := client.GetVideoInfo(item.Bvid)
	if err != nil {
		return err
	}
	if item.Cid == 0 && len(info.Pages) > 0 {
		item.Cid = info.Pages[0].Cid
	}
	if item.Title == "" {
		item.Title = info.Title
	}
	if item.Cover == "" {
		item.Cover = info.Pic
	}
	if item.Owner == "" {
		item.Owner = info.Owner.Name
	}
	if item.Duration == 0 {
		item.Duration = info.Duration
	}
	if item.PubTime == 0 {
		item.PubTime = int64(info.Pubdate)
	}
	return nil
}
