package SawoothExtendedIntKey

import (
	"fmt"
	"github.com/jffp113/Thesis_Client/Client"
	"time"
)

type sawtoothXHandler struct{
	cli IntkeyClient
}

func (h *sawtoothXHandler) InitHandler(configFilePath string) {
	cli,err := NewIntkeyClient(DEFAULT_URL,"")
	if err != nil {
		panic(err)
	}
	h.cli = cli
	h.cli.Set("alice",1,1)
}

func (h sawtoothXHandler) DoRequest() Client.RequestStatus {
	var stats Client.RequestStatus
	stats.StartTime = time.Now()
	_ , err := h.cli.Inc("alice",1,1000) //big number to wait forever
	stats.EndTime = time.Now()
	if err == nil {
		stats.Success = true
	}else{
		fmt.Println(err)
	}

	return stats
}

func NewHandler() Client.Handler {
	return &sawtoothXHandler{}
}

