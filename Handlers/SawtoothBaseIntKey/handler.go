package SawtoothBaseIntKey

import (
	"fmt"
	"github.com/jffp113/Thesis_Client/Client"
	"github.com/jffp113/Thesis_Client/conf"
	"time"
)

type sawtoothHandler struct {
	cli IntkeyClient
}

func (h *sawtoothHandler) InitHandler(config conf.Configuration) {
	cli, err := NewIntkeyClient("http://localhost:8008", "")
	//cli,err := NewIntkeyClient(DEFAULT_URL,"")
	if err != nil {
		panic(err)
	}
	h.cli = cli
	h.cli.Set("johny", 1, 1)
}

func (h sawtoothHandler) DoRequest() Client.RequestStatus {
	var stats Client.RequestStatus
	stats.StartTime = time.Now()
	_, err := h.cli.Inc("johny", 1, 1000) //big number to wait forever
	stats.EndTime = time.Now()
	if err == nil {
		stats.Success = true
	} else {
		fmt.Println(err)
	}

	return stats
}

func NewHandler() Client.Handler {
	return &sawtoothHandler{}
}
