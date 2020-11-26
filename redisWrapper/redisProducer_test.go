package redisWrapper

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

//func TestRedisProducer(t *testing.T) {
func TestRedisProducer(t *testing.T) {
	var executionRequestWrapper ExecutionRequestWrapper
	executionRequest := ExecutionRequest{
		ExecutionId:   "qqskfskafjlsfslffsljfs",
		WorkflowId:    "12324566545455556",
		Authorization: "",
		Environments:  []string{"shuffle", "shuffle1"},
	}
	executionRequestWrapper.Data = append(executionRequestWrapper.Data, executionRequest)
	data_byte, err := json.Marshal(&executionRequestWrapper)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("Failed executionrequest in queue marshaling")
		return
	}

	logrus.SetLevel(logrus.TraceLevel) //设置Trace以上的信息都显示

	cfg := []RedisConfig{
		{RedisServerIP: "localhost", RedisConnType: "tcp", RedisServerPort: "6379", RedisServerPass: "shuffle123"},
		{RedisServerIP: "10.1.5.68", RedisConnType: "tcp", RedisServerPort: "6379", RedisServerPass: "123456"},
	}
	redisCli, err := NewClient(cfg)
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("new redis client Failed")
		return
	}
	for {
		mqName := "workflowqueueShuffle"
		err = redisCli.Push(mqName, string(data_byte))
		if err != nil {
			logrus.WithFields(logrus.Fields{"error": err}).Error("redis Push Failed")
			return
		}
		time.Sleep(1 * time.Second)
	}
}
