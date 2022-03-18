package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = 10 * time.Second
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
		if err := ws.WriteMessage(websocket.TextMessage, []byte("job id not integer")); err != nil {
			return
		}
	}
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()

	listopt := metav1.ListOptions{}
	listopt.LabelSelector = "job-name=" + jobname
	pods, err := GetClientSet().CoreV1().Pods(metav1.NamespaceDefault).List(context.TODO(), listopt)
	if err != nil {
		if err = ws.WriteMessage(websocket.TextMessage, []byte(err.Error())); err != nil {
			return
		}
		return
	}
	// buf := new(bytes.Buffer)
	if len(pods.Items) == 0 {
		if err = ws.WriteMessage(websocket.TextMessage, []byte("no items in this job name:"+jobname)); err != nil {
			return
		}
	}
	// for _, pod := range pods.Items {
	pod := pods.Items[0]
	req := GetClientSet().CoreV1().Pods(metav1.NamespaceDefault).GetLogs(pod.Name, &corev1.PodLogOptions{Follow: true})
	podLogs, err := req.Stream(context.TODO())

	if err != nil {
		if err = ws.WriteMessage(websocket.TextMessage, []byte(err.Error())); err != nil {
			return
		}
		return
	}
	defer podLogs.Close()
	jobAPI := GetClientSet().BatchV1()
	job, err := jobAPI.Jobs(metav1.NamespaceDefault).Get(context.TODO(), jobname, metav1.GetOptions{})
	if err != nil {
		if err = ws.WriteMessage(websocket.TextMessage, []byte(err.Error())); err != nil {
			return
		}
		return
	}

	tempBytes := make([]byte, 100)
	for {
		n, err := podLogs.Read(tempBytes)
		if err != nil {
			if err == io.EOF {
				job, err = jobAPI.Jobs(metav1.NamespaceDefault).Get(context.TODO(), jobname, metav1.GetOptions{})
				if job.Status.Succeeded > *job.Spec.Completions {
					// JOB_STATUS_SUCCEED
					logData := fmt.Sprintf("----------build success ----  ")
					result["data"] = logData
					result["code"] = 1
					// make a full download iso url
					result["url"] = fmt.Sprintf(util.GetConfig().BuildParam.DownloadIsoUrl, jobname)
					resultBytes, err := json.Marshal(result)
					if err = ws.WriteMessage(websocket.TextMessage, resultBytes); err != nil {
						break
					}
					return
				} else if job.Status.Failed > *job.Spec.BackoffLimit {
					// JOB_STATUS_FAILED
					result["data"] = "----------build failed ----"
					result["code"] = -1
					resultBytes, err := json.Marshal(result)
					if err = ws.WriteMessage(websocket.TextMessage, resultBytes); err != nil {
						break
					}
					return
				} else if job.Status.Succeeded == 0 || job.Status.Failed == 0 {
					// runnig status
					//after some time ,check job status
					<-pingTicker.C
				}
			} else {
				if err = ws.WriteMessage(websocket.TextMessage, []byte("----------build error ----"+err.Error())); err != nil {
					break
				}
			}
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
