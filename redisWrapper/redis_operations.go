package redisWrapper

import (
	"errors"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

type RedisConfig struct {
	RedisServerIP   string `json:"redisServerIP"`
	RedisServerPort string `json:"redisServerPort"`
	RedisConnType   string `json:"redisConnType"`
	RedisServerPass string `json:"redisServerPass"`
}

type RedisClient struct {
	conn     redis.Conn
	RedisCfg RedisConfig
}

func NewClient(cfg RedisConfig) (*RedisClient, error) {
	redisCli := &RedisClient{RedisCfg: cfg}
	err := redisCli.Connect()
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("connect redis server failed!")
		return nil, err
	}
	return redisCli, err
}

func (cli *RedisClient) Connect() error {
	if cli == nil {
		return errors.New("redis client is nil")
	}
	conn, err := redis.Dial(cli.RedisCfg.RedisConnType, cli.RedisCfg.RedisServerIP+":"+cli.RedisCfg.RedisServerPort,
		redis.DialPassword("shuffle123"))

	// conn, err := redis.Dial(redisConnType, cli.redisServerIP + cli.redisServerPort,
	// 	redis.DialPassword(g_redisServerPasswd), redis.DialConnectTimeout(time.Duration(10)*time.Second),
	// 	redis.DialWriteTimeout(time.Duration(1000)*time.Millisecond), redis.DialReadTimeout(time.Duration(1000)*time.Millisecond),
	// 	redis.DialDatabase(0)) // 设置连接，读写超时时间，并设置连接默认连接数据库0
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("connect Redis failed")
		return err
	}
	cli.conn = conn
	return nil
}

func (cli *RedisClient) CheckRedisConPeriod() {
	for {
		if !cli.isRedisConnect() {
			logrus.Debug("try to connect redis server")
			if err := cli.Connect(); err != nil {
				logrus.WithFields(logrus.Fields{"err": err}).Error("reconnect redis server failed")
			}
		}
		ticker := time.NewTicker(time.Second * 60)
		logrus.Trace("redis period check")
		<-ticker.C
	}
}

func (cli *RedisClient) isRedisConnect() bool {
	if cli.conn == nil {
		logrus.Error(errors.New("redis client connect is nil"))
		return false
	}
	_, err := cli.conn.Do("PING")
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("PING Redis server failed!")
		return false
	}
	return true
}

func (cli *RedisClient) Close() error {
	if cli == nil {
		return errors.New("redis client is nil")
	}
	err := cli.conn.Close()
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("close Redis failed")
	}
	return err
}

func (cli *RedisClient) Hset(hashMapName, key, value string) error {
	if cli == nil {
		return errors.New("redis client is nil")
	}
	_, err := cli.conn.Do("hset", hashMapName, key, value)
	if err != nil {
		logrus.WithFields(logrus.Fields{"key": key, "err": err}).Debug("redis Hset failed")
	}
	return err
}

func (cli *RedisClient) Hget(hashMapName, key string) (string, error) {
	if cli == nil {
		return "", errors.New("redis client is nil")
	}
	resp, err := cli.conn.Do("hget", hashMapName, key)
	if err != nil {
		logrus.WithFields(logrus.Fields{"key": key, "err": err}).Error("redis Hget failed")
		return "", err
	}

	data, err := redis.String(resp, err)
	if err != nil {
		//logrus.WithFields(logrus.Fields{"err": err}).Error("redis Hget failed")
	}
	return data, err
}

func (cli *RedisClient) GetKeys(pattern string) ([]string, error) {
	if cli == nil {
		return []string{}, errors.New("redis client is nil")
	}
	resp, err := cli.conn.Do("keys", pattern)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("redis Get keys failed")
		return []string{}, err
	}

	data, err := redis.Strings(resp, err)
	if err != nil {
		//logrus.WithFields(logrus.Fields{"err": err}).Error("redis Hget failed")
	}

	return data, err
}
