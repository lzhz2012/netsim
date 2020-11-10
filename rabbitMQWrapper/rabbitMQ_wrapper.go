package rabbitMQWrapper

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type RabbitMQClient struct {
	conn               *amqp.Connection
	rabbitMQServerUsr  string
	rabbitMQServerPass string
	rabbitMQServerIP   string
	rabbitMQServerPort string
	rabbitMQVhost      string
}

func NewClient(rabbitMQServerIP, rabbitMQServerPort, rabbitMQServerUsr, rabbitMQServerPass, rabbitMQVhost string) (*RabbitMQClient, error) {
	rabbitMQCli := &RabbitMQClient{rabbitMQServerUsr: rabbitMQServerUsr, rabbitMQServerIP: rabbitMQServerIP,
		rabbitMQServerPort: rabbitMQServerPort, rabbitMQServerPass: rabbitMQServerPass}
	mqurl := "amqp://" + rabbitMQServerUsr + ":" + rabbitMQServerPass + "@" + rabbitMQServerIP + ":" + rabbitMQServerPort + rabbitMQVhost
	err := rabbitMQCli.Connect(mqurl)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("connect rabbitMQ server failed!")
		return nil, err
	}
	return rabbitMQCli, err
}

func (cli *RabbitMQClient) Connect(mqurl string) error {

	//"amqp://admin:admin@10.10.15.147:5672/my_vhost"

	conn, err := amqp.Dial(mqurl)

	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("failed to connect tp rabbitmq")
	}
	cli.conn = conn
	return err
}

func (cli *RabbitMQClient) Close() error {
	err := cli.conn.Close()
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("failed to close connect tp rabbitmq")
	}
	return err
}

func (cli *RabbitMQClient) GetQueueLength(queueName string) (int, error) {
	if cli == nil || cli.conn == nil {
		return 0, errors.New("rabbitMQ client or connect is nil")
	}
	channel, err := cli.conn.Channel()
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("failed to open a MQ channel")
		return 0, err
	}
	defer channel.Close()

	//检查需要消费的队列是否存在
	q, err := channel.QueueDeclare(
		queueName, // 队列名称
		true,      // 是否持久化
		false,     // 是否自动删除
		false,     // 是否独立
		false,     // 是否自动应答
		nil,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{"QueueName": q.Name, "err": err}).Error("declare MQ queue failed")
		return 0, err
	}
	return q.Messages, nil
}

func (cli *RabbitMQClient) Push(queueName, msgContent string) error {

	if cli == nil || cli.conn == nil {
		return errors.New("rabbitMQ client or connect is nil")
	}

	//setup channel for producer
	channel, err := cli.conn.Channel()
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("failed to open a channel")
		return err
	}
	defer channel.Close()

	// publish one msg
	logrus.Printf("put msg in rabbit queue msgContent :%s", msgContent)
	err = channel.Publish("", queueName, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msgContent),
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("Failed to publish a MQ message")
	}
	return nil
}

func (cli *RabbitMQClient) Pop(queueName string, workQueue bool) (string, error) {

	if cli == nil || cli.conn == nil {
		return "", errors.New("rabbitMQ client or connect is nil")
	}

	//setup channel for consumer
	channel, err := cli.conn.Channel()
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("failed to open a MQ channel")
		return "", err
	}
	defer channel.Close()

	//检查需要消费的队列是否存在
	q, err := channel.QueueDeclare(
		queueName, // 队列名称
		true,      // 是否持久化
		false,     // 是否自动删除
		false,     // 是否独立
		false,     // 是否自动应答
		nil,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{"QueueName": q.Name, "err": err}).Error("declare MQ queue failed")
		return "", err
	}
	if workQueue {
		err := channel.Qos(
			1,     // prefetch count
			0,     // prefetch size
			false, // global
		)
		if err != nil {
			logrus.Printf("set workqueue failed,err:%s", err)
		}
	}

	// consume one msg
	msg, ok, err := channel.Get(queueName, false) //get one msg once
	if err != nil {
		logrus.WithFields(logrus.Fields{"QueueName": q.Name, "err": err}).Error("can't pop a msg in MQ queue")
		return "", err
	}

	if !ok {
		//logrus.WithFields(logrus.Fields{"QueueName": q.Name}).Debug("Queue is nil")
		return "", err
	}
	msg.Ack(false)
	logrus.WithFields(logrus.Fields{"data": string(msg.Body)}).Debug("get msg in MQ queue msgContent")
	return string(msg.Body), nil
}

func (cli *RabbitMQClient) PopAll(queueName string) (string, error) {

	// check the conn is open, if it's closed, connect again
	if cli == nil || cli.conn == nil {
		return "", errors.New("rabbitMQ client or connect is nil")
	}

	//setup channel for sub
	channel, err := cli.conn.Channel()
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("failed to open a channel")
		return "", err
	}
	defer channel.Close()

	//检查需要消费的队列是否存在,不存在就声明一个队列
	q, err := channel.QueueDeclare(
		queueName, // 队列名称
		true,      // 是否持久化
		false,     // 是否自动删除
		false,     // 是否独立
		false,
		nil,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{"QueueName": q.Name, "err": err}).Error("declare MQ queue failed")
		return "", err
	}

	// consume multiple msgs
	msgs, err := channel.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("Failed to register a consumer:")
		return "", err
	}

	data := []byte{}
	for msg := range msgs {
		data = append(data, msg.Body...)
		msg.Ack(false)
	}

	logrus.WithFields(logrus.Fields{"data": string(data)}).Debug("get all msgs in MQ queue msgContent")
	return string(data), nil
}
