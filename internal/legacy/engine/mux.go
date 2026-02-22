package engine

import (
	"encoding/json"
	"net/http"

	"visualizer/internal/ws"
)

func (t *Meta[K, V]) VisualHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp, err := t.GetBucketsJSON(MapBucketNew)
	if err != nil {
		t.Console.PrintlnLogError(err)
	}
	w.Write(resp)
}

func (t *Meta[K, V]) VisualOldHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp, err := t.GetBucketsJSON(MapBucketOld)
	if err != nil {
		t.Console.PrintlnLogError(err)
	}
	w.Write(resp)
}

func (t *Meta[K, V]) HmapHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	res, err := GetHmapJSON(GetHmap(t.Map))
	if err != nil {
		t.Console.PrintlnLogError(err)
	}
	w.Write(res)
}

type KVreq[K comparable, V any] struct {
	Key   K `json:"key"`
	Value V `json:"value,omitempty"`
}

func (t *Meta[K, V]) DeleteKey(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response KVreq[K, V]
	err := json.NewDecoder(req.Body).Decode(&response)
	if err != nil {
		t.Console.PrintlnLogError(err)
		return
	}

	delete(t.Map, response.Key)
	t.VisualHandler(w, req)
	t.VisualOldHandler(w, req)
	t.HmapHandler(w, req)
	ws.NotifyUpdate()
	t.Console.PrintlnLogGood("The key has been successfully deleted")
}

func (t *Meta[K, V]) UpdateKey(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response KVreq[K, V]
	err := json.NewDecoder(req.Body).Decode(&response)
	if err != nil {
		t.Console.PrintlnLogError(err)
		return
	}

	t.Map[response.Key] = response.Value
	t.VisualHandler(w, req)
	t.VisualOldHandler(w, req)
	t.HmapHandler(w, req)
	ws.NotifyUpdate()
	t.Console.PrintlnLogGood("The key has been successfully updated")
}
