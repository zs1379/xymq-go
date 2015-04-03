package operation

import (
	log4go "code.google.com/p/log4go"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

/**
 * 方法处理，用于某种方法的调用
 */
func OperationHandle(w http.ResponseWriter, method string, param string) (string, error) {
	//解析json
	var jsonData interface{}
	err := json.Unmarshal([]byte(param), &jsonData)
	if err != nil {
		log4go.Error("json解析失败哟,err:", err)
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
				return "", err
			}

			return result, nil
		} else {
			log4go.Error(fmt.Sprintf("method:push,参数长度异常%s", params))
			return "", errors.New("参数长度异常")
		}
	default:
		return "", errors.New("无有效方法")
	}
}

/**
 * 推入队列
 */
func push(queueName string, messageBody string) (string, error) {
	return "success", nil
}
