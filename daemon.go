package main

import (
	"github.com/sirupsen/logrus"
	"github.com/sniperLx/task-agent/api"
	"github.com/sniperLx/task-agent/api/task"
	"github.com/sniperLx/task-agent/common"
	engine "github.com/sniperLx/task-agent/engine"
	"github.com/sniperLx/task-agent/register"
	"os"
	"time"
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

	//启动执行引擎
	engineWait := make(chan error)
	engine.InitTaskEngine()
	go engine.StartEngine(engineWait)

	//启动api server
	server := api.NewServer(common.ServerConfig)
	server.Accept(common.ServerConfig)
	registerRoutes(server)
	serveAPIWait := make(chan error)
	go server.ServeAPI(serveAPIWait)

	//start heartbeat register
	heartbeatRegister := register.NewKafkaRegister()
	err := heartbeatRegister.Init()
	if err != nil {
		logrus.Panicf("start heartbeat register failed: %v", err)
		//todo send signal to stop api server and task engine
	}
	go heartbeatRegister.SendHeartbeatPeriodic(60 * time.Second)

	select {
	case serverErr := <-serveAPIWait:
		if serverErr != nil {
			logrus.Errorf("agent server exit due to error: %v", serverErr)
		}
	case engineErr := <-engineWait:
		if engineErr != nil {
			logrus.Errorf("agent engine exit due to error: %v", engineErr)
		}
	}
}

func initLog() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.WarnLevel)
	logrus.SetReportCaller(true)
}

func registerRoutes(server *api.Server) {
	routers := []api.Router{
		task.NewRouter(),
	}

	server.InitRouter(routers...)
}
