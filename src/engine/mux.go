package engine

import (
	"net/http"
	"log"
	"encoding/json"
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
	Key K`json:"key"`
	Value V`json:"value","omitempty"`
}

func (t *Type[K, V]) DeleteKey(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	 
	var response KVreq[K, V] 
	err := json.NewDecoder(req.Body).Decode(&response)
	if err != nil {
		log.Println("delete key handler:", err)
		return
	}

	delete(t.Data, response.Key)
	t.VisualHandler(w, req)
	t.VisualOldHandler(w, req)
	t.HmapHandler(w, req)
	ws.NotifyUpdate()
	log.Println("delete key ok!")
}

func (t *Type[K, V]) UpdateKey(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var response KVreq[K, V]
	err := json.NewDecoder(req.Body).Decode(&response)
	if err != nil {
		log.Println("update key handler:", err)
		return
	}

	t.Data[response.Key] = response.Value
	t.VisualHandler(w, req)
	t.VisualOldHandler(w, req)
	t.HmapHandler(w, req)
	ws.NotifyUpdate()
	log.Println("update key ok")
}
