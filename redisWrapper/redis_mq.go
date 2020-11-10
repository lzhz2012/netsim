package redisWrapper

import (
	"errors"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

func (cli *RedisClient) Push(mqName, message string) error {
	if cli == nil {
		return errors.New("redis client is nil")
	}

	_, err := cli.conn.Do("rpush", mqName, message)
	if err != nil {
		log.Error("produce error")
		return err
	}
	log.Debug("produce element", message)
	return nil
}

func (cli *RedisClient) Pop(mqName string) (string, error) {
	if cli == nil {
		return "", errors.New("redis client is nil")
	}

	ele, err := redis.String(cli.conn.Do("lpop", mqName))
	if err != nil {
		//log.Error("no msg in queue")
		return "", err
	}

	log.Debug("cosume element", ele)
	return ele, nil
}

// 阻塞式从队列首读取一个元素，支持多个队列设置优先级
func (cli *RedisClient) Bpop(mqName []string, timeout uint32) (string, error) {
	if cli == nil {
		return "", errors.New("redis client is nil")
	}
	//FIXME:redis支持的队列最大的个数(与一个orborus支持的workflowQueues相关)
	if len(mqName) == 0 {
		return "", errors.New("mq Name length is 0")
	}

	args := redis.Args{}
	for _, q := range mqName {
		args = args.Add(q)
	}
	args = args.Add(timeout)
	eles, err := cli.conn.Do("blpop", args...)
	if err != nil {
		//log.Error("no msg in queue")
		return "", err
	}

	data, err := redis.Strings(eles, err)
	var content string
	for idx, ele := range data {
		if idx%2 != 0 {
			content = ele
		}
	}

	log.Println("cosume element", content)
	return content, nil
}
