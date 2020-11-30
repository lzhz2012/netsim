package redisWrapper

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/sirupsen/logrus"
)

var clusterClient *redis.ClusterClient

func init() {
	log.SetFlags(log.Llongfile | log.Lshortfile)
	// 连接redis集群
	clusterClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{ // 填写master主机
			"10.1.5.68:6379",
			"10.1.5.69:6379",
			"10.1.5.70:6379",
		},
		// Addrs: []string{ // 填写master主机
		// 	"10.1.4.68:6379",
		// 	"10.1.4.69:6379",
		// 	"10.1.4.70:6379",
		// },
		Password:     "123456",              // 设置密码
		DialTimeout:  60 * time.Millisecond, // 设置连接超时
		ReadTimeout:  60 * time.Millisecond, // 设置读取超时
		WriteTimeout: 60 * time.Millisecond, // 设置写入超时
	})
	// 发送一个ping命令,测试是否通
	s := clusterClient.Do("ping").String()
	fmt.Println(s)
}

func TestConnByRedisCluster(t *testing.T) {
	// 测试一个set功能
	s := clusterClient.Set("name", "barry", time.Second*60).String()
	fmt.Println(s)
	s1 := clusterClient.Get("name")
	// workflowQueueNames := []string{"{name}1111", "{name}1"}
	workflowQueueNames := []string{"{name}1111", "{name}1"}
	result := clusterClient.LPush(workflowQueueNames[0], "this is a test").String()
	clusterClient.LPush(workflowQueueNames[1], "this is a test1")
	fmt.Println(result)
	slice1, err := clusterClient.LPop("name1111").Result()
	fmt.Println(slice1, err)
	//queues := []string{"{name}1111", "{name}1"}
	//slice, err := clusterClient.BLPop(time.Second*0, queues...).Result()
	slice, err := clusterClient.BLPop(time.Second*0, "{name}111", "{name}1").Result()
	// slice, err := clusterClient.BLPop(time.Second*0, "name1111", "name1").Result()
	var data string
	if err == nil {
		data = slice[1]
	}
	fmt.Println(data, err)

	fmt.Println(s1)
}

func TestConnByRedisClusterConsumer(t *testing.T) {
	queues := []string{"{workflowqueue}Shuffle", "{workflowqueue}Shuffle1111111"}
	slice, err := clusterClient.BLPop(time.Second*0, queues...).Result()
	// slice, err := clusterClient.BLPop(time.Second*0, "name1111", "name1").Result()
	var data string
	if err == nil {
		data = slice[1]
	}
	logrus.Info(data)
}

func TestConnByRedisClusterProducer(t *testing.T) {

	var executionRequestWrapper ExecutionRequestWrapper
	executionRequest := ExecutionRequest{
		ExecutionId:   "qqskfskafjlsfslffsljfs",
		WorkflowId:    "12324566545455556",
		Authorization: "",
		Environments:  []string{"shuffle", "shuffle1"},
	}
	executionRequestWrapper.Data = append(executionRequestWrapper.Data, executionRequest)
	dataByte, err := json.Marshal(&executionRequestWrapper)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("Failed executionrequest in queue marshaling")
		return
	}
	mqName := "workflowqueueShuffle1111111"
	for {
		result := clusterClient.LPush(mqName, string(dataByte)).String()
		logrus.Info(result)

		time.Sleep(2 * time.Second)
	}
}
