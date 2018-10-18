package redisdb

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

var Pool *redis.Pool

// 初始化RedisDb
func InitRedisDb(maxIdle, maxActive, idleTimeOut int, address, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: 5 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}
			/*if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}*/
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

// 设置
func set(key string, val string) bool {

	return false
}

// 得到
func get(key string) (string, bool) {

	return "", false
}

// 删除
func del(key string) bool {

	return false
}

//RPOPLPUSH 循环列表
// 设置RoomID
func AddRoomId(roomid int, json string) bool {

	return false
}

// 得到RoomID
func GetRoomId(roomid int) (string, bool) {

	return "", false
}
