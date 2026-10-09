package subscription

import "testing"

func TestParseSubscriptionURLSeason(t *testing.T) {
	src, err := ParseSubscriptionURL("https://space.bilibili.com/282565107/lists/1427135?type=season")
	if err != nil {
		t.Fatal(err)
	}
	if src.Type != TypeSeason || src.Mid != "282565107" || src.ResourceID != "1427135" {
		t.Fatalf("unexpected source: %+v", src)
	}
}

func TestParseSubscriptionURLSeries(t *testing.T) {
	src, err := ParseSubscriptionURL("https://space.bilibili.com/282565107/channel/seriesdetail?sid=998877")
	if err != nil {
		t.Fatal(err)
	}
	if src.Type != TypeSeries || src.ResourceID != "998877" {
		t.Fatalf("unexpected source: %+v", src)
	}
}

func TestParseSubscriptionURLFavAndSpace(t *testing.T) {
	fav, err := ParseSubscriptionURL("https://www.bilibili.com/list/ml123456")
	if err != nil {
		t.Fatal(err)
	}
	if fav.Type != TypeFav || fav.ResourceID != "123456" {
		t.Fatalf("unexpected fav: %+v", fav)
	}
	space, err := ParseSubscriptionURL("https://space.bilibili.com/1176277996")
	if err != nil {
		t.Fatal(err)
	}
	if space.Type != TypeSpace || space.ResourceID != "1176277996" {
		t.Fatalf("unexpected space: %+v", space)
	}
}

func TestParseSubscriptionURLBangumi(t *testing.T) {
	ss, err := ParseSubscriptionURL("https://www.bilibili.com/bangumi/play/ss48831")
	if err != nil {
		t.Fatal(err)
	}
	if ss.Type != TypeBangumi || ss.ResourceID != "48831" || ss.FromEP {
		t.Fatalf("unexpected ss: %+v", ss)
	}
	ep, err := ParseSubscriptionURL("ep12345")
	if err != nil {
		t.Fatal(err)
	}
	if !ep.FromEP || ep.EPID != 12345 {
		t.Fatalf("unexpected ep: %+v", ep)
	}
}

func TestParseSubscriptionURLVideo(t *testing.T) {
	src, err := ParseSubscriptionURL("https://www.bilibili.com/video/BV1LLDCYJEU3/")
	if err != nil {
		t.Fatal(err)
	}
	if src.Type != TypeVideo || src.ResourceID != "BV1LLDCYJEU3" {
		t.Fatalf("unexpected video: %+v", src)
	}
}

func TestParseSubscriptionURLB23(t *testing.T) {
	src, err := ParseSubscriptionURL("https://b23.tv/lzMEJVu")
	if err != nil {
		t.Skip(err.Error())
	}
	if src.Type != TypeVideo || src.ResourceID != "BV1XSpp62Emi" {
		t.Fatalf("unexpected b23 source: %+v", src)
	}
}
