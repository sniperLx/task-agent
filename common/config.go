package common

import (
	"errors"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	LOCAL_IP     = "LOCAL_IP"
	TASK_TIMEOUT = 10 * time.Second
)

var (
	Debug = flag.Bool("debug", false, "set debug model to print more log, default value is false")
	Host = flag.String("host", "0.0.0.0", "host ip used by agent to monitor incoming request, and use 0.0.0.0 by default")
	LocalIp = flag.String("local-ip", os.Getenv("LOCAL_IP"), "ip of current node used to communicate with other nodes and report heartbeat")
	KafkaBrokers = flag.String("kafka-brokers", os.Getenv("KAFKA_PEERS"), "The Kafka brokers to connect to store task result, as a comma separated list")
	KafkaTopicForResult = flag.String("kafka-result-topic", os.Getenv("KAFKA_RESULT_TOPIC"), "The Kafka topic to store the result of all tasks")
	KafkaTopicForRegister = flag.String("kafka-heartbeat-topic", os.Getenv("KAFKA_HEARTBEAT_TOPIC"), "The Kafka topic to store ip info of all registered nodes")
	Port = flag.Int("port", 12357, "port used by agent to monitor incoming request, default value is 12357")
	ServerConfig *Config
)

type Config struct {
	Debug bool
	Host  string
	Port  int
	LocalIp string
}

func ParseAndInitCmdParams() {
	flag.Parse()
	if *LocalIp == "" {
		localIp, err := GetLocalIp()
		if err != nil {
			logrus.Panicf("cannot get local ip: %v", err)
		}
		LocalIp = &localIp
		//_ = flag.Set("local-ip", localIp)
	}

	ServerConfig = &Config{
		Debug:   false,
		Host:    *Host,
		Port:    *Port,
		LocalIp: *LocalIp,
	}
}

func GetLocalIp() (string, error) {
	localIp := os.Getenv(LOCAL_IP)
	if localIp != "" {
		return localIp, nil
	}
	kafkaBrokers := strings.Split(*KafkaBrokers, ",")
	kafkaBrokerIp := strings.Split(kafkaBrokers[0], ":")[0]
	conn, err := net.Dial("udp", fmt.Sprintf("%s:80", kafkaBrokerIp))
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	defer conn.Close()
	//how to choose right localIp? we choose the one which can reach kafka ip
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	localIp = localAddr.IP.String()
	logrus.Debugf("get local ip: %v", localIp)
	_ = os.Setenv(LOCAL_IP, localIp)
	return localIp, nil
}

func GetAllExternalIps() ([]net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, 0, 5)
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			ips = append(ips, ip)
		}
	}
	if len(ips) > 0 {
		return ips, nil
	}
	return nil, errors.New("seem like don't connected to the network")
}