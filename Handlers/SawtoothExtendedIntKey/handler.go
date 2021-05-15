package SawoothExtendedIntKey

import (
	"fmt"
	"github.com/jffp113/Thesis_Client/Client"
	"github.com/jffp113/Thesis_Client/conf"
	"math/rand"
	"time"
)

type sawtoothXHandler struct {
	cli            IntkeyClient
	signerNodesURL []string
	validatorURL   []string
}

func (h *sawtoothXHandler) InitHandler(config conf.Configuration) {
	cli, err := NewIntkeyClient("", config.Conf.KeyPath, config.Conf.KeyName)

	if err != nil {
		panic(err)
	}
	h.signerNodesURL = config.Conf.SignerNodes
	h.validatorURL = config.Conf.ValidatorNodes

	h.cli = cli

	i, validator := chooseOne(h.validatorURL)
	h.cli.Set("alice", 1, 1, validator, h.signerNodesURL[i])
}

func (h sawtoothXHandler) DoRequest() Client.RequestStatus {
	var stats Client.RequestStatus
	i, validator := chooseOne(h.validatorURL)
	stats.StartTime = time.Now()
	_, err := h.cli.Inc("alice", 1, 1000, validator, h.signerNodesURL[i]) //big number to wait forever
	stats.EndTime = time.Now()
	if err == nil {
		stats.Success = true
	} else {
		fmt.Println(err)
	}

	return stats
}

func chooseOne(v []string) (int, string) {
	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)

	pos := rnd.Intn(len(v))
	return pos, v[pos]
}

func NewHandler() Client.Handler {
	return &sawtoothXHandler{}
}
