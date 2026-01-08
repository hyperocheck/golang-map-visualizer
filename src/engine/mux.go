package engine

import (
	"encoding/json"
	"fmt"
	"net/http"

	"visualizer/src/logger"
	"visualizer/src/ws"
)

func (t *Type[K, V]) VisualHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(t.GetBucketsJSON("buckets"))
}

func (t *Type[K, V]) VisualOldHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(t.GetBucketsJSON("oldbuckets"))
}

func (t *Type[K, V]) HmapHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	res, _ := GetHmapJSON(t.GetHmap())
	w.Write(res)
}

type KVreq[K comparable, V any] struct {
	Key   K `json:"key"`
	Value V `json:"value","omitempty"`
}

func (t *Type[K, V]) DeleteKey(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response KVreq[K, V]
	err := json.NewDecoder(req.Body).Decode(&response)
	if err != nil {
		logger.Log.Log("error", fmt.Sprintf("DeleteKey handler err: %s", err))
		return
	}

	delete(t.Data, response.Key)
	t.VisualHandler(w, req)
	t.VisualOldHandler(w, req)
	t.HmapHandler(w, req)
	ws.NotifyUpdate()
	logger.Log.Log("info", "delete key ok!")
}

func (t *Type[K, V]) UpdateKey(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response KVreq[K, V]
	err := json.NewDecoder(req.Body).Decode(&response)
	if err != nil {
		logger.Log.Log("error", fmt.Sprintf("UpdateKey handler err: %s", err))
		return
	}

	t.Data[response.Key] = response.Value
	t.VisualHandler(w, req)
	t.VisualOldHandler(w, req)
	t.HmapHandler(w, req)
	ws.NotifyUpdate()
	logger.Log.Log("info", "update key ok!")
}
