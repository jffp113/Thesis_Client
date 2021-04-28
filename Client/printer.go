package Client

import (
	"fmt"
	"time"
)

type DefaultPrinter struct {
}

func (DefaultPrinter) Print(stats Stats) {
	avgDur := stats.TotDuration / time.Duration(stats.ClientsResponses)

	throughput := float64(stats.NumRequests) / avgDur.Seconds()
	avgLatency := stats.TotDuration / time.Duration(stats.NumRequests)

	fmt.Printf("%v requests in %v\n", stats.NumRequests, avgDur)
	fmt.Printf("Transactions/sec:\t\t%.2f s\n", throughput)
	fmt.Printf("Avg Latency:\t\t\t%v\n", avgLatency)
	fmt.Printf("Min Transactiom Latency:\t%v\n", stats.MinRequestTime)
	fmt.Printf("Max Transactiom Latency:\t%v\n", stats.MaxRequestTime)
	fmt.Printf("Number of Errors:\t\t%v\n", stats.NumErrs)
}
