package redisWrapper

import (
	"errors"

	redis "github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// SubscriberCallback subscriber callback function
type SubscriberCallback func(channel, message string)

// Subscrible subscribe
func Subscrible(ctx context.Context, cli interface{}, clusterFlag bool, cb SubscriberCallback, channelName ...string) error {
	if cli == nil {
		log.Error("cli is nil")
		return errors.New("cli is nil")
	}
	var pubSub *redis.PubSub
	if clusterFlag {
		pubSub = cli.(*redis.ClusterClient).Subscribe(ctx, channelName...)
	} else {
		pubSub = cli.(*redis.Client).Subscribe(ctx, channelName...)
	}

	if pubSub == nil {
		log.Error("subscribler is nil")
		return errors.New("subscribler is nil")
	}

	// 订阅后获取消息
	go func() {
		ch := pubSub.Channel()
		for msg := range ch {
			cb(msg.Channel, msg.Payload)
		}
	}()

	return nil
}

// Publish publish to redis
func Publish(ctx context.Context, cli interface{}, clusterFlag bool, channelName, message string) error {
	if cli == nil {
		log.Error("cli is nil")
		return errors.New("cli is nil")
	}
	if clusterFlag {
		if _, err := cli.(*redis.ClusterClient).Publish(ctx, channelName, message).Result(); err != nil {
			log.Error("[cluster]redis publish failed: ", err)
			return err
		}
	} else {
		if _, err := cli.(*redis.Client).Publish(ctx, channelName, message).Result(); err != nil {
			log.Error("redis publish failed: ", err)
			return err
		}
	}

	return nil
}
