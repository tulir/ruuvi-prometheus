package metrics

import (
	"encoding/hex"
	"encoding/json"
	"maps"
	"net/http"
	"strings"
	"sync"
	"time"
)

type GatewayTagData struct {
	RSSI      int8   `json:"rssi"`
	Timestamp int64  `json:"timestamp"`
	Data      string `json:"data"`
}

type GatewayResponse struct {
	Timestamp   int64                     `json:"timestamp"`
	GatewayMAC  string                    `json:"gw_mac"`
	Tags        map[string]GatewayTagData `json:"tags"`
	Coordinates string                    `json:"coordinates,omitempty"`
}

var data = make(map[string]GatewayTagData)
var dataLock sync.RWMutex
var GatewayMAC string

func ObserveRaw(o RuuviReading) {
	dataLock.Lock()
	defer dataLock.Unlock()
	data[strings.ToUpper(o.Address.String())] = GatewayTagData{
		RSSI:      o.Rssi,
		Timestamp: time.Now().Unix(),
		// TODO don't hardcode the prefix
		Data: "0201061BFF" + strings.ToUpper(hex.EncodeToString(o.Raw)),
	}
}

type WrappedResponse struct {
	Data *GatewayResponse `json:"data"`
}

func ruuviGatewayHandler(w http.ResponseWriter, r *http.Request) {
	dataLock.RLock()
	dataVals := maps.Clone(data)
	dataLock.RUnlock()
	resp := &GatewayResponse{
		Timestamp:  time.Now().Unix(),
		GatewayMAC: GatewayMAC,
		Tags:       dataVals,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&WrappedResponse{Data: resp})
}
