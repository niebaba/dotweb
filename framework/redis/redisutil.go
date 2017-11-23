// redisclient

// redis 命令的使用方式参考
// http://doc.redisfans.com/index.html
package redisutil

import (
	"github.com/garyburd/redigo/redis"
	"sync"
)

type RedisClient struct {
	pool    *redis.Pool
	Address string
}

var (
	redisMap map[string]*RedisClient
	mapMutex *sync.RWMutex
)

const (
	defaultTimeout = 60 * 10 //默认10分钟
)

func init() {
	redisMap = make(map[string]*RedisClient)
	mapMutex = new(sync.RWMutex)
}

// 重写生成连接池方法
// redisURL: connection string, like "redis://:password@10.0.1.11:6379/0"
func newPool(redisURL string) *redis.Pool {

	return &redis.Pool{
		MaxIdle:   5,
		MaxActive: 20, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(redisURL)
			return c, err
		},
	}
}

//GetRedisClient 获取指定Address的RedisClient
func GetRedisClient(address string) *RedisClient {
	var redis *RedisClient
	var mok bool
	mapMutex.RLock()
	redis, mok = redisMap[address]
	mapMutex.RUnlock()
	if !mok {
		redis = &RedisClient{Address: address, pool: newPool(address)}
		mapMutex.Lock()
		redisMap[address] = redis
		mapMutex.Unlock()
	}
	return redis
}

//GetObj 获取指定key的内容, interface{}
func (rc *RedisClient) GetObj(key string) (interface{}, error) {
	// 从连接池里面获得一个连接
	conn := rc.pool.Get()
	// 连接完关闭，其实没有关闭，是放回池里，也就是队列里面，等待下一个重用
	defer conn.Close()
	reply, errDo := conn.Do("GET", key)
	return reply, errDo
}

//Get 获取指定key的内容, string
func (rc *RedisClient) Get(key string) (string, error) {
	val, err := redis.String(rc.GetObj(key))
	return val, err
}

//Exists 检查指定key是否存在
func (rc *RedisClient) Exists(key string) (bool, error) {
	// 从连接池里面获得一个连接
	conn := rc.pool.Get()
	// 连接完关闭，其实没有关闭，是放回池里，也就是队列里面，等待下一个重用
	defer conn.Close()

	reply, errDo := redis.Bool(conn.Do("EXISTS", key))
	return reply, errDo
}

//Del 删除指定key
func (rc *RedisClient) Del(key string) (int64, error) {
	// 从连接池里面获得一个连接
	conn := rc.pool.Get()
	// 连接完关闭，其实没有关闭，是放回池里，也就是队列里面，等待下一个重用
	defer conn.Close()
	reply, errDo := conn.Do("DEL", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int64(reply, errDo)
	return val, err
}

//INCR 对存储在指定key的数值执行原子的加1操作
func (rc *RedisClient) INCR(key string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("INCR", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int(reply, errDo)
	return val, err
}

//DECR 对存储在指定key的数值执行原子的减1操作
func (rc *RedisClient) DECR(key string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("DECR", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int(reply, errDo)
	return val, err
}


//Append 如果 key 已经存在并且是一个字符串， APPEND 命令将 value 追加到 key 原来的值的末尾。
// 如果 key 不存在， APPEND 就简单地将给定 key 设为 value ，就像执行 SET key value 一样。
func (rc *RedisClient) Append(key string, val interface{}) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("APPEND", key, val)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Uint64(reply, errDo)
	return val, err
}

//Set 设置指定Key/Value
func (rc *RedisClient) Set(key string, val interface{}) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("SET", key, val))
	return val, err
}

//SetWithExpire 设置指定key的内容
func (rc *RedisClient) SetWithExpire(key string, val interface{}, timeOutSeconds int64) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("SET", key, val, "EX", timeOutSeconds))
	return val, err
}

//SetNX  将 key 的值设为 value ，当且仅当 key 不存在。
// 若给定的 key 已经存在，则 SETNX 不做任何动作。 成功返回1, 失败返回0
func (rc *RedisClient) SetNX(key, value string) (interface{}, error){
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := conn.Do("SETNX", key, value)
	return val, err
}


//HGet 获取指定hashset的内容
func (rc *RedisClient) HGet(hashID string, field string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("HGET", hashID, field)
	if errDo == nil && reply == nil {
		return "", nil
	}
	val, err := redis.String(reply, errDo)
	return val, err
}

//HGetAll 获取指定hashset的所有内容
func (rc *RedisClient) HGetAll(hashID string) (map[string]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, err := redis.StringMap(conn.Do("HGetAll", hashID))
	return reply, err
}

//HSet 设置指定hashset的内容
func (rc *RedisClient) HSet(hashID string, field string, val string) error {
	conn := rc.pool.Get()
	defer conn.Close()
	_, err := conn.Do("HSET", hashID, field, val)
	return err
}

//HSetNX 设置指定hashset的内容, 如果field不存在, 该操作无效
func (rc *RedisClient) HSetNX(key, field, value string) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := conn.Do("HSETNX", key, field, value)
	return val, err
}

//HLen 返回哈希表 key 中域的数量, 当 key 不存在时，返回0
func (rc *RedisClient) HLen(key string) (int64, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("HLEN", key))
	return val, err
}

//HDel 设置指定hashset的内容, 如果field不存在, 该操作无效, 返回0
func (rc *RedisClient) HDel(args ...interface{}) (int64, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("HDEL", args...))
	return val, err
}

//HVals 返回哈希表 key 中所有域的值, 当 key 不存在时，返回空
func (rc *RedisClient) HVals(key string) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.Strings(conn.Do("HVALS", key))
	return val, err
}



//BRPop 删除，并获得该列表中的最后一个元素，或阻塞，直到有一个可用
func (rc *RedisClient) BRPop(key string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.StringMap(conn.Do("BRPOP", key, defaultTimeout))
	if err != nil {
		return "", err
	} else {
		return val[key], nil
	}
}

//LPush 将所有指定的值插入到存于 key 的列表的头部
func (rc *RedisClient) LPush(key string, val string) (int64, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	ret, err := redis.Int64(conn.Do("LPUSH", key, val))
	if err != nil {
		return -1, err
	} else {
		return ret, nil
	}
}


//Expire 设置指定key的过期时间
func (rc *RedisClient) Expire(key string, timeOutSeconds int64) (int64, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int64(conn.Do("EXPIRE", key, timeOutSeconds))
	return val, err
}

//FlushDB 删除当前数据库里面的所有数据
//这个命令永远不会出现失败
func (rc *RedisClient) FlushDB() {
	conn := rc.pool.Get()
	defer conn.Close()
	conn.Do("FLUSHALL")
}


//返回一个从连接池获取的redis连接,  需要手动释放redis连接
func (rc *RedisClient) ConnGet() redis.Conn{
	conn := rc.pool.Get()

	return conn
}



func (rc *RedisClient) SAdd(args ...interface{}) (int64, error){
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("SADD", args...))
	return val, err
}

func (rc *RedisClient) SCard(key string) (int64, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("SCARD", key))
	return val, err
}

func (rc *RedisClient) SPop(key string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.String(conn.Do("SPOP", key))
	return val, err
}

func (rc *RedisClient) SRandMember(args ...interface{}) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.String(conn.Do("SRANDMEMBER", args...))
	return val, err
}

func (rc *RedisClient) SRem(args ...interface{}) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.String(conn.Do("SREM", args...))
	return val, err
}

func (rc *RedisClient) DBSize()(int64, error){
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("DBSIZE"))
	return val, err
}