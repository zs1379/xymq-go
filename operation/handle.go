package operation

import (
	"crypto"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type returnStruct struct {
	Ret     int    `json:"ret"`
	Msg     string `json:"msg"`
	MsgId   string `json:"msg_id"`
	ReqId   string `json:"req_id"`
	MsgBody string `json:"msg_body"`
}

var (
	queueName   string
	msgBody     string
	returnValue returnStruct
)

/**
 * 方法处理，用于某种方法的调用
 */
func OperationHandle(w http.ResponseWriter, method string, param string) string {
	returnValue.ReqId = getReqId()
	returnValue.Ret = 0

	//解析json
	var jsonData interface{}
	err := json.Unmarshal([]byte(param), &jsonData)
	if err != nil {
		panic(fmt.Sprintf("json解析失败哟,err:%s", err))
	}

	//转化参数
	var params []string

	for _, v := range jsonData.([]interface{}) {
		params = append(params, v.(string))
	}

	//根据方法名称选择处理方法
	switch method {
	case "push":
		if len(params) == 2 {
			result, err := push(params[0], params[1])

			if err != nil {
				returnValue.Msg = err.Error()
			}

			returnValue.MsgId = result
		} else {
			panic(fmt.Sprintf("method:%s,参数长度异常%s", method, params))
		}
	case "prePush":
		if len(params) == 2 {
			result, err := prePush(params[0], params[1])

			if err != nil {
				returnValue.Msg = err.Error()
			}

			returnValue.MsgId = result
		} else {
			panic(fmt.Sprintf("method:%s,参数长度异常%s", method, params))
		}
	default:
		panic(fmt.Sprintf("method:%s,无有效方法%s", method, params))
	}

	returnJsonValue := getReturnValue()

	return returnJsonValue
}

/**
 * 构造返回值
 */
func getReturnValue() string {
	jsonResult, err := json.Marshal(returnValue)

	if err != nil {
		panic(fmt.Sprintf("生成json数据异常:%s", err))
	}

	return string(jsonResult)
}

/**
 * 推入队列
 */
func push(pQueueName string, pMessageBody string) (string, error) {
	queueName = pQueueName
	msgBody = pMessageBody

	err := check()

	if err != nil {
		return "", err
	}

	msgId, err := pushBack(false)

	if err != nil {
		return "", err
	}

	return msgId, nil
}

/**
 * 推入优先队列
 */
func prePush(pQueueName string, pMessageBody string) (string, error) {
	queueName = pQueueName
	msgBody = pMessageBody

	err := check()

	if err != nil {
		return "", err
	}

	msgId, err := pushBack(true)

	if err != nil {
		return "", err
	}

	return msgId, nil
}

/**
 * 将消息插入队列
 */
func pushBack(pIsPre bool) (string, error) {
	msgId := getMsgId()

	//获取redis
	redis := getRedis(queueName)

	//设置对应关系
	_, err := redis.Do("hSet", getKey("msgIdHash"), msgId, msgBody)

	if err != nil {
		returnValue.Ret = 9999
		return "", err
	}

	//插入活跃消息队列
	if pIsPre {
		_, err = redis.Do("lPush", getKey("msgIdPreList"), msgId)
	} else {
		_, err = redis.Do("lPush", getKey("msgIdList"), msgId)
	}

	if err != nil {
		returnValue.Ret = 9999
		return "", err
	} else {
		return msgId, nil
	}
}

func pop(pQueueName string, pVisibilityTimeOut int) (string, error) {
	queueName = pQueueName
}

/**
 * 获取redis的key值
 */
func getKey(pTypeName string) string {
	redisKey := ""

	switch pTypeName {
	case "msgIdHash":
		redisKey = fmt.Sprintf("%s%s", queueName, "_message_id_hash")
	case "msgIdList":
		redisKey = fmt.Sprintf("%s%s", queueName, "_message_list")
	case "msgIdPreList":
		redisKey = fmt.Sprintf("%s%s", queueName, "_message_pre_list")
	case "unActivitySet":
		redisKey = fmt.Sprintf("%s%s", queueName, "_message_no_activity_set")
	case "errorHash":
		redisKey = fmt.Sprintf("%s%s", queueName, "_message_error_num_hash")
	}

	return redisKey
}

/**
 * 检测请求体是否有异常
 */
func check() error {
	//获取redis
	redis := getRedis(queueName)

	//检测队列上限值
	listLen, err := redis.Do("lLen", getKey("msgIdList"))

	if err != nil {
		returnValue.Ret = 9999
		return err
	}

	preListLen, err := redis.Do("lLen", getKey("msgIdPreList"))

	if err != nil {
		returnValue.Ret = 9999
		return err
	}

	if listLen.(int64)+preListLen.(int64) > 10000 {
		returnValue.Ret = 10001
		return errors.New("redis队列长度超过限制!")
	}

	//检测8K大小限制
	if len(msgBody) > 8192 {
		returnValue.Ret = 10001
		return errors.New("内容过长!")
	}

	//消息不能为空
	if len(msgBody) == 0 {
		returnValue.Ret = 10001
		return errors.New("内容不能为空!")
	}

	return nil
}

/**
 * 获取请求的唯一编号
 */
func getReqId() string {

	reqString := fmt.Sprintf("%s%d", time.Now().UnixNano(), rand.New(rand.NewSource(time.Now().UnixNano())))

	//获取消息的md5值
	h := crypto.MD5.New()
	h.Write([]byte(reqString))
	reqId := hex.EncodeToString(h.Sum(nil))

	return reqId
}

/**
 * 获取消息的唯一id
 */
func getMsgId() string {
	//获取消息的md5值
	h := crypto.MD5.New()
	h.Write([]byte(msgBody))
	hash := hex.EncodeToString(h.Sum(nil))

	//获取随机值
	randNum := rand.New(rand.NewSource(time.Now().UnixNano())).Int63()

	msgId := fmt.Sprintf("%d%s%d", time.Now().Unix(), hash, randNum)

	return msgId
}
