package roombase

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

var (
	RedisPool *redis.Pool
	RedisConn redis.Conn
)

//"{\"unique_code\":\"595e26f39f6f79xlim8b\",\"server_room_id\":621272,\"game_type\":3000,\"rules\":[\"25\",\"19\",\"23\"],\"user_id\":\"3054\"}"

// Redis中保存的Json数据结构
type RedisRoomInfo struct {
	Unique_code    string
	Server_room_id int
	Game_type      int
	User_id        string
	Rules          []int32
	//RuleInt      []int32
}

// 初始化RedisDb
func Redis_InitRedisDb(address, password string, maxIdle, maxActive, idleTimeOut int) (*redis.Pool, error) {
	//	var err error
	//	RedisConn, err = redis.Dial("tcp", address)
	//	if err != nil {
	//		fmt.Println(err)
	//		return nil, err
	//	}
	//	defer RedisConn.Close()

	//RedisConn = &c

	RedisPool = &redis.Pool{
		MaxIdle:     0,
		MaxActive:   1000,
		IdleTimeout: 0,
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
		//		TestOnBorrow: func(c redis.Conn, t time.Time) redis.Conn, error {
		//			if time.Since(t) < time.Minute {
		//				return c, nil
		//			}
		//			c, err := c.Do("PING")
		//			return c, err
		//		},
	}
	RedisConn, _ = RedisPool.Dial()
	if RedisPool != nil && RedisConn != nil {
		fmt.Println("Connect RedisDB [" + address + "] Successed")
	} else {
		fmt.Println("Connect RedisDB [" + address + "] Error ")
	}
	return RedisPool, nil
}

//@andy
func GetRedisConn() (redis.Conn, error) {
	return RedisPool.Dial()
}

// 清空Redis RoomKey
func Redis_ClearRedis() {
	conn, _ := GetRedisConn()
	//v, err := RedisConn.Do("keys", "room:*")
	v, err := conn.Do("keys", "room:*")
	if err != nil {
		fmt.Println("Error ", err)
		return
	} else {
		oldrooms, _ := redis.MultiBulk(v, nil)
		names := make([]interface{}, len(oldrooms))
		for i := 0; i < len(oldrooms); i++ {
			n, _ := redis.String(oldrooms[i], nil)
			names[i] = n
		}
		if len(names) > 0 {
			//v, err := RedisConn.Do("del", names...)
			v, err := conn.Do("del", names...)
			if err == nil {
				fmt.Println("Clear All Redis ", v)
			} else {
				fmt.Println("Clear Redis Error ", err)
			}
		}
	}

	//_, errr := RedisConn.Do("del", "global:playingUser")
	_, errr := conn.Do("del", "global:playingUser")
	if errr != nil {
		fmt.Println("Error ", err)
	}

}

// 检查Room是否存在
func Redis_CheckRoom(roomid int, unique_code string) (int, *RedisRoomInfo) {
	return 0, nil
	var v interface{}
	var err error

	if roomid > 0 {
		str := "room:server_room_id:" + strconv.Itoa(roomid)
		conn, _ := GetRedisConn()
		//v, err = RedisConn.Do("GET", str)
		v, err = conn.Do("GET", str)

		if err != nil {
			fmt.Println("Error redis  ", roomid, err)
			return 0, nil
		}
	}

	s, _ := redis.String(v, nil)

	b := []byte(s)
	roominfo := RedisRoomInfo{}
	json.Unmarshal(b, &roominfo)

	if roominfo.Server_room_id > 0 {
		return roominfo.Server_room_id, &roominfo
	} else {
		return 0, nil
	}
}

// 添加房间内的Playing用户
func Redis_AddPlayingUser(roomid int, uid string) bool {
	return false
	conn, _ := GetRedisConn()

	_, err := conn.Do("hset", strconv.Itoa(roomid), uid, "1")

	if err != nil {
		fmt.Println("Redis_AddPlayingUser Error ", err)
		return false
	}

	_, err = conn.Do("set", uid, strconv.Itoa(roomid))
	if err != nil {
		fmt.Println("Redis_AddPlayingUser Error ", err)
		return false
	}

	return true
}

// 删除房间内所有用户
func Redis_RemovePlayingUser(roomid int) bool {
	return false
	conn, _ := GetRedisConn()
	values, err := redis.Values(conn.Do("hkeys", strconv.Itoa(roomid)))

	if err != nil {
		fmt.Println("Redis_RemoveRoomUsers Error ", err)
		return false
	}

	for _, v := range values {
		uid := string(v.([]byte))
		fmt.Println("shanchuwanjia.." + uid)
		_, err = conn.Do("del", uid)
		if err != nil {
			fmt.Println("Redis_RemoveRoomUsers Error ", err)
			return false
		}
	}

	return true
}

// 删除房间内的用户
func Redis_RemoveUser(roomid string, uid string) bool {
	return false
	conn, _ := GetRedisConn()

	_, err := conn.Do("del", uid)
	if err != nil {
		fmt.Println("Redis_RemoveUser Error ", err)
		return false
	}
	_, err = conn.Do("hdel", roomid, uid)
	if err != nil {
		fmt.Println("Redis_RemoveUser Error ", err)
		return false
	}
	return true
}

// 删除房间
func Redis_RemoveRoom(roomid int) bool {
	return false
	conn, _ := GetRedisConn()
	_, err := conn.Do("del", strconv.Itoa(roomid))
	if err != nil {
		fmt.Println("Redis_RemoveRoom Error ", err)
		return false
	}
	return true
}

//RPOPLPUSH 循环列表
// 设置RoomID
func Redis_AddRoomId(roomid int, unique_code string, j string) bool {
	return false
	//	conn, _ := GetRedisConn()
	//	//_, err := RedisConn.Do("EXISTS", "room:unique_code:"+unique_code)
	//	_, err := conn.Do("EXISTS", "room:unique_code:"+unique_code)
	//	if err != nil {
	//		fmt.Println("Error ", err)
	//		return false
	//	}

	return true
}

// 得到RoomID
func Redis_GetRoomForId(unique_code string) (string, bool) {
	//	conn, _ := GetRedisConn()
	//	//v, err := RedisConn.Do("GET", "room:unique_code:"+unique_code)
	//	v, err := conn.Do("GET", "room:unique_code:"+unique_code)
	//	if err != nil {
	//		fmt.Println("Error ", err)
	//		return "", false
	//	} else {
	//		fmt.Println(v)
	//	}
	return "", false
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
