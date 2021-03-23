package Client

import (
	"context"
	"fmt"
	"github.com/jffp113/Thesis_Client/Client/util"
	"time"
)

type requester struct {
	duration           	time.Duration //seconds
	concurrentClients  	int
	configFilePath 		string
	handlers 			map[string]Handler
	ctx					context.Context
	cancelFunc          context.CancelFunc
	printer				StatusPrinter
}

type Stats struct {
	TotDuration      time.Duration
	MinRequestTime   time.Duration
	MaxRequestTime   time.Duration
	NumRequests      int
	NumErrs          int
	ClientsResponses int
}

func NewRequester() requester{
	return requester{ctx: context.Background(),
					printer: DefaultPrinter{},
					handlers: make(map[string]Handler)}
}

func (r *requester) SetDuration(duration time.Duration) {
	r.duration = duration
}

func (r *requester) SetConcurrentClients(clients int) {
	r.concurrentClients = clients
}

func (r *requester) SetConfigFilePath(path string) {
	r.configFilePath = path
}

func (r *requester) AddHandler(key string,handler Handler){
	r.handlers[key] = handler
}

func (r *requester) Start(handlerName string){
	ctx,cancel := context.WithCancel(r.ctx)

	r.cancelFunc = cancel

	handler,ok := r.handlers[handlerName]

	if !ok {
		panic("Handler Does not exit")
	}

	handler.InitHandler(r.configFilePath)

	responseChan := make(chan Stats)
	for i := 0 ; i < r.concurrentClients; i++{
		go r.worker(handler,ctx,responseChan)
	}

	r.aggregateResponses(responseChan)
}

func (r *requester) Stop() error {
	if r.cancelFunc == nil {
		return fmt.Errorf("requester not started")
	}
	r.cancelFunc()
	return nil
}

func (r *requester) aggregateResponses(responseChan <-chan Stats) {
	aggregatedStats := Stats{MinRequestTime: time.Hour}
	for s := range responseChan {
		aggregatedStats.NumErrs += s.NumErrs
		aggregatedStats.NumRequests += s.NumRequests
		aggregatedStats.TotDuration += s.TotDuration
		aggregatedStats.MaxRequestTime = util.MaxDuration(aggregatedStats.MaxRequestTime,s.MaxRequestTime)
		aggregatedStats.MinRequestTime = util.MinDuration(aggregatedStats.MinRequestTime,s.MinRequestTime)
		aggregatedStats.ClientsResponses++

		if aggregatedStats.ClientsResponses >= r.concurrentClients {
			break
		}
	}

	r.printer.Print(aggregatedStats)
}


func (r *requester) worker(handler Handler, ctx context.Context, responseChan chan<- Stats) {
	stats := Stats{MinRequestTime: time.Hour} //TODO improve Min

	start := time.Now()



	for {
		select {
			case <-ctx.Done():
				responseChan<-stats
				return
			default:
				if time.Since(start) > r.duration {
					responseChan<-stats
					return
				}

				s := handler.DoRequest()
				duration := s.EndTime.Sub(s.StartTime)
				stats.TotDuration += duration
				stats.MaxRequestTime = util.MaxDuration(stats.MaxRequestTime,duration)
				stats.MinRequestTime = util.MinDuration(stats.MinRequestTime,duration)
				stats.NumRequests++
				if !s.Success {
					stats.NumErrs++
				}

		}
	}

}