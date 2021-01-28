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
	RedisCfg []RedisConfig
}

func NewClient(cfg []RedisConfig) (*RedisClient, error) {
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
	var conn redis.Conn
	for _, cfg := range cli.RedisCfg {
		connTemp, err := redis.Dial(cfg.RedisConnType, cfg.RedisServerIP+":"+cfg.RedisServerPort,
			redis.DialPassword(cfg.RedisServerPass))
		if err != nil {
			logrus.WithFields(logrus.Fields{"err": err}).Error("connect Redis failed")
		} else {
			conn = connTemp
		}
	}

	// conn, err := redis.Dial(redisConnType, cli.redisServerIP + cli.redisServerPort,
	// 	redis.DialPassword(g_redisServerPasswd), redis.DialConnectTimeout(time.Duration(10)*time.Second),
	// 	redis.DialWriteTimeout(time.Duration(1000)*time.Millisecond), redis.DialReadTimeout(time.Duration(1000)*time.Millisecond),
	// 	redis.DialDatabase(0)) // 设置连接，读写超时时间，并设置连接默认连接数据库0
	if conn == nil {
		logrus.Error("connect Redis failed")
		return errors.New("connect redis client failed")
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

func (cli *RedisClient) Lock(key, value string, lifeTime uint64) error {
	if key == "" {
		return errors.New("lock failed: key is nil")
	}
	var resp interface{}
	var err error
	if lifeTime == 0 {
		resp, err = cli.conn.Do("setnx", key, value)
	} else {
		resp, err = cli.conn.Do("set", key, value, "NX", "PX", lifeTime) // lifeTime unit: ms
	}
	data, _ := redis.String(resp, err)
	if err != nil || data != "OK" {
		logrus.WithFields(logrus.Fields{"err": err}).Error("lock failed:setNX failed")
	}

	return err
}

func (cli *RedisClient) UnLock(key string) error {
	if key == "" {
		return errors.New("lock failed: key is nil")
	}
	resp, err := cli.conn.Do("del", key)
	number, _ := redis.Uint64(resp, err)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("Unlock failed:del key failed")
		return err
	}
	if number != 1 {
		logrus.WithFields(logrus.Fields{"err": err}).Debug("key of the lock is not exist")
	}
	return nil
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

func (cli *RedisClient) Hmget(hashMapName string, keys []string) ([]string, error) {
	if cli == nil {
		return []string{}, errors.New("redis client is nil")
	}
	args := redis.Args{}
	args = args.Add(hashMapName)
	for _, q := range keys {
		args = args.Add(q)
	}
	eles, err := cli.conn.Do("hmget", args...)
	if err != nil {
		logrus.WithFields(logrus.Fields{"key": keys, "err": err}).Error("redis Hmget failed")
		return []string{}, err
	}
	data, err := redis.Strings(eles, err)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("redis convert to strings failed")
		return []string{}, err
	}
	return data, err
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

func (cli *RedisClient) Hdel(hashMapName, key string) error {
	if cli == nil {
		return errors.New("redis client is nil")
	}
	resp, err := cli.conn.Do("hdel", hashMapName, key)
	if err != nil {
		logrus.WithFields(logrus.Fields{"key": key, "err": err}).Error("redis Hget failed")
		return err
	}

	number, err := redis.Uint64(resp, err)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("redis Hdel failed")
		return err
	}
	if number <= 0 {
		logrus.WithFields(logrus.Fields{"err": err}).Error("redis Hdel failed")
		return errors.New("redi hdel return number is less than, del fail!")
	}
	return nil
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

func (cli *RedisClient) HKeys(hashMapName string) ([]string, error) {
	if cli == nil {
		return []string{}, errors.New("redis client is nil")
	}
	resp, err := cli.conn.Do("hkeys", hashMapName)
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

func (cli *RedisClient) Sadd(key, value string) error {
	if cli == nil {
		return errors.New("redis client is nil")
	}
	resp, err := cli.conn.Do("sadd", key, value)
	_ = resp
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("redis sadd failed")
		return err
	}

	return err
}

func (cli *RedisClient) Smembers(key string) ([]string, error) {
	if cli == nil {
		return []string{}, errors.New("redis client is nil")
	}
	resp, err := cli.conn.Do("smembers", key)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("redis smembers failed")
		return []string{}, err
	}

	data, err := redis.Strings(resp, err)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("redis convert to strings failed")
		return []string{}, err
	}
	logrus.Println(data)
	return data, nil
}

func (cli *RedisClient) Srem(key, value string) error {
	if cli == nil {
		return errors.New("redis client is nil")
	}
	resp, err := cli.conn.Do("srem", key, value)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("redis smembers failed")
		return err
	}

	data, _ := redis.Int(resp, err)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("key is not the element of set")
		return err
	}
	if data <= 0 {
		logrus.WithFields(logrus.Fields{"err": err}).Error("Can't find the element in set")
		return errors.New("Can't find the element in set")
	}
	return nil
}
