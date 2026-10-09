package subscription

const FirstRunWindowSeconds int64 = 24 * 60 * 60

func ShouldCreateTask(firstRun bool, publishedAt, now int64) bool {
	if publishedAt <= 0 {
		return !firstRun
	}
	if publishedAt > now {
		return false
	}
	if !firstRun {
		return true
	}
	return now-publishedAt <= FirstRunWindowSeconds
}
