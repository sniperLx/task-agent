package register

import (
	"fmt"
	"github.com/sniperLx/task-agent/common"
	"testing"
)

func Test_externalIp(t *testing.T) {
	ips, err := getAllExternalIps()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	fmt.Printf("%v", ips)
}

func Test_getLocalIp(t *testing.T) {
	mockIp := "192.168.0.1:9092"
	common.KafkaBrokers = &mockIp

	localIp, err := getLocalIp()
	if err != nil {
		fmt.Printf("err: %v", err)
	}
	fmt.Printf("localIp: %v", localIp)
}