package engine

import (
	"encoding/json"
	"fmt"
	"os"

	"octopus/task-agent/common"

	"github.com/sirupsen/logrus"
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

	kafkaTank RetTank
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
		logrus.Panicf("init kafka producer failed: %v", err)
	}

	return execEngine
}

//启动执行引擎
func StartEngine(sigChan chan os.Signal) {
	taskChan := make(chan Task)
	retChan := make(chan Task)

	go pollTask(taskChan)
	go pollRet(retChan)

	for {
		select {
		case task := <-taskChan:
			go do(task)
		case ret := <-retChan:
			go reportRetToKafka(ret)
			go reportRetToStdout(ret)
		case status := <-sigChan:
			logrus.Infof("task engine exit after receive signal %v", status)
			break;
		}
	}
}

func SubmitTask(task Task) {
	execEngine.taskQueue.Offer(task)
	emitDataAddEvent(execEngine.taskEventChan)
	logrus.Debugf("submit task %v success", task.GetId())
}

func ReportRet(ret Task) {
	execEngine.retQueue.Offer(ret)
	emitDataAddEvent(execEngine.retEventChan)
	logrus.Debugf("report result of task %v success", ret.GetId())
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

func pollRet(retChan chan Task) {
	retQ := execEngine.retQueue
	for {
		select {
		case val := <-execEngine.retEventChan:
			if val == DataAdd {
				if ele := retQ.Poll(); ele != nil {
					logrus.Debugf("get ret of task %v from ret Queue: %v", ele.GetId(), ele)
					retChan <- ele
					emitDataRemoveEvent(execEngine.retEventChan)
				}
			}
		}
	}
}

//监听任务队列，取出任务交给一个协程执行
func pollTask(taskChan chan Task) {
	taskQ := execEngine.taskQueue
	for {
		select {
		case val := <-execEngine.taskEventChan:
			logrus.Debugf("value: %v", val)
			if val == DataAdd {
				logrus.Debugf("taskQ size: %d", taskQ.Size())
				if ele := taskQ.Poll(); ele != nil {
					logrus.Debugf("get task %v from task Queue: %v", ele.GetId(), ele)
					taskChan <- ele
					emitDataRemoveEvent(execEngine.taskEventChan)
				}
			}
		}
	}
}

func do(task Task) {
	//协程执行完后，将结果存放到结果队列
	switch v := task.(type) {
	case *CmdTask:
		logrus.Debugf("start to do task: %v", task.GetId())
		ret := task.(*CmdTask).Do(common.TaskTimeout)
		if ret != nil {
			logrus.Debugf("result is: %v", ret)
			ReportRet(ret)
		}
	default:
		logrus.Errorf("this task type is not supported yet: %v", v)
	}
}

//上报任务执行结果到kafka
func reportRetToKafka(ret Task) {
	//监听结果队列，上报协程会将结果写到对应的kafka topic中
	bytes, err := json.Marshal(ret)
	if err != nil {
		logrus.Errorf("send msg failed: %v", err)
		return
	}
	msg := string(bytes)
	logrus.Debug("sending message: %v", msg)
	err = execEngine.kafkaTank.IssueRet(msg)
	if err != nil {
		logrus.Errorf("failed to send ret of task %s: %v", ret.GetId(), err)
	}
}

func reportRetToStdout(ret Task) {
	//logrus.Infof("result of %s is %s", ret.GetId(), ret.(*CmdResult).Ret)
	fmt.Printf("ret of task %s is: %v\n", ret.GetId(), ret)
}
