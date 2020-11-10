package rabbitMQWrapper

import (
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
)

// below is the test!
func TestRabbitMQWrapper(t *testing.T) {
	// log module init
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetReportCaller(true) // 打印文件及行数、调用函数
	var (
		rabbitMQServerUsr  string   = "admin"
		rabbitMQServerPass string   = "admin"
		rabbitMQServerIP   string   = "localhost"
		rabbitMQServerPort string   = "5672"
		rabbitMQVhost      string   = "/"
		rabbitMQQueues     []string = []string{"testqueue"}
	)
	rabbitMQCli, err := NewClient(rabbitMQServerIP, rabbitMQServerPort, rabbitMQServerUsr, rabbitMQServerPass, rabbitMQVhost)
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("rabbitMQCli is nil")
		return
	}
	for i := 0; i < 10; i++ {
		rabbitMQCli.Push(rabbitMQQueues[0], "this is a test!"+strconv.Itoa(i))
	}
	length, err := rabbitMQCli.GetQueueLength(rabbitMQQueues[0])
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("get rabbitMQ queue length failed")
		return
	}
	logrus.WithFields(logrus.Fields{"length": length}).Info("get rabbitMQ queue length success")

	rabbitMQCli.Pop(rabbitMQQueues[0], true)
	rabbitMQCli.PopAll(rabbitMQQueues[0])
}
