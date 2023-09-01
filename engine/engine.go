package engine

import (
	"encoding/json"
	"fmt"
	"os"

	"octopus/task-agent/common"
)

var execEngine *taskEngine

type Event int

const (
	DataAdd    Event = 1
	DataRemove Event = -1
)

type taskEngine struct {
	//存放未执行的任务
	taskQueue Queue

	//存放已执行任务的结果
	retQueue Queue

	taskEventChan chan Event
	retEventChan  chan Event
	kafkaTank     RetTank
	status        bool
}

func InitTaskEngine() *taskEngine {
	if execEngine != nil {
		return execEngine
	}
	execEngine = &taskEngine{
		taskQueue:     NewBlockingQueue(),
		retQueue:      NewBlockingQueue(),
		taskEventChan: make(chan Event),
		retEventChan:  make(chan Event),
		kafkaTank:     NewKafkaTank(),
	}

	err := execEngine.kafkaTank.Init()
	if err != nil {
		logger.Panicf("init kafka producer failed: %v", err)
	}

	return execEngine
}

//启动执行引擎
func StartEngine(sigChan chan os.Signal) {
	go pollTask(sigChan)
	pollRet(sigChan)
}

func SubmitTask(task Task) {
	execEngine.taskQueue.Offer(task)
	logger.Debugf("submit task %v success", task.GetId())
}

func ReportRet(ret Task) {
	execEngine.retQueue.Offer(ret)
	logger.Debugf("report result of task %v success", ret.GetId())
}

func emitDataAddEvent(eventChan chan Event) {
	go func() {
		eventChan <- DataAdd
	}()
}

func emitDataRemoveEvent(eventChan chan Event) {
	go func() {
		eventChan <- DataRemove
	}()
}

func pollRet(sigChan chan os.Signal) {
	retQ := execEngine.retQueue
	for ret := retQ.Poll(); ; {
		logger.Debugf("get ret of task %v from ret Queue: %v", ret.GetId(), ret)
		go reportRetToKafka(ret)
		go reportRetToStdout(ret)
		select {
		case status := <-sigChan:
			logger.Infof("stop polling task result after receive signal %v", status)
			break
		}
	}
}

//监听任务队列，取出任务交给一个协程执行
func pollTask(sigChan chan os.Signal) {
	taskQ := execEngine.taskQueue
	for task := taskQ.Poll(); ; {
		logger.Debugf("get task %v from task Queue with len %d: %v", task.GetId(), taskQ.Size(), task)
		//todo some task may dont support running in parallel
		go do(task)
		select {
		case status := <-sigChan:
			//todo finish left tasks in queue and exit
			logger.Infof("stop polling task after receive signal %v", status)
			break
		}
	}
}

func do(task Task) {
	//协程执行完后，将结果存放到结果队列
	switch v := task.(type) {
	case *CmdTask:
		logger.Debugf("start to do task: %v", task.GetId())
		ret := task.(*CmdTask).Do(common.TaskTimeout)
		if ret != nil {
			logger.Debugf("result is: %v", ret)
			ReportRet(ret)
		}
	default:
		logger.Errorf("this task type is not supported yet: %v", v)
	}
}

//上报任务执行结果到kafka
func reportRetToKafka(ret Task) {
	//监听结果队列，上报协程会将结果写到对应的kafka topic中
	bytes, err := json.Marshal(ret)
	if err != nil {
		logger.Errorf("send msg failed: %v", err)
		return
	}
	msg := string(bytes)
	logger.Debug("sending message: %v", msg)
	err = execEngine.kafkaTank.IssueRet(msg)
	if err != nil {
		logger.Errorf("failed to send ret of task %s: %v", ret.GetId(), err)
	}
}

func reportRetToStdout(ret Task) {
	//logger.Infof("result of %s is %s", ret.GetId(), ret.(*CmdResult).Ret)
	fmt.Printf("ret of task %s is: %v\n", ret.GetId(), ret)
}
