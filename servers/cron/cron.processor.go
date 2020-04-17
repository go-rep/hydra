package cron

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/micro-plat/hydra/registry/conf/server/task"
	"github.com/micro-plat/hydra/servers"
	"github.com/micro-plat/hydra/servers/pkg/dispatcher"
	"github.com/micro-plat/lib4go/concurrent/cmap"
	"github.com/micro-plat/lib4go/utility"
)

//Processor cron管理程序，用于管理多个任务的执行，暂停，恢复，动态添加，移除
type Processor struct {
	*dispatcher.Dispatcher
	lock      sync.Mutex
	once      sync.Once
	done      bool
	closeChan chan struct{}
	engine    servers.IRegistryEngine
	length    int
	index     int
	span      time.Duration
	slots     []cmap.ConcurrentMap //time slots
	startTime time.Time
	isPause   bool
}

//NewProcessor 创建processor
func NewProcessor(engine servers.IRegistryEngine) (p *Processor, err error) {
	p = &Processor{
		engine:     engine,
		Dispatcher: dispatcher.New(),
		closeChan:  make(chan struct{}),
		span:       time.Second,
		length:     60,
		startTime:  time.Now(),
	}
	p.slots = make([]cmap.ConcurrentMap, p.length, p.length)
	for i := 0; i < p.length; i++ {
		p.slots[i] = cmap.New(2)
	}
	return p, nil
}

//Start 所有任务
func (s *Processor) Start() error {
START:
	for {
		select {
		case <-s.closeChan:
			break START
		case <-time.After(s.span):
			s.execute()
		}
	}
	return nil
}

//Add 添加任务
func (s *Processor) Add(ts ...*task.Task) (err error) {
	for _, t := range ts {
		task, err := newCronTask(t, s.engine)
		if err != nil {
			return fmt.Errorf("构建cron.task失败:%v", err)
		}
		if !s.Dispatcher.Find(task.GetService()) {
			s.Dispatcher.Handle(task.GetMethod(), task.GetService(), task.GetHandler().(dispatcher.HandlerFunc))
		}
		if _, _, err := s.add(task); err != nil {
			return err
		}
	}
	return

}
func (s *Processor) add(task iCronTask) (offset int, round int, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.done {
		return -1, -1, nil
	}
	now := time.Now()
	nextTime := task.NextTime(now)
	if nextTime.Sub(now) < 0 {
		return -1, -1, errors.New("next time less than now.1")
	}
	offset, round = s.getOffset(now, nextTime)
	if offset < 0 || round < 0 {
		return -1, -1, errors.New("next time less than now.2")
	}
	task.SetRound(round)
	s.slots[offset].Set(utility.GetGUID(), task)
	return
}

//Remove 移除服务
func (s *Processor) Remove(name string) {
	if name == "" {
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, slot := range s.slots {
		slot.RemoveIterCb(func(k string, value interface{}) bool {
			task := value.(iCronTask)
			task.SetDisable()
			return task.GetName() == name
		})
	}
}

//Pause 暂停所有任务
func (s *Processor) Pause() error {
	s.isPause = true
	return nil
}

//Resume 恢复所有任务
func (s *Processor) Resume() error {
	s.isPause = false
	return nil
}

//Close 退出
func (s *Processor) Close() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.done = true
	s.once.Do(func() {
		close(s.closeChan)
	})
}

//-------------------------------------内部处理------------------------------------

func (s *Processor) getOffset(now time.Time, next time.Time) (pos int, circle int) {
	d := next.Sub(now) //剩余时间
	delaySeconds := int(d/1e9) + 1
	intervalSeconds := int(s.span.Seconds())
	circle = int(delaySeconds / intervalSeconds / s.length)
	pos = int(s.index+delaySeconds/intervalSeconds) % s.length
	return
}

func (s *Processor) execute() {
	s.startTime = time.Now()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.index = (s.index + 1) % s.length
	current := s.slots[s.index]
	current.RemoveIterCb(func(k string, value interface{}) bool {
		task := value.(iCronTask)
		task.ReduceRound(1)
		if task.GetRound() <= 0 {
			go s.handle(task)
			return true
		}
		return false
	})
}
func (s *Processor) handle(task iCronTask) error {
	if s.done || !task.Enable() {
		return nil
	}
	if !s.isPause {
		task.AddExecuted()
		rw, err := s.Dispatcher.HandleRequest(task)
		if err != nil {
			task.Errorf("%s执行出错:%v", task.GetName(), err)
		}
		task.SetResult(rw.Status(), rw.Data())
	}
	_, _, err := s.add(task)
	if err != nil {
		return err
	}
	return nil

}
