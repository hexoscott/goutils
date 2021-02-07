package requestmonitor

import (
	"sync"
	"time"
)

/*
RequestMonitor holds details about inbound requests keyed on an arbitrary string
to use it create one using NewRequestMonitor which will initialise the mutex and map.

Once created usage would typically be along the lines of starting the janitor routine,
then when a new request comes in e.g. a web request check if the monitor has anything for
the request IP for instance using CheckRequest, if above a threshold bin the request off,
if not increment the count using NewRequest
*/
type RequestMonitor struct {
	mtx      sync.Mutex
	counters map[string]*Counter
	ticker   <-chan time.Time
}

// NewRequestMonitor returns a pointer to a new instance of RequestMonitor
func NewRequestMonitor() *RequestMonitor {
	return &RequestMonitor{
		mtx:      sync.Mutex{},
		counters: map[string]*Counter{},
	}
}

// Counter holds the last attempt time for a request and the count, this will be held against
// a key in the RequestMonitor and used by the janitor routine when cleaning up
type Counter struct {
	lastAttempt time.Time
	count       int
}

// CheckRequest takes a key and looks to the map to determine how many requests this key has had
func (r *RequestMonitor) CheckRequest(key string) int {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	c, ok := r.counters[key]

	if !ok {
		return 0
	}

	return c.count
}

// NewRequest creates a key in the RequestMonitors map if it doesn't already exist and increments the
// count
func (r *RequestMonitor) NewRequest(key string) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	c, ok := r.counters[key]

	if !ok {
		c = &Counter{
			lastAttempt: time.Now(),
			count:       1,
		}
		r.counters[key] = c
	} else {
		c.lastAttempt = time.Now()
		c.count++
	}
}

// StartJanitor must be called with 'go' as it will spin forever otherwise
// the janitor process will wake at the defined interval and clear out records that older
// than the requested timeframe
func (r *RequestMonitor) StartJanitor(when, remove time.Duration, stop chan struct{}) {
	r.ticker = time.Tick(when)

LOOP:
	for {
		select {
		case <-r.ticker:
			func() {
				r.mtx.Lock()
				defer r.mtx.Unlock()
				now := time.Now()

				for key, val := range r.counters {
					if now.Sub(val.lastAttempt) > remove {
						delete(r.counters, key)
					}
				}
			}()
		case <-stop:
			break LOOP
		}
	}
}
