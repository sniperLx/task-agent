package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"octopus/task-agent/api"
	"octopus/task-agent/common"
	"octopus/task-agent/engine"
	logg "octopus/task-agent/log"
	"octopus/task-agent/register"
)

var logger = logg.InitLogger()

//./task-agent -kafka-brokers 192.168.0.107:9092 -kafka-heartbeat-topic test -kafka-result-topic result -debug true
// -local-ip 192.168.0.107
func main() {
	//parse command line params
	common.ParseAndInitCmdParams()
	if *common.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	//start heartbeat register
	if err := register.StartDefaultRegister(); err != nil {
		logger.Errorf("will exit, as start heartbeat failed: %s", err)
		return
	}

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	//启动执行引擎
	engine.InitTaskEngine()
	go engine.StartEngine(sigChan)

	//启动api server
	server := api.NewServer(common.ServerConfig)
	server.ServeAPI()
}
