package register

import (
	"strings"
	"time"

	"octopus/task-agent/common"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

type kafkaRegister struct {
	brokers  []string
	topic    string
	producer sarama.SyncProducer
}

func NewKafkaRegister() HeartBeatRegister {
	return &kafkaRegister{}
}

func (kr *kafkaRegister) Init() error {
	brokers := strings.Split(*common.KafkaBrokers, ",")
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return err
	}
	kr.brokers = brokers
	kr.topic = *common.KafkaTopicForRegister
	kr.producer = producer
	logrus.Debugf("kafkaRegister info: %v", kr)
	return nil
}

func (kr *kafkaRegister) SendHeartbeatPeriodic(period time.Duration) {
	logrus.Debugf("start to report heartbeat every %v", period)
	for {
		select {
		case <-time.After(period):
			kr.SendHeartbeat()
		}
	}
}

func (kr *kafkaRegister) SendHeartbeat() {
	localIp := *common.LocalIp

	heartbeatMsg := &sarama.ProducerMessage{
		Topic:     kr.topic,
		Value:     sarama.StringEncoder(localIp),
		Timestamp: time.Now(),
	}

	_, _, err := kr.producer.SendMessage(heartbeatMsg)
	if err != nil {
		logrus.Warnf("send heartbeat err: %v", err)
	}
}
