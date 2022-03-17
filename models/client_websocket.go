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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second
	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Poll file for changes with this period.
	filePeriod = 10 * time.Second
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

func writer(ws *websocket.Conn, jobname string) {
	pingTicker := time.NewTicker(pingPeriod)
	fileTicker := time.NewTicker(filePeriod)
	defer func() {
		pingTicker.Stop()
		fileTicker.Stop()
		ws.Close()
	}()
	for {
		select {
		case <-fileTicker.C:
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
			const JOB_STATUS_RUNNING = "running"
			const JOB_STATUS_SUCCEED = "succeed"
			const JOB_STATUS_FAILED = "failed"
			result := make(map[string]interface{})
			result["name"] = jobname
			result["startTime"] = job.Status.StartTime

			// check status
			if job.Status.Succeeded > *completions {
				result["status"] = JOB_STATUS_SUCCEED
				result["completionTime"] = job.Status.CompletionTime
			} else if job.Status.Failed > *backoffLimit {
				result["status"] = JOB_STATUS_FAILED
				result["completionTime"] = job.Status.CompletionTime
			} else if job.Status.Succeeded == 0 || job.Status.Failed == 0 {
				result["status"] = JOB_STATUS_RUNNING
			}

			p, _ = json.Marshal(result)

			if p != nil {
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				if err := ws.WriteMessage(websocket.TextMessage, p); err != nil {

					return
				}
			}
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

//connect each websocket
func QueryJobStatus(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token != "tokentest" {
		return
	}
	jobname := r.URL.Query().Get("jobname")
	if jobname == "" {
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}
	writer(ws, jobname)
	// reader(ws)
}

func StartWebSocket() {
	http.HandleFunc("/wsQueryJobStatus", QueryJobStatus)
	addr := fmt.Sprintf(":%d", util.GetConfig().WSConfig.Port)
	util.Log.Warnf("websocket start at %s port ", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		util.Log.Errorf("websocket startup failed ,error: %s", err)
	}
}
