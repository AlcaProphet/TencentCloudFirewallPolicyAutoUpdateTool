package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alcaprophet/fwalizer/syncer"
)

func (d *Deps) handleSyncStatus(w http.ResponseWriter, r *http.Request) {
	if d.Syncer == nil {
		writeJSON(w, http.StatusOK, syncer.SyncStatus{Running: false})
		return
	}
	writeJSON(w, http.StatusOK, d.Syncer.Status())
}

func (d *Deps) handleSyncTrigger(w http.ResponseWriter, r *http.Request) {
	if d.Syncer == nil {
		writeError(w, http.StatusBadRequest, "同步引擎未启动，请先配置目标和规则")
		return
	}
	d.Syncer.TriggerSync()
	writeJSON(w, http.StatusAccepted, map[string]string{"message": "同步已触发"})
}

func (d *Deps) handleSyncDryRun(w http.ResponseWriter, r *http.Request) {
	if d.Syncer == nil {
		writeError(w, http.StatusBadRequest, "同步引擎未启动，请先配置目标和规则")
		return
	}
	results, err := d.Syncer.DryRun()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, results)
}

// handleSyncEvents SSE 实时事件推送
func (d *Deps) handleSyncEvents(w http.ResponseWriter, r *http.Request) {
	if d.EventBus == nil {
		writeError(w, http.StatusBadRequest, "事件总线不可用")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "SSE 不可用")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, unsubscribe := d.EventBus.SubscribeChan()
	defer unsubscribe()

	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				return // channel 已关闭
			}
			data, err := json.Marshal(ev)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (d *Deps) handleGetSyncLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := d.Store.GetSyncLogs(100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, logs)
}
