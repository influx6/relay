package relay

import (
	"fmt"
	"testing"

	"github.com/influx6/flux"
)

func TestCollector(t *testing.T) {
	c := NewCollector()

	c.Set("name", "bond")

	if !c.HasMatch("name", "bond") {
		flux.FatalFailed(t, "unable to match added key %s and value %s", "name", "bond")
	}

	flux.LogPassed(t, "correctly matched added key %s and value %s", "name", "bond")
}

func TestSyncCollector(t *testing.T) {
	c := NewSyncCollector()

	c.Set("name", "bond")

	if !c.HasMatch("name", "bond") {
		flux.FatalFailed(t, "unable to match added key %s and value %s", "name", "bond")
	}

	flux.LogPassed(t, "correctly matched added key %s and value %s", "name", "bond")
}

func BenchmarkSyncCollector(t *testing.B) {
	for i := 0; i < t.N; i++ {
		c := NewSyncCollector()

		//add absurd amount of digits (3000) into the buffer
		for i := 0; i <= 7000; i++ {
			c.Set(fmt.Sprintf("%d", i), 1+2)
		}

		//remove 1000 total digits from map
		for i := 0; i <= 7000; i++ {
			c.Remove(fmt.Sprintf("%d", i))
		}

		//clear out map and ensure its empty
		c.Clear()
	}
}
