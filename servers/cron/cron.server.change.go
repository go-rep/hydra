package cron

import (
	"fmt"

	"github.com/micro-plat/hydra/registry/conf/server/metric"
	"github.com/micro-plat/hydra/registry/conf/server/task"
)

//SetMetric 重置metric
func (s *CronServer) SetMetric(metric *metric.Metric) error {
	s.metric.Stop()
	if metric.Disable {
		return nil
	}
	if err := s.metric.Restart(metric.Host, metric.DataBase, metric.UserName, metric.Password, metric.Cron, s.Logger); err != nil {
		err = fmt.Errorf("metric设置有误:%v", err)
		return err
	}
	return nil
}

//SetTasks 设置定时任务
func (s *CronServer) SetTasks(tasks []*task.Task) (err error) {
	s.Processor, err = s.getProcessor()
	if err != nil {
		return err
	}
	return s.Processor.Add(tasks...)
}

//ShowTrace 显示跟踪信息
func (s *CronServer) ShowTrace(b bool) {
	s.conf.Set("show-trace", b)
	return
}
