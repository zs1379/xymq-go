package main

import (
	//"encoding/json"
	log4go "code.google.com/p/log4go"
	"fmt"
	"net/http"
	oper "xiaoy.name/xymq/operation"
)

const (
	//配置文件所在位置
	//configPath = "/etc/xyDog.conf"
	//配置日志存放位置
	logFile = "/home/xiaoy/dev/goapp/src/xiaoy.name/xymq/mq.log"
)

func main() {
	//创建日志类
	logOption := log4go.NewFileLogWriter(logFile, false)
	log4go.AddFilter("file", log4go.FINE, logOption)

	mux := http.NewServeMux()

	mux.HandleFunc("/Soa/service", xHandle(serviceHandle))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("listen and server error %s", err)
	}
}

/**
 * 入口处理，用于错误处理、ip鉴权等
 */
func xHandle(xHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				log4go.Warn(fmt.Sprintf("%v %v", xHandler, e))
			}
		}()

		//ip鉴权

		xHandler(w, r)
	}
}

/**
 * 服务处理，用于全局日志记录，分发服务处理
 */
func serviceHandle(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	service := r.Form.Get("service")
	method := r.Form.Get("method")
	param := r.Form.Get("form")
	requestId := r.Form.Get("distinctRequestId")
	basic := r.Form.Get("soa_basic")

	log4go.Info(fmt.Sprintf("ip:%s service:%s method:%s param:%s requestId:%s basic:%s", r.RemoteAddr, service, method, param, requestId, basic))

	switch service {
	case "operation":
		oper.OperationHandle(w, method, param)
	}
}
