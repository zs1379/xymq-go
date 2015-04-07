package operation

import (
	redis "github.com/garyburd/redigo/redis"
	"time"
)

var (
	redisList map[string]redis.Conn
)

func getRedis(redisName string) redis.Conn {

	redisList = make(map[string]redis.Conn)

	if redisList[redisName] == nil {
		//todo:将配置抽离成配置文件,根据不同的redis配置选择不同的redis实例
		c, err := redis.DialTimeout("tcp", "192.168.10.5:8003", 3*time.Second, 1*time.Second, 1*time.Second)

		if err != nil {
			panic("redis连接异常！")
		}

		redisList[redisName] = c
	}

	return redisList[redisName]
}
