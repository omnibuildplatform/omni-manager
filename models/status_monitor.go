package models

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"omni-manager/util"
	"strconv"
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

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return util.GetConfig().WSConfig.CheckOrigin
		},
	}
)

// write Message to Client
func writeMessage2Client(ws *websocket.Conn, jobDBID, jobname string) {
	result := make(map[string]interface{}, 0)
	jobid, _ := strconv.Atoi(jobDBID)
	if jobid <= 0 {
		result["data"] = "job id not integer"
		result["code"] = -1
		resultBytes, err := json.Marshal(result)
		if err = ws.WriteMessage(websocket.TextMessage, resultBytes); err != nil {
			return
		}
	}
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
		if reTry < 10 {
			result["data"] = "----------checking job status ----"
			result["code"] = 0
		} else {
			result["data"] = "/api/v1/images/queryJobStatus/" + jobname
			result["code"] = 1
		}
		resultBytes, err := json.Marshal(result)
		if err = ws.WriteMessage(websocket.TextMessage, resultBytes); err != nil {
			return
		}
		time.Sleep(time.Second)
		if reTry < 10 {
			goto checkJobStatus
		} else {
			return
		}
	}

	listopt := metav1.ListOptions{}
	listopt.LabelSelector = "job-name=" + jobname
	pods, err := GetClientSet().CoreV1().Pods(util.GetConfig().K8sConfig.Namespace).List(context.TODO(), listopt)
	if err != nil {
		result["data"] = err.Error()
		result["code"] = -1
		resultBytes, err := json.Marshal(result)
		if err = ws.WriteMessage(websocket.TextMessage, resultBytes); err != nil {
			return
		}
		return
	}

	// buf := new(bytes.Buffer)
	if len(pods.Items) == 0 {
		result["data"] = []byte("no items in this job name:" + jobname)
		result["code"] = -1
		resultBytes, err := json.Marshal(result)
		if err = ws.WriteMessage(websocket.TextMessage, resultBytes); err != nil {
			return
		}
	}
	// for _, pod := range pods.Items {
	pod := pods.Items[0]
	req := GetClientSet().CoreV1().Pods(util.GetConfig().K8sConfig.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Follow: true})
	podLogs, err := req.Stream(context.TODO())

	if err != nil {
		result["data"] = err.Error()
		result["code"] = -1
		resultBytes, err := json.Marshal(result)
		if err = ws.WriteMessage(websocket.TextMessage, resultBytes); err != nil {
			return
		}
		return
	}
	defer podLogs.Close()
	tempBytes := make([]byte, 100)
	for {
		n, err := podLogs.Read(tempBytes)
		if err != nil {
			//wait some seconds for update the job status
			time.Sleep(time.Second * 5)
			// fmt.Println("-----------test call api-----", err)
			// resp, err := http.Get("http://localhost:8080/api/v1/images/queryJobStatus/" + jobname)
			// jobstatusResp, _ := ioutil.ReadAll(resp.Body)
			// defer resp.Body.Close()
			// fmt.Println(string(jobstatusResp))
			//----------------------------------------
			// if some  err occured,then tell client to call follow api to query job status
			result["data"] = "/api/v1/images/queryJobStatus/" + jobname
			result["code"] = 1
			resultBytes, err := json.Marshal(result)
			if err = ws.WriteMessage(websocket.TextMessage, resultBytes); err != nil {
				break
			}
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
	token := r.URL.Query().Get("token")
	if token != "tokentest" {
		return
	}
	jobname := r.URL.Query().Get("jobname")
	if jobname == "" {
		return
	}
	jobDBID := r.URL.Query().Get("jobDBID")
	if jobDBID == "" {
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}
	writeMessage2Client(ws, jobDBID, jobname)
}

func StartWebSocket() {
	http.HandleFunc("/wsQueryJobStatus", wsQueryJobStatus)
	addr := fmt.Sprintf(":%d", util.GetConfig().WSConfig.Port)
	util.Log.Errorf("websocket Listening and serving at %s port ", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		util.Log.Fatalf("websocket startup failed ,error: %s", err)
	}
}
