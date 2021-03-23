package SimpleHttp

import (
	"github.com/jffp113/Thesis_Client/Client"
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
	_,err := http.Get("https://google.com/")
	stats.EndTime = time.Now()
	if err == nil {
		stats.Success = true
	}

	return stats
}



func NewHandler() Client.Handler {
	return httpHandler{}
}
