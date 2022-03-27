package client

import (
	"fmt"
	"testing"

	"octopus/task-agent/common"
)

func TestSubmitTask(t *testing.T) {
	addr := "192.168.0.107:1234"
	req := &common.CmdRequest{
		Name:  "test",
		Id:    "",
		Nodes: []string{"192.168.0.107"},
		Cmd:   "ls -al; pwd",
		Type:  "",
	}

	err := SubmitTask(addr, req)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
}
