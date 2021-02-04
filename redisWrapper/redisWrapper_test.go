package redisWrapper

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

type ExecutionRequestWrapper struct {
	Data []ExecutionRequest `json:"data"`
}

type ExecutionRequest struct {
	ExecutionId       string   `json:"execution_id"`
	ExecutionArgument string   `json:"execution_argument"`
	ExecutionSource   string   `json:"execution_source"`
	WorkflowId        string   `json:"workflow_id"`
	Environments      []string `json:"environments"`
	Authorization     string   `json:"authorization"`
	Status            string   `json:"status"`
	Start             string   `json:"start"`
	Type              string   `json:"type"`
}

func TestRedisWrapper(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel) //设置Trace以上的信息都显示
	testing := []string{"qqqq", "sssss"}
	tmp := fmt.Sprintf("%s", testing)
	logrus.Printf(tmp)
	cfg := []RedisConfig{
		{RedisServerIP: "localhost", RedisConnType: "tcp", RedisServerPort: "6379", RedisServerPass: "shuffle123"},
	}
	redisCli, err := NewClient(cfg)
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("new redis client Failed")
		return
	}
	// test Hset and Hget
	WorkflowHashMapName := "shuffleWorkflowTest"
	keys := []string{"abc", "abc1", "abc2"}
	message := "this is a book test1"
	for _, key := range keys {
		redisCli.Hset(WorkflowHashMapName, key, message)
	}

	value, _ := redisCli.Hget(WorkflowHashMapName, keys[0])
	logrus.WithFields(logrus.Fields{"value": value}).Info("redis Hget value")

	testlzz, err := redisCli.Hmget(WorkflowHashMapName, keys)
	logrus.Println(testlzz)
	testlzz2, err := redisCli.HVals(WorkflowHashMapName)
	logrus.Println(testlzz2)
	err = redisCli.Hdel(WorkflowHashMapName, keys[0])
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("redis Hdel Failed")
	}
	//test get keys
	pattern := "workflowqueue*"
	keys, err = redisCli.GetKeys(pattern)
	logrus.WithFields(logrus.Fields{"keys": keys}).Info("redis Getkeys value")

	// test mq
	mqName := "mqtest"
	err = redisCli.Push(mqName, "this is the test message")
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("redis Push Failed")
		return
	}

	// test lock && unlock
	lockKey := "qqqq"
	lockKey1 := "qqqq1" // dead lock
	keys, err = redisCli.HKeys("shuffle_workflowexecutions_new1")
	logrus.Println(keys, err)
	err = redisCli.Lock(lockKey, "qqq", 30000)
	err = redisCli.Lock(lockKey1, "qqq1", 0)
	err = redisCli.UnLock(lockKey)

	// test set operations
	setName := "set1"
	err = redisCli.Sadd(setName, "test1")
	err = redisCli.Sadd(setName, "test2")
	err = redisCli.Sadd(setName, "test3")

	temp1, err := redisCli.Smembers(setName)
	logrus.Println(temp1)
	err = redisCli.Srem(setName, "test3")
	// test push
	err = redisCli.Push("mqtest1", "this is the test message")
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("redis Push Failed")
		return
	}
	// data, err := redisCli.Pop(mqName)
	// logrus.WithFields(logrus.Fields{"value": data}).Info("redis Pop value")
	// if err != nil {
	// 	logrus.WithFields(logrus.Fields{"error": err}).Error("redis Pop Failed")
	// 	return
	// }
	timeout := uint32(0) // unit:s
	mqNameList := []string{"mqtest1", "mqtest"}
	data, err := redisCli.Bpop(mqNameList, timeout) //bpop阻塞读取1s，如果没有读取到则返回空；当timeout为0的时候则是无穷大一直阻塞
	logrus.WithFields(logrus.Fields{"value": data}).Info("redis BPop value")
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("redis Bpop Failed")
		return
	}
	redisCli.Close()
}

func TestRedisWrapperBpop(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel) //设置Trace以上的信息都显示

	cfg := []RedisConfig{{RedisServerIP: "localhost", RedisConnType: "tcp", RedisServerPort: "6379", RedisServerPass: "shuffle123"}}
	redisCli, err := NewClient(cfg)
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("new redis client Failed")
		return
	}
	defer redisCli.Close()
	mqNameList := []string{"workflowqueueShuffle", "mqtest"}
	timeout := uint32(0) // unit:s
	for {
		data, err := redisCli.Bpop(mqNameList, timeout) //bpop阻塞读取1s，如果没有读取到则返回空；当timeout为0的时候则是无穷大一直阻塞
		if err != nil {
			logrus.WithFields(logrus.Fields{"error": err}).Error("redis Bpop Failed")
			continue
		}
		logrus.WithFields(logrus.Fields{"value": data}).Info("redis BPop value")
	}

}

func TestRedisExecutionReq(t *testing.T) {
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

	cfg := []RedisConfig{{RedisServerIP: "localhost", RedisConnType: "tcp", RedisServerPort: "6379", RedisServerPass: "shuffle123"}}
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
		time.Sleep(2 * time.Second)
	}
}

func TestCheckRedisConPeriod(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)
	cfg := []RedisConfig{
		{RedisServerIP: "localhost", RedisConnType: "tcp", RedisServerPort: "6379", RedisServerPass: "shuffle123"},
		{RedisServerIP: "10.1.5.68", RedisConnType: "tcp", RedisServerPort: "6379", RedisServerPass: "123456"},
	}
	redisCli, err := NewClient(cfg)
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("new redis client Failed")
		return
	}
	redisCli.CheckRedisConPeriod()
}
