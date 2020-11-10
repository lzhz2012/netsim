package redisWrapper

import (
	"strconv"
	"time"
	"unsafe"

	redis "github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

type SubscriberCallback func(channel, message string)

type Subscriber struct {
	client redis.PubSubConn
	cbMap  map[string]SubscriberCallback
}

func (c *Subscriber) Close() error {
	err := c.client.Close()
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("redis close error.")
	}
	return err
}

func (c *Subscriber) Subscribe(channel interface{}, cb SubscriberCallback) error {
	err := c.client.Subscribe(channel)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("redis Subscribe error.")
		return err
	}

	c.cbMap[channel.(string)] = cb
	return nil
}

// FIXME pub和sub目前分别都做了tcp连接，这样消耗比较大，后续采用有需求连接方式（共同tcp连接）
func (c *Subscriber) Connect(connType string, serverAddr string) error {
	conn, err := redis.Dial(connType, serverAddr)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("redis dial failed.")
		return err
	}

	c.client = redis.PubSubConn{}
	c.client.Conn = conn
	c.cbMap = make(map[string]SubscriberCallback)

	go func() {
		for {
			logrus.Debug("wait...")

			switch res := c.client.Receive().(type) {
			case redis.Message:
				channel := (*string)(unsafe.Pointer(&res.Channel))
				message := (*string)(unsafe.Pointer(&res.Data))
				c.cbMap[*channel](*channel, *message)
			case redis.Subscription:
				logrus.WithFields(logrus.Fields{
					"channel": res.Channel,
					"Kind":    res.Kind,
					"count":   res.Count,
				}).Debug("subscription info")
			case error:
				logrus.WithFields(logrus.Fields{"err": err}).Error("error handle...")
				continue
			}
		}
	}()
	logrus.Trace("Connect finished!")
	return nil
}

func Publish(connType, serverAddr, channelName, message string) error {
	client, err := redis.Dial(connType, serverAddr)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("redis dial failed.")
		return err
	}
	defer client.Close()
	_, err = client.Do("Publish", channelName, message)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("redis Publish failed.")
	}
	return err
}

// below is the test
func TestCallback1(chann, msg string) {
	logrus.Debug("TestCallback1 channel : ", chann, " message : ", msg)
}

func TestCallback2(chann, msg string) {
	logrus.Debug("TestCallback2 channel : ", chann, " message : ", msg)
}

func TestCallback3(chann, msg string) {
	logrus.Debug("TestCallback3 channel : ", chann, " message : ", msg)
}

func TestPubSub1() {

	var (
		redisServerHost string = "redis"
		//redisServerHost string = "localhost"
		redisConnType   string = "tcp"
		redisServerPort string = "6379"
		redisServerURL  string = redisServerHost + ":" + redisServerPort
		//g_checkRedisConPeriod uint   = 60 //second
	)
	//设置log的打印级别
	logrus.SetLevel(logrus.TraceLevel) //设置debug以上的信息都显示
	logrus.Info("===========test PubSub1 start============")

	var sub Subscriber
	sub.Connect(redisConnType, redisServerURL)
	sub.Subscribe("test_chan1", TestCallback1)
	sub.Subscribe("test_chan2", TestCallback2)
	sub.Subscribe("test_chan3", TestCallback3)

	for {
		for i := 1; i <= 3; i++ {
			Publish(redisConnType, redisServerURL, "test_chan"+strconv.Itoa(i), "hello")
		}

		time.Sleep(2 * time.Second)
	}
}
