package xwindowmgr

import (
	"context"
	"time"
)

type EventSource interface {
	Pending() int
	NextEvent() WMEvent
}

const eventPollInterval = 15 * time.Millisecond

func (xm *XWindowManager) EventLoop(ctx context.Context, eventChan chan<- WMEvent) {

	for {
		if ctx.Err() != nil {
			return
		}
		if xm.backend.Pending() == 0 {
			timer := time.NewTimer(eventPollInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			continue
		}

		event := xm.backend.NextEvent()
		if !sendWMEvent(ctx, eventChan, event) {
			return
		}
	}
}

func sendWMEvent(ctx context.Context, events chan<- WMEvent, event WMEvent) bool {
	select {
	case <-ctx.Done():
		return false
	case events <- event:
		return true
	}
}
