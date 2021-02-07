package requestmonitor

import (
	"testing"
	"time"
)

func Test_RequestCounterWorks(t *testing.T) {
	key := "foo"
	monitor := NewRequestMonitor()
	check := monitor.CheckRequest(key)

	if check != 0 {
		t.Errorf("expected check to be 0 but was %v", check)
	}

	monitor.NewRequest(key)

	check = monitor.CheckRequest(key)

	if check != 1 {
		t.Errorf("expected check to be 1 but was %v", check)
	}

	monitor.NewRequest(key)

	check = monitor.CheckRequest(key)

	if check != 2 {
		t.Errorf("expected check to be 2 but was %v", check)
	}
}

func Test_RequestCounterJanitorWorks(t *testing.T) {
	key := "foo"
	monitor := NewRequestMonitor()
	check := monitor.CheckRequest(key)
	if check != 0 {
		t.Errorf("expected check to be 0 but was %v", check)
	}

	monitor.NewRequest(key)
	monitor.NewRequest(key)

	check = monitor.CheckRequest(key)
	if check != 2 {
		t.Errorf("expected check to be 2 but was %v", check)
	}

	stop := make(chan struct{})
	go monitor.StartJanitor(20*time.Millisecond, 50*time.Millisecond, stop)

	// wait for the first janitor cycle
	time.Sleep(30 * time.Millisecond)
	check = monitor.CheckRequest(key)
	if check != 2 {
		t.Errorf("expected check to be 2 but was %v", check)
	}

	// wait for the janitor to have cleaned up the state resetting the count to 0
	time.Sleep(500 * time.Millisecond)
	check = monitor.CheckRequest(key)
	if check != 0 {
		t.Errorf("expected check to be 0 but was %v", check)
	}

	stop <- struct{}{}
}
