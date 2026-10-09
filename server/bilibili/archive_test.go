package bilibili

import "testing"

func TestFilterBangumiSectionsSelected(t *testing.T) {
	sections := []BangumiSection{
		{ID: "main", Title: "正片", Items: []ArchiveItem{{ItemKey: "ep_1"}, {ItemKey: "ep_2"}}},
		{ID: "section_0", Title: "精彩看点", Items: []ArchiveItem{{ItemKey: "ep_3"}}},
	}
	got := FilterBangumiSections(sections, []string{"main"})
	if len(got) != 2 || got[0].ItemKey != "ep_1" {
		t.Fatalf("got %#v", got)
	}
}

func TestFilterBangumiSectionsAllWhenEmpty(t *testing.T) {
	sections := []BangumiSection{
		{ID: "main", Items: []ArchiveItem{{ItemKey: "ep_1"}}},
		{ID: "section_0", Items: []ArchiveItem{{ItemKey: "ep_2"}}},
	}
	got := FilterBangumiSections(sections, nil)
	if len(got) != 2 {
		t.Fatalf("got %d", len(got))
	}
}
