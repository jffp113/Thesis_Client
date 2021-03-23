package SawoothExtendedIntKey

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jffp113/Thesis_Client/Client"
	"github.com/jffp113/Thesis_Client/Handlers/SawtoothExtendedIntKey/pb"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"time"
)

type httpHandler struct{
}

func (h httpHandler) InitHandler(configFilePath string) {
	//TODO here you should parse config
	//For example parse config to get url, tls settings, etc,..
}

func (h httpHandler) DoRequest() Client.RequestStatus {
	var stats Client.RequestStatus
	stats.StartTime = time.Now()

	err := performTransaction(&stats)

	stats.EndTime = time.Now()
	if err == nil {
		stats.Success = true
	}



	return stats
}



func NewHandler() Client.Handler {
	return httpHandler{}
}


func performTransaction(stats *Client.RequestStatus) error {

	uuid := fmt.Sprint(uuid.NewV4())
	msg:=pb.ClientSignMessage{
		UUID:          fmt.Sprint(uuid),
		Content:       []byte("Hello"),
		SmartContractAddress: "intkey",
	}

	b,err := proto.Marshal(&msg)

	reader := bytes.NewReader(b)

	//fmt.Println(uuid)
	stats.StartTime = time.Now()
	_, err = http.Post("http://localhost:8080/sign","application/protobuf",reader)

	stats.EndTime = time.Now()


	//fmt.Println(resp)
	//body,_ :=ioutil.ReadAll(resp.Body)
	//fmt.Println(body)

	return err
}
