package redisWrapper

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	redis "github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// type SubscriberCallback func(channel, message string)

var gClusterClientTest *redis.ClusterClient

func initRedisCluster() (*redis.ClusterClient, error) {
	addrs := []string{"10.1.5.68:6379", "10.1.5.69:6379", "10.1.5.70:6379"}
	//var addrs []string
	if os.Getenv("REDIS_HOST") != "" {
		addrs = strings.Split(os.Getenv("REDIS_HOST"), " ")
		for idx, add := range addrs {
			addrs[idx] = add + ":" + strings.Split(os.Getenv("REDIS_PORT"), " ")[idx]
		}
	}
	passWd := strings.Split(os.Getenv("REDIS_PASSWD"), " ")[0]
	if len(passWd) == 0 {
		passWd = "123456"
	}

	log.Info("redis addrs and passwd", addrs, passWd)

	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        addrs,                   // 填写master主机
		Password:     passWd,                  // 设置密码
		DialTimeout:  1000 * time.Millisecond, // 设置连接超时
		ReadTimeout:  600 * time.Millisecond,  // 设置读取超时
		WriteTimeout: 600 * time.Millisecond,  // 设置写入超时
	})
	// 发送一个ping命令,测试是否通
	s := clusterClient.Do(context.Background(), "ping").String()
	if strings.Contains(s, "cluster has no nodes") {
		log.Error("init redis cluster client failed")
		return clusterClient, errors.New(s)
	}
	if !strings.Contains(s, "PONG") {
		log.Debug("init redis cluster client failed")
		return clusterClient, errors.New("cluster client ping failed")
	}
	return clusterClient, nil
}

func testCb1(channel, message string) {
	log.Debug("Enter the testCB1")
	log.Debug(channel, message)
}

func TestPubSub(t *testing.T) {

	// var (
	// 	redisServerHost string = "redis"
	// 	//redisServerHost string = "localhost"
	// 	redisConnType   string = "tcp"
	// 	redisServerPort string = "6379"
	// 	redisServerURL  string = redisServerHost + ":" + redisServerPort
	// 	//g_checkRedisConPeriod uint   = 60 //second
	// )
	gClusterClientTest, _ = initRedisCluster()
	//设置log的打印级别
	logrus.SetLevel(logrus.TraceLevel) //设置debug以上的信息都显示
	logrus.Info("===========test PubSub start============")

	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		Subscrible(ctx, gClusterClientTest, true, testCb1, "test_chan"+strconv.Itoa(i))
	}

	for {
		for i := 1; i <= 3; i++ {
			Publish(ctx, gClusterClientTest, true, "test_chan"+strconv.Itoa(i), "hello")
		}

		time.Sleep(2 * time.Second)
	}
}
