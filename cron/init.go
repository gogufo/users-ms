package cron

import (
	"sync"
	"time"

	. "github.com/gogufo/gufo-api-gateway/gufodao"
)

var (
	// Local flag (admin/config level)
	localEnabled bool

	// Heartbeat permission flag (cluster level)
	heartbeatAllowed bool

	// Indicates whether cron is currently running
	cronEnabled bool

	stopChan chan struct{}
	doneChan chan struct{}
	mu       sync.Mutex
)

// ApplyCronState is called from heartbeat (cluster permission)
func ApplyCronState(flag bool) {
	mu.Lock()
	heartbeatAllowed = flag
	mu.Unlock()

	reconcile()
}

// SetLocalCronState is called from admin/config
func SetLocalCronState(flag bool) {
	mu.Lock()
	localEnabled = flag
	mu.Unlock()

	reconcile()
}

// reconcile starts or stops cron based on both flags
func reconcile() {
	mu.Lock()
	defer mu.Unlock()

	shouldRun := localEnabled && heartbeatAllowed

	// Start cron
	if shouldRun && !cronEnabled {
		stopChan = make(chan struct{})
		doneChan = make(chan struct{})
		go Init()
		cronEnabled = true
		SetLog("[CRON] started (local + heartbeat)")
		return
	}

	// Stop cron
	if !shouldRun && cronEnabled {
		close(stopChan)
		<-doneChan
		cronEnabled = false
		SetLog("[CRON] stopped (local or heartbeat disabled)")
		return
	}
}

func Init() {
	defer close(doneChan)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			CronJob()

		case <-stopChan:
			SetLog("[CRON] stop signal received")
			return
		}
	}
}

func CronJob() {
	// Put your cron job codes here
}
