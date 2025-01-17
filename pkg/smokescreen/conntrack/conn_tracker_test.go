package conntrack

import (
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestConnTrackerDelete is a sanity check to ensure we aren't leaking
// connection references in the tracker's sync.Map
func TestConnTrackerDelete(t *testing.T) {
	tr := NewTestTracker(time.Second * 1)

	ic := tr.NewInstrumentedConn(&net.UnixConn{}, "testDeleteConn", "localhost")
	ic.Close()

	tr.Range(func(k, v interface{}) bool {
		t.Fatal("conn map should be empty")
		return false
	})
}

// TestConnTrackerMaybeIdleIn tests that our `IdleIn` calculations are correct
func TestConnTrackerMaybeIdleIn(t *testing.T) {
	assert := assert.New(t)

	tr := NewTestTracker(1 * time.Nanosecond)
	ic := tr.NewInstrumentedConn(&net.UnixConn{}, "testMaybeIdle", "localhost")

	time.Sleep(time.Millisecond)

	// All connections should be idle
	assert.Zero(tr.MaybeIdleIn())

	tr.IdleThreshold = time.Second
	ic.Write([]byte("egress"))

	idleIn := tr.MaybeIdleIn().Round(time.Second)
	assert.Equal(time.Second, idleIn)
}

func NewTestTracker(idle time.Duration) *Tracker {
	sd := atomic.Value{}
	sd.Store(false)

	return NewTracker(idle, nil, logrus.New(), sd)
}
