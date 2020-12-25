package engine

import (
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
	"github.com/sniperLx/task-agent/common"
	"strings"
	"time"
)

type RetTank interface {
	Init() error
	IssueRet(string) error
}

type kafkaTank struct {
	brokers     []string
	topic    string
	producer sarama.SyncProducer
}

func NewKafkaTank() RetTank {
	return &kafkaTank{}
}

func (kc *kafkaTank) Init() error {
	brokers := strings.Split(*common.KafkaBrokers, ",")
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return err
	}
	kc.brokers = brokers
	kc.topic = *common.KafkaTopicForResult
	kc.producer = producer
	logrus.Debugf("kafkaTank info: %v", kc)
	return nil
}

func (kc *kafkaTank) IssueRet(msg string) error {
	kafkaMsg := &sarama.ProducerMessage{
		Topic:     kc.topic,
		Value:     sarama.StringEncoder(msg),
		Timestamp: time.Now(),
	}

	_, _, err := kc.producer.SendMessage(kafkaMsg)
	return err
}
