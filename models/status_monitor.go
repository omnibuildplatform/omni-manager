package models

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"omni-manager/util"
	"time"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second
	//job status
	JOB_STATUS_RUNNING = "running"
	JOB_STATUS_SUCCEED = "succeed"
	JOB_STATUS_FAILED  = "failed"
)

// write Message to Client
func writeMessage2Client(ws *websocket.Conn, jobname string) {
	result := make(map[string]interface{}, 0)

	defer func() {
		ws.Close()
	}()

	heart := make(map[string]interface{})
	//send heart data
	go func() {
		for {
			time.Sleep(time.Second * 30)
			heart["data"] = ""
			heart["code"] = 99
			heartBytes, err := json.Marshal(heart)
			if err = ws.WriteMessage(websocket.TextMessage, heartBytes); err != nil {
				return
			}
		}
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
		} else {
			result["data"] = "/api/v1/images/queryJobStatus/" + jobname
			result["code"] = 1
		}
		resultBytes, _ := json.Marshal(result)
		if err = ws.WriteMessage(websocket.TextMessage, resultBytes); err != nil {
			util.Log.Warnln("4.1 wsQueryJobStatus token :", err)
			return
		}
		time.Sleep(time.Second)
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
		if err = ws.WriteMessage(websocket.TextMessage, resultBytes); err != nil {
			util.Log.Warnln("5 WriteMessage token :", err)
			return
		}
		if reTry < 30 {
			time.Sleep(time.Second)
			goto queryNext
		}
	}
	// buf := new(bytes.Buffer)
	if len(pods.Items) == 0 {
		result["data"] = []byte("no items in this job name:" + jobname)
		result["code"] = -1
		resultBytes, err := json.Marshal(result)
		if err = ws.WriteMessage(websocket.TextMessage, resultBytes); err != nil {
			util.Log.Warnln("6 WriteMessage token :", err)
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
		resultBytes, err := json.Marshal(result)
		if err = ws.WriteMessage(websocket.TextMessage, resultBytes); err != nil {
			util.Log.Warnln("7 WriteMessage token :", err)
			return
		}
		if reTry < 30 {
			time.Sleep(time.Second)
			goto queryNextLog
		}
	}
	defer podLogs.Close()
	tempBytes := make([]byte, 1024)
	for {
		n, err := podLogs.Read(tempBytes)
		if err != nil {
			CheckPodStatus(util.GetConfig().K8sConfig.Namespace, jobname)
			//----------------------------------------
			// if some  err occured,then tell client to call follow api to query job status
			result["data"] = "/api/v1/images/queryJobStatus/" + jobname
			result["code"] = 1
			resultBytes, err := json.Marshal(result)
			if err = ws.WriteMessage(websocket.TextMessage, resultBytes); err != nil {
				util.Log.Warnln("8.9 close websocket :", err)
				break
			}
			util.Log.Warnln("9 close websocket :", err)
			return
		}
		if n > 0 {
			result["data"] = string(tempBytes[:n])
			result["code"] = 0
			resultBytes, err := json.Marshal(result)
			if err = ws.WriteMessage(websocket.TextMessage, resultBytes); err != nil {
				break
			}
		}
	}
	// }
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
		Subprotocols:    []string{r.Header.Get("Sec-WebSocket-Protocol")},
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
	writeMessage2Client(ws, jobname)
}

func StartWebSocket() {
	http.HandleFunc("/wsQueryJobStatus", wsQueryJobStatus)
	http.HandleFunc("/ws/queryJobStatus", wsQueryJobStatus)
	addr := fmt.Sprintf(":%d", util.GetConfig().WSConfig.Port)
	util.Log.Errorf("websocket Listening and serving at %s port ", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		util.Log.Fatalf("websocket startup failed ,error: %s", err)
	}
}
