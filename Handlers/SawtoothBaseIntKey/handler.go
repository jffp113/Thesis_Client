package SawtoothBaseIntKey

import (
	"fmt"
	"github.com/jffp113/Thesis_Client/Client"
	"github.com/jffp113/Thesis_Client/conf"
	"math/rand"
	"time"
)

type sawtoothHandler struct {
	cli []IntkeyClient
}

func (h *sawtoothHandler) InitHandler(config conf.Configuration) {
	var cli []IntkeyClient
	for _,url := range config.Conf.ValidatorNodes {
		c, err := NewIntkeyClient(fmt.Sprintf("http://%v",url), "")
		if err != nil {
			panic(err)
		}
		cli = append(cli,c)
	}
	h.cli = cli
	h.cli[0].Set("johny", 1, 1) //We should check if config ValidatorNodes is empty
}

func (h sawtoothHandler) DoRequest() Client.RequestStatus {
	var stats Client.RequestStatus
	_,cli := chooseOne(h.cli)
	stats.StartTime = time.Now()
	_, err := cli.Inc("johny", 1, 1000) //big number to wait forever
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

func chooseOne(v []IntkeyClient) (int, IntkeyClient) {
	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)

	pos := rnd.Intn(len(v))
	return pos, v[pos]
}