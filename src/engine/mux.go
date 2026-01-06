package engine

import "net/http"

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
