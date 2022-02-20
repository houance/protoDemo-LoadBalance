package trace

import (
	"encoding/json"
	"houance/protoDemo-LoadBalance/internal/innerData"
	"net/http"
	"strconv"
)

var (
	slm map[uint32]*innerData.StreamLatency
)

func ExposeToGrafana(
	listenPort int,
	streamIDStreamLatencyMap map[uint32]*innerData.StreamLatency,
) {
	slm = streamIDStreamLatencyMap
	http.HandleFunc("/query", Query)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
	http.ListenAndServe("0.0.0.0:"+strconv.Itoa(listenPort), nil)
}

func Query(w http.ResponseWriter, r *http.Request) {
	streamID, err := strconv.ParseUint(r.Header.Get("streamID"), 10, 32)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	sl, ok := slm[uint32(streamID)]
	if !ok {
		w.Write([]byte("not found"))
		return
	}
	l := sl.GetAverageLatency()

	w.Header().Set("Content-Type", "application/json")
	tmp := make(map[string]int64)
	tmp["OverAll-Latency"] = l[0]
	tmp["Client-LB-Latency"] = l[1]
	tmp["LB-Server-Latency"] = l[2]
	tmp["Server-LB-Latency"] = l[3]
	tmp["LB-Client-Latency"] = l[4]
	data, err := json.Marshal(tmp)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(data)
}
