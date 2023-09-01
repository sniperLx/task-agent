package register

import (
	"time"
)

type HeartBeatRegister interface {
	Init() error
	SendHeartbeatPeriodic(time time.Duration)
	SendHeartbeat()
}

func StartDefaultRegister() error {
	register := NewKafkaRegister()

	err := register.Init()
	if err != nil {
		return err
	}

	go register.SendHeartbeatPeriodic(60 * time.Second)
	return nil
}
