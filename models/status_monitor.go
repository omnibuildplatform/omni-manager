package models

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/omnibuildplatform/omni-manager/util"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second
	//job status
	JOB_STATUS_START   = "start"
	JOB_STATUS_RUNNING = "running"
	JOB_STATUS_SUCCEED = "succeed"
	JOB_STATUS_FAILED  = "failed"
	JOB_STATUS_CREATED = "created"
	JOB_STATUS_STOPPED = "stopped"

	JOB_BUILD_STATUS_SUCCEED = "JobSucceed"
)

// write Message to Client
func writeMessage2Client(ws *websocket.Conn, jobname string) {
	result := make(map[string]interface{}, 0)
	defer func() {
		ws.Close()
	}()

	//check job status first
	var reTry = 0
checkJobStatus:
	jobAPI := GetClientSet().BatchV1()
	_, err := jobAPI.Jobs(util.GetConfig().K8sConfig.Namespace).Get(context.TODO(), jobname, metav1.GetOptions{})
	if err != nil {
		//retry 10 times if some err , one time.second each time
		if reTry < 30 {
			result["data"] = "----------checking job status ----\n"
			result["code"] = 0
			result["other"] = err
		} else {
			result["data"] = "/api/v1/images/queryJobStatus/" + jobname
			result["code"] = 1
		}
		resultBytes, _ := json.Marshal(result)
		if err = sendNormalData(ws, resultBytes); err != nil {
			return
		}
		time.Sleep(time.Second)
		reTry++
		if reTry < 30 {
			goto checkJobStatus
		}
	}
	listopt := metav1.ListOptions{}
	listopt.LabelSelector = "job-name=" + jobname
	reTry = 0
queryNext:
	pods, err := GetClientSet().CoreV1().Pods(util.GetConfig().K8sConfig.Namespace).List(context.TODO(), listopt)
	if err != nil {
		result["data"] = err.Error() + "\n"
		result["code"] = -2
		resultBytes, err := json.Marshal(result)
		if err = sendNormalData(ws, resultBytes); err != nil {
			return
		}
		reTry++
		if reTry < 30 {
			time.Sleep(time.Second)
			goto queryNext
		}
	}
	// buf := new(bytes.Buffer)
	if len(pods.Items) == 0 {
		result["data"] = []byte("no items in this job name:" + jobname)
		result["code"] = -1
		resultBytes, _ := json.Marshal(result)
		if err = sendNormalData(ws, resultBytes); err != nil {

			return
		}
	}
	// for _, pod := range pods.Items {
	pod := pods.Items[0]
	req := GetClientSet().CoreV1().Pods(util.GetConfig().K8sConfig.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Follow: true})
	reTry = 0
queryNextLog:
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		result["data"] = err.Error() + "\n"
		result["code"] = -2
		resultBytes, _ := json.Marshal(result)
		if err = sendNormalData(ws, resultBytes); err != nil {
			return
		}
		reTry++
		if reTry < 30 {
			time.Sleep(time.Second)
			goto queryNextLog
		}
	}
	defer podLogs.Close()
	var tempBytes []byte

	for {
		tempBytes = make([]byte, 618)
		n, readErr := podLogs.Read(tempBytes)

		if readErr != nil {
			// if some err occured, check job status .
			statusResult, _, err := CheckPodStatus(util.GetConfig().K8sConfig.Namespace, jobname)
			if err != nil {
				if reTry > 30 {
					result["data"] = "CheckPodStatus error:" + err.Error()
					result["code"] = -1
					resultBytes, _ := json.Marshal(result)
					sendNormalData(ws, resultBytes)
					return
				}
				reTry++
				time.Sleep(time.Second)
				continue
			}
			if statusResult["status"] == JOB_STATUS_FAILED || statusResult["status"] == JOB_STATUS_SUCCEED {
				// if not running statu. then  tell client to call follow api to query job status. and return
				result["data"] = "/api/v1/images/queryJobStatus/" + jobname
				result["code"] = 1
				resultBytes, _ := json.Marshal(result)
				sendNormalData(ws, resultBytes)

				return
			} else {
				// continue read log 30 times
				if reTry > 30 {
					result["data"] = "some error:" + readErr.Error()
					result["code"] = -1
					resultBytes, _ := json.Marshal(result)
					sendNormalData(ws, resultBytes)
					return
				}
				time.Sleep(time.Second)
				reTry++
				continue
			}

		}
		if n > 0 {
			result["data"] = string(tempBytes[:n])
			result["code"] = 0
			resultBytes, _ := json.Marshal(result)
			sendNormalData(ws, resultBytes)
			//reset when normal
			reTry = 0
		}
	}
	// }
}
func sendNormalData(ws *websocket.Conn, msg []byte) error {
	ws.SetWriteDeadline(time.Now().Local().Add(1200 * time.Second))
	err := ws.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		util.Log.Warnln(" websocket sendNormalData Error :", err)
	}
	return err
}

//connect each websocket
func wsQueryJobStatus(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("Sec-WebSocket-Protocol")
	if len(token) < 20 {
		return
	}
	_, err := CheckAuthorization(token)
	if err != nil {
		//非法用户
		util.Log.Warnln(" unAuthing user :", err)
		return
	}
	jobname := r.URL.Query().Get("jobname")
	if jobname == "" {
		return
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		Subprotocols:    []string{token},
		CheckOrigin: func(r *http.Request) bool {
			return util.GetConfig().WSConfig.CheckOrigin
		},
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}
	// ctx, finishFunc := context.WithCancel(context.Background())

	// go readMessageFromClient(ws, ctx, jobname)
	writeMessage2Client(ws, jobname)

}

func StartWebSocket() {
	http.HandleFunc("/wsQueryJobStatus", wsQueryJobStatus)
	http.HandleFunc("/ws/queryJobStatus", wsQueryJobStatus)
	addr := fmt.Sprintf(":%d", util.GetConfig().WSConfig.Port)
	util.Log.Infof("websocket Listening and serving at %s port ", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		util.Log.Fatalf("websocket startup failed ,error: %s", err)
	}
}
