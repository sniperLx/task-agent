package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"octopus/task-agent/api"
	"octopus/task-agent/common"
	"octopus/task-agent/engine"
	"octopus/task-agent/register"

	"github.com/sirupsen/logrus"
)

//./task-agent -kafka-brokers 192.168.0.107:9092 -kafka-heartbeat-topic test -kafka-result-topic result -debug true
// -local-ip 192.168.0.107
func main() {
	initLog()

	//parse command line params
	common.ParseAndInitCmdParams()
	if *common.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	//start heartbeat register
	heartbeatRegister := register.NewKafkaRegister()
	err := heartbeatRegister.Init()
	if err != nil {
		logrus.Errorf("will exit, as start heartbeat register failed: %s", err)
		return
	}

	go heartbeatRegister.SendHeartbeatPeriodic(60 * time.Second)

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	//启动执行引擎
	engine.InitTaskEngine()
	go engine.StartEngine(sigChan)

	//启动api server
	server := api.NewServer(common.ServerConfig)
	server.Accept(common.ServerConfig)
	registerRoutes(server)
	server.ServeAPI()
}

func initLog() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.WarnLevel)
	logrus.SetReportCaller(true)
}

func registerRoutes(server *api.Server) {
	routers := []api.Router{
		api.NewRouter(),
	}

	server.InitRouter(routers...)
}
