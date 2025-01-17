package conntrack

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/sirupsen/logrus"
)

type Tracker struct {
	*sync.Map
	ShuttingDown  atomic.Value
	Wg            *sync.WaitGroup
	IdleThreshold time.Duration // A connection is idle if it has been inactive (no bytes in/out) for this many seconds.
	Log           *logrus.Logger
	statsc        *statsd.Client
}

func NewTracker(idle time.Duration, statsc *statsd.Client, logger *logrus.Logger, sd atomic.Value) *Tracker {
	return &Tracker{
		Map:           &sync.Map{},
		ShuttingDown:  sd,
		Wg:            &sync.WaitGroup{},
		IdleThreshold: idle,
		Log:           logger,
		statsc:        statsc,
	}
}

// MaybeIdleIn returns the longest amount of time it will take for all tracked
// connections to become idle based on the configured IdleThreshold.
//
// A duration of 0 indicates all connections are idle.
func (tr *Tracker) MaybeIdleIn() time.Duration {
	longest := 0 * time.Nanosecond
	tr.Range(func(k, v interface{}) bool {
		c := k.(*InstrumentedConn)

		lastActivity := time.Unix(0, atomic.LoadInt64(c.LastActivity))
		idleAt := lastActivity.Add(tr.IdleThreshold)
		idleIn := idleAt.Sub(time.Now())

		if idleIn > longest {
			longest = idleIn
		}
		return true
	})
	return longest
}
