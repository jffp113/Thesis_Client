package Client

import (
	"github.com/jffp113/Thesis_Client/conf"
	"time"
)

//RequestStatus used to return the information of
// method call from DoRequest
type RequestStatus struct {
	StartTime time.Time
	EndTime   time.Time
	Success   bool
}

type StatusPrinter interface {
	Print(stats Stats)
}

//Handler interface to be implemented to be able
//to send requests
type Handler interface {
	//DoRequest Should send a request to the system under test
	//and return the result of a single execution
	//DoRequest Should be reentrant
	DoRequest() RequestStatus

	//InitHandler is executed a single time
	//The Handler should init everything
	//needed here
	InitHandler(config conf.Configuration)
}
