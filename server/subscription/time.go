package subscription

import (
	"time"
	_ "time/tzdata"
)

func shanghaiLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}

func FormatNow() string {
	return time.Now().In(shanghaiLocation()).Format("2006-01-02 15:04:05")
}
