package engine

import (
	"fmt"
	"github.com/sirupsen/logrus"
	httpclient "github.com/sniperLx/task-agent/client"
	"github.com/sniperLx/task-agent/common"
	"os/exec"
	"time"
)

type Task interface {
	GetId() string
}

type TaskNodes struct {
	Ips  []string
	Size int
}

type CmdTask struct {
	Name string
	//全局唯一
	Id    string
	Nodes *TaskNodes
	Cmd   string
	//任务是否在本节点执行
	IsTarget bool
	//读任务，写任务
	Type string
}

type CmdResult struct {
	Name     string `json:"name"`
	Id       string `json:"id"`
	Node     string `json:"node"`
	IsTarget bool   `json:"isTarget"` //任务是否在本节点执行
	Ret      string `json:"ret"`
}

//本地任务执行器：
//解析任务包含的node列表，剔除本节点ip，从剩下的节点中选择两个节点，探测连通性
//如果连通，则将节点分为两部分发送给这两个节点，提交任务，保存得到的任务id，用于后面跟踪这些子任务的进度
//如果不连通，保存起来，后面上报给调用者，重新选ip来提交子任务
//节点列表中保护本节点，则也执行这个任务，执行玩后将结果上报给kafka，同时定期查看之前分发子任务进度，如果有问题则重新调度。
//每个节点的行为相同：执行任务和调度子任务
func (ct *CmdTask) Do(timeout time.Duration) Task {
	//用 *CmdResult，返回nil, engine侧chan收到的值变为了<nil>，不知道为什么
	waitRetChan := make(chan Task)
	defer close(waitRetChan)
	localIp := *common.LocalIp

	go func() {
		logrus.Debugf("will do this task %v locally: %v", ct.GetId(), ct.IsTarget)
		//调度任务到其他节点
		ct.Schedule()
		if ct.IsTarget {
			ct.localDo(waitRetChan)
		} else {
			waitRetChan <- nil
		}
	}()

	select {
	case ret := <-waitRetChan:
		return ret
	case <-time.After(timeout):
		logrus.Errorf("timeout to exec task %s at %s: %s", ct.GetId(), localIp, ct.Cmd)
		return &CmdResult{
			Name:     ct.Name,
			Id:       ct.Id,
			Node:     localIp,
			IsTarget: ct.IsTarget,
			Ret:      fmt.Sprintf("do task timeout after %v seconds", timeout),
		}
	}
}

func (ct *CmdTask) localDo(retChan chan Task) {
	logrus.Debugf("start to exec cmd: %v", ct.Cmd)
	cmd := exec.Command("/bin/sh", "-c", ct.Cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("task %s exec failed: %v", ct.GetId(), err)
		return
	}
	logrus.Debugf("task %s output: %s", ct.GetId(), out)

	localIp := *common.LocalIp
	ret := &CmdResult{
		Name:     ct.Name,
		Id:       ct.Id,
		Node:     localIp,
		IsTarget: true,
		Ret:      fmt.Sprintf("%s", out),
	}

	retChan <- ret
}

func (ct *CmdTask) Schedule() {
	localIp := *common.LocalIp
	size := ct.Nodes.Size
	if size <= 0 {
		return
	} else if size == 1 {
		node := ct.Nodes.Ips[0]
		if err := httpclient.CheckHealth(getAddr(node)); err != nil {
			errNodeRet := &CmdResult{
				Name:     ct.Name,
				Id:       ct.Id,
				Node:     node,
				IsTarget: true,
				Ret:      fmt.Sprintf("%s check health to %s failed, %s", localIp, node, err),
			}
			ReportRet(errNodeRet)
		}
		request := &common.CmdRequest{
			Name:  ct.Name,
			Id:    ct.Id,
			Nodes: ct.Nodes.Ips,
			Cmd:   ct.Cmd,
			Type:  ct.Type,
		}
		_ = httpclient.SubmitTask(getAddr(node), request)
		return
	}
	idx := size / 2
	subTask1 := &common.CmdRequest{
		Name:  ct.Name,
		Id:    ct.Id,
		Nodes: ct.Nodes.Ips[:idx],
		Cmd:   ct.Cmd,
		Type:  ct.Type,
	}
	subTask2 := &common.CmdRequest{
		Name:  ct.Name,
		Id:    ct.Id,
		Nodes: ct.Nodes.Ips[idx:],
		Cmd:   ct.Cmd,
		Type:  ct.Type,
	}
	if node1, err := selectNode(subTask1.Nodes, localIp); err == nil {
		_ = httpclient.SubmitTask(getAddr(node1), subTask1)
	} else {
		logrus.Errorf("%v", err)
	}

	if node2, err := selectNode(subTask2.Nodes, localIp); err == nil {
		_ = httpclient.SubmitTask(getAddr(node2), subTask2)
	} else {
		logrus.Errorf("%v", err)
	}
}

func getAddr(node string) string {
	return fmt.Sprintf("%s:%v", node, common.ServerConfig.Port)
}

func selectNode(nodes []string, localIp string) (string, error) {
	if len(nodes) == 0 {
		return "", fmt.Errorf("empty nodes list to select from")
	}

	for _, node := range nodes {
		if err := httpclient.CheckHealth(getAddr(node)); err == nil {
			return node, nil
		}
	}

	return "", fmt.Errorf("check health from %s to all nodes %v failed", localIp, nodes)
}

func (ct *CmdTask) GetId() string {
	return ct.Id
}

func (cr *CmdResult) GetId() string {
	return cr.Id
}
