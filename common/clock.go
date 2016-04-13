package common

import (
	"sync"
	"time"
)

// Used where I want to make use of the functions of the time package,
// but also want to be able to mock it out during testing
// (e.g. to simulate timeouts)

type Clock interface {
	Now() time.Time
	Sleep(d time.Duration)
	After(d time.Duration) <-chan time.Time
}

type RealClock struct{}

func (*RealClock) Now() time.Time                         { return time.Now() }
func (*RealClock) Sleep(d time.Duration)                  { time.Sleep(d) }
func (*RealClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

type FakeClock struct {
	lock    sync.Mutex
	cond    *sync.Cond
	nowTime time.Time
}

func NewFakeClock(initialTime time.Time) *FakeClock {
	fc := new(FakeClock)
	fc.lock = sync.Mutex{}
	fc.cond = sync.NewCond(&fc.lock)
	fc.nowTime = initialTime
	return fc
}
func (fc *FakeClock) Now() time.Time {
	fc.lock.Lock()
	defer fc.lock.Unlock()
	return fc.nowTime
}
func (fc *FakeClock) Sleep(d time.Duration) {
	fc.lock.Lock()
	endTime := fc.nowTime.Add(d)
	for fc.nowTime.Before(endTime) {
		fc.cond.Wait()
	}
	fc.lock.Unlock()
}
func (fc *FakeClock) After(d time.Duration) <-chan time.Time {
	ret := make(chan time.Time)
	go func() {
		fc.lock.Lock()
		endTime := fc.nowTime.Add(d)
		for fc.nowTime.Before(endTime) {
			fc.cond.Wait()
		}
		fc.lock.Unlock()
		ret <- endTime
	}()
	return ret
}
func (fc *FakeClock) SetNowTime(nowTime time.Time) {
	fc.lock.Lock()
	defer fc.lock.Unlock()
	fc.nowTime = nowTime
	fc.cond.Broadcast()
}
func (fc *FakeClock) AdvanceNowTime(length time.Duration) {
	fc.lock.Lock()
	defer fc.lock.Unlock()
	fc.nowTime = fc.nowTime.Add(length)
	fc.cond.Broadcast()
}
