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
	for {
		select {
		case <-pingTicker.C:
			var p []byte
			jobAPI := GetClientSet().BatchV1()
			job, err := jobAPI.Jobs(metav1.NamespaceDefault).Get(context.TODO(), jobname, metav1.GetOptions{})
			if err != nil {
				if err = ws.WriteMessage(websocket.TextMessage, []byte(err.Error())); err != nil {
					return
				}
				return
			}
			completions := job.Spec.Completions
			backoffLimit := job.Spec.BackoffLimit

			result := make(map[string]interface{})
			result["name"] = jobname
			result["startTime"] = job.Status.StartTime
			// check status
			if job.Status.Succeeded > *completions {
				result["status"] = JOB_STATUS_SUCCEED
				result["completionTime"] = job.Status.CompletionTime
			} else if job.Status.Failed > *backoffLimit {
				result["status"] = JOB_STATUS_FAILED
			} else if job.Status.Succeeded == 0 || job.Status.Failed == 0 {
				result["status"] = JOB_STATUS_RUNNING
			}

			p, _ = json.Marshal(result)
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.TextMessage, p); err != nil {
				return
			}
			// close websocket if status not qual running
			if result["status"] != JOB_STATUS_RUNNING {
				ws.Close()
				var updateMetaData Metadata
				updateMetaData.Status = result["status"].(string)
				updateMetaData.Id = jobid
				UpdateJobStatus(&updateMetaData)
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
	util.Log.Warnf("websocket Listening and serving at %s port ", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		util.Log.Errorf("websocket startup failed ,error: %s", err)
	}
}
