package subscription

import (
	"log"
	"sync"

	"bilidown/util"

	"github.com/robfig/cron/v3"
)

var (
	schedulerMu sync.Mutex
	cronRunner  *cron.Cron
	cronEntries = map[int64]cron.EntryID{}
	checkMu     sync.Mutex
)

func cronParser() cron.Parser {
	return cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
}

func StartScheduler() {
	ReloadScheduler()
}

func ReloadScheduler() {
	db := util.MustGetDB()
	defer db.Close()
	sources, err := ListSources(db)
	if err != nil {
		log.Printf("subscription scheduler: %v", err)
		return
	}
	schedulerMu.Lock()
	defer schedulerMu.Unlock()
	ensureRunnerLocked()
	for _, entryID := range cronEntries {
		cronRunner.Remove(entryID)
	}
	cronEntries = map[int64]cron.EntryID{}
	for _, src := range sources {
		if !src.Enabled {
			continue
		}
		if err := addSourceLocked(src); err != nil {
			log.Printf("subscription scheduler id=%d: %v", src.ID, err)
		}
	}
}

func SyncSource(src Source) {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()
	ensureRunnerLocked()
	removeSourceLocked(src.ID)
	if !src.Enabled {
		return
	}
	if err := addSourceLocked(src); err != nil {
		log.Printf("subscription scheduler id=%d: %v", src.ID, err)
	}
}

func RemoveScheduledSource(id int64) {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()
	removeSourceLocked(id)
}

func ensureRunnerLocked() {
	if cronRunner != nil {
		return
	}
	cronRunner = cron.New(cron.WithParser(cronParser()), cron.WithLocation(shanghaiLocation()))
	cronRunner.Start()
}

func addSourceLocked(src Source) error {
	expr := SourceCron(src.Cron)
	id := src.ID
	entryID, err := cronRunner.AddFunc(expr, func() {
		RunCheck(id)
	})
	if err != nil {
		return err
	}
	cronEntries[id] = entryID
	return nil
}

func removeSourceLocked(id int64) {
	if cronRunner == nil {
		return
	}
	if entryID, ok := cronEntries[id]; ok {
		cronRunner.Remove(entryID)
		delete(cronEntries, id)
	}
}

func ValidateCron(expr string) error {
	_, err := cronParser().Parse(expr)
	return err
}

func RunCheck(onlyID int64) {
	checkMu.Lock()
	defer checkMu.Unlock()
	db := util.MustGetDB()
	defer db.Close()
	if err := CheckAll(db, onlyID); err != nil {
		log.Printf("subscription check: %v", err)
	}
}
