package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sniperLx/task-agent/common"
	"net"
	"net/http"
	"time"
)

//https://golang.org/pkg/net/http/
//submit task to target by call rest api
//addr: host:port
func SubmitTask(addr string, request *common.CmdRequest) error {
	client := InitHttpClient()
	defer client.CloseIdleConnections()

	url := fmt.Sprintf("http://%s/submitCmdTask", addr)
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buff := make([]byte, resp.ContentLength)
	_, _ = resp.Body.Read(buff)
	logrus.Debug(fmt.Sprintf("%v", string(buff)))
	return nil
}

func CheckHealth(addr string) error {
	client := InitHttpClient()
	defer client.CloseIdleConnections()

	url := fmt.Sprintf("http://%s/checkHealth", addr)
	logrus.Debugf("will call %v to check health", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func InitHttpClient() *http.Client {
	direct := &net.Dialer{
		Timeout:   15 * time.Second,
		KeepAlive: 15 * time.Second,
	}

	// TODO(dmcgowan): Call close idle connections when complete, use keep alive
	transport := &http.Transport{
		//Proxy:             http.ProxyFromEnvironment,
		DialContext:       direct.DialContext,
		DisableKeepAlives: true,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}
}
