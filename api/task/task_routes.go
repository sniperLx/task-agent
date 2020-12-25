package task

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/sniperLx/task-agent/common"
	"github.com/sniperLx/task-agent/engine"
	"net"
	"net/http"
)

func (tr *taskRouter) submitCmdTask(write http.ResponseWriter, request *http.Request) {
	bodyArr := make([]byte, request.ContentLength)
	_, _ = request.Body.Read(bodyArr)
	logrus.Debugf("%s\n", string(bodyArr))

	cmdReq := &common.CmdRequest{}
	if err := json.Unmarshal(bodyArr, cmdReq); err != nil {
		logrus.Errorf("parse task failed: %v", err)
		write.WriteHeader(404)
		return
	}
	taskId := cmdReq.Id
	if taskId == "" {
		if _, err := uuid.Parse(taskId); err != nil {
			logrus.Infof("generate uuid for task: %v", cmdReq.Name)
			taskId = uuid.New().String()
		}
	}

	cmdTask := &engine.CmdTask{
		Name: cmdReq.Name,
		Id:   taskId,
		Cmd:  cmdReq.Cmd,
	}

	if err := checkParam(cmdReq, cmdTask); err != nil {
		reply := fmt.Sprintf("invalid request body: %v", err)
		_, _ = write.Write([]byte(reply))
		write.WriteHeader(404)
		return
	}

	logrus.Debugf("submit task: %v", *cmdTask)

	engine.SubmitTask(cmdTask)

	cmdRet := &common.CmdReply{
		Id: cmdTask.GetId(),
	}
	replyArr, err := json.Marshal(cmdRet)
	if err != nil {
		logrus.Errorf("encode reply failed: %v", err)
		write.WriteHeader(500)
		return
	}

	if _, err := write.Write(replyArr); err != nil {
		logrus.Errorf("write reply failed: %v", err)
		write.WriteHeader(500)
		return
	}
}

func checkParam(request *common.CmdRequest, cmdTask *engine.CmdTask) error {
	nodes := request.Nodes
	//检测所有node的是否为标准ip格式
	localIpStr := *common.LocalIp
	localIp := net.ParseIP(localIpStr)
	if localIp == nil {
		logrus.Panicf("wrong LOCAL_IP value: %s", localIpStr)
	}
	var size int
	var isTarget bool
	nodesSet := make(map[string]bool)
	for _, node := range nodes {
		ip := net.ParseIP(node)
		if ip == nil {
			return fmt.Errorf("invalid ip: %v", node)
		}
		if ip.Equal(localIp) {
			isTarget = true
		} else if _, ok := nodesSet[node]; !ok {
			nodesSet[node] = true
			size++
		}
	}
	//remove duplicate node
	ips := make([]string, 0, size)
	for node := range nodesSet {
		ips = append(ips, node)
	}
	cmdTask.IsTarget = isTarget
	cmdTask.Nodes = &engine.TaskNodes{
		Ips:  ips,
		Size: size,
	}

	logrus.Debugf("exclude local ip, cmdTask Nodes info is: %v", cmdTask.Nodes)

	return nil
}

func (tr *taskRouter) cancelCmdTask(write http.ResponseWriter, request *http.Request) {
	write.WriteHeader(501)
	//todo 任务的取消，由于是分布式任务，因此可能有些节点已经执行完成，尽力取消
	//发送取消指令给主节点，节点收到后，会查询之前子任务状态，从节点如果没执行完，则会取消执行。
}

func (tr *taskRouter) queryCmdTask(write http.ResponseWriter, request *http.Request) {
	//todo 查询任务状态
	write.WriteHeader(501)
}

func (tr *taskRouter) checkHealth(write http.ResponseWriter, request *http.Request) {
	if _, err := write.Write([]byte("ok")); err != nil {
		logrus.Errorf("err: %v", err)
		write.WriteHeader(500)
	}
	write.WriteHeader(200)
}
