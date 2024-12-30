package utils

import (
	"context"
	"sync"
	"time"

	"code.cd.local/sagittarius/sagittarius-golang/logger"

	"github.com/pkg/errors"
)

type transition struct {
	from    int
	to      int
	sec     int64
	handler func()
}

type TimerFSM struct {
	sync.Mutex
	running       bool
	state         int
	nextTimestamp int64
	stopCh        chan struct{}
	closeState    int
	transitions   []*transition
	wg            sync.WaitGroup
}

func NewTimerFSM() *TimerFSM {
	return &TimerFSM{
		running:     false,
		stopCh:      make(chan struct{}),
		closeState:  -1,
		transitions: make([]*transition, 0),
	}
}

func (fsm *TimerFSM) AddTransition(from int, to int, sec int64, h func()) error {
	for _, t := range fsm.transitions {
		if t.from == from {
			return errors.New("status clash")
		}
	}
	fsm.transitions = append(fsm.transitions, &transition{
		from:    from,
		to:      to,
		sec:     sec,
		handler: h,
	})
	return nil
}

func (fsm *TimerFSM) NextSec() int64 {
	return fsm.nextTimestamp / 1000
}

func (fsm *TimerFSM) State() int {
	return fsm.state
}

func (fsm *TimerFSM) Start(startState int, startSec int64) {
	fsm.Lock()
	defer fsm.Unlock()
	if fsm.running {
		return
	}
	fsm.running = true
	go fsm.run(startState, startSec)
}

func (fsm *TimerFSM) Stop() {
	fsm.stopCh <- struct{}{}
	fsm.wg.Wait()
}

func (fsm *TimerFSM) CloseWith(state int) {
	fsm.closeState = state
}

func (fsm *TimerFSM) findTransition(currentState int) *transition {
	for _, tr := range fsm.transitions {
		if tr.from == currentState {
			return tr
		}
	}
	return nil
}

func (fsm *TimerFSM) run(startState int, startSec int64) {
	fsm.state = startState

	fsm.nextTimestamp = time.Now().UnixMilli()/1000*1000 + 1 + startSec*1000
	cur := fsm.findTransition(fsm.state)
	for {
		var timer *time.Timer
		mi := time.Duration(fsm.nextTimestamp-time.Now().UnixMilli()) * time.Millisecond
		timer = time.NewTimer(mi)
		select {
		case <-fsm.stopCh:
			timer.Stop()
			return
		case <-timer.C:
			timer.Stop()
			next := fsm.findTransition(cur.to)
			fsm.state = cur.to
			if next == nil {
				break
			}
			cur = next
			fsm.nextTimestamp = time.Now().UnixMilli()/1000*1000 + 100 + cur.sec*1000
			if cur.handler != nil {
				fsm.wg.Add(1)
				go func(tr *transition) {
					defer fsm.wg.Done()

					tr.handler()
				}(cur)
			}
			if cur.from == fsm.closeState {
				fsm.wg.Wait()
				logger.Debug(context.TODO(), "TimerFSM closed")
				return
			}
		}
	}
}
