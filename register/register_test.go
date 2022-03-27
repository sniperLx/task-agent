package register

import (
	"fmt"
	"testing"

	"octopus/task-agent/common"
)

func Test_externalIp(t *testing.T) {
	ips, err := common.GetAllExternalIps()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	fmt.Printf("%v", ips)
}

func Test_getLocalIp(t *testing.T) {
	mockIp := "192.168.0.1:9092"
	common.KafkaBrokers = &mockIp

	localIp, err := common.GetLocalIp()
	if err != nil {
		fmt.Printf("err: %v", err)
	}
	fmt.Printf("localIp: %v", localIp)
}
