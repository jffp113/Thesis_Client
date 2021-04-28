package SignerNode

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jffp113/Thesis_Client/Client"
	"github.com/jffp113/Thesis_Client/Handlers/SawtoothExtendedIntKey/pb"
	"github.com/jffp113/Thesis_Client/conf"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"time"
)

type signerNode struct {
}

func (h signerNode) InitHandler(config conf.Configuration) {
	//TODO here you should parse config
	//For example parse config to get url, tls settings, etc,..
}

func (h signerNode) DoRequest() Client.RequestStatus {
	var stats Client.RequestStatus

	err := performTransaction(&stats)

	if err == nil {
		stats.Success = true
	}

	return stats
}

func NewHandler() Client.Handler {
	return signerNode{}
}

func performTransaction(stats *Client.RequestStatus) error {

	uuid := fmt.Sprint(uuid.NewV4())
	msg := pb.ClientSignMessage{
		UUID:                 fmt.Sprint(uuid),
		Content:              []byte("Hello"),
		SmartContractAddress: "intkey",
	}

	b, err := proto.Marshal(&msg)

	reader := bytes.NewReader(b)

	//fmt.Println(uuid)
	stats.StartTime = time.Now()
	_, err = http.Post("http://localhost:8080/sign", "application/protobuf", reader)
	stats.EndTime = time.Now()

	//fmt.Println(resp)
	//body,_ :=ioutil.ReadAll(resp.Body)
	//fmt.Println(body)

	return err
}
