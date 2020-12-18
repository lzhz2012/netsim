package kafkaCliWrapper

import (
	"fmt"
	"os"
	"os/signal"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

// 支持brokers cluster的消费者
func ClusterConsumer(wg *sync.WaitGroup, brokers, topics []string, groupID string) {
	defer wg.Done()
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// init consumer
	consumer, err := cluster.NewConsumer(brokers, groupID, topics, config)
	if err != nil {
		log.Printf("%s: sarama.NewSyncProducer err, message=%s \n", groupID, err)
		return
	}
	defer consumer.Close()

	// trap SIGINT to trigger a shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// consume errors
	go func() {
		for err := range consumer.Errors() {
			log.Printf("%s:Error: %s\n", groupID, err.Error())
		}
	}()

	// consume notifications
	go func() {
		for ntf := range consumer.Notifications() {
			log.Printf("%s:Rebalanced: %+v \n", groupID, ntf)
		}
	}()

	// consume messages, watch signals
	var successes int
Loop:
	for {
		select {
		case msg, ok := <-consumer.Messages():
			if ok {
				fmt.Fprintf(os.Stdout, "%s:%s/%d/%d\t%s\t%s\n", groupID, msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
				consumer.MarkOffset(msg, "") // mark message as processed
				successes++
			}
		case <-signals:
			break Loop
		}
	}
	fmt.Fprintf(os.Stdout, "%s consume %d messages \n", groupID, successes)
}

var (
	wg sync.WaitGroup
)

// 普通消费者
func Consumer(wg *sync.WaitGroup, topic string, address []string) {
	// 可以设置消费者的配置
	consumer, err := sarama.NewConsumer(address, nil)
	if err != nil {
		log.Error("Failed to start consumer: ", err)
		return
	}
	//设置分区
	partitionList, err := consumer.Partitions(topic)
	if err != nil {
		log.Error("Failed to get the list of partitions: ", err)
		return
	}
	log.Println(partitionList)
	//循环分区
	for partition := range partitionList {
		pc, err := consumer.ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			log.Printf("Failed to start consumer for partition %d: %s\n", partition, err)
			return
		}
		defer pc.AsyncClose()
		go func(pc sarama.PartitionConsumer) {
			wg.Add(1)
			for msg := range pc.Messages() {
				log.Printf("Partition:%d, Offset:%d, Key:%s, Value:%s", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
				log.Println()
			}
			wg.Done()
		}(pc)
	}
	//time.Sleep(time.Hour)
	wg.Wait()
	consumer.Close()
}
