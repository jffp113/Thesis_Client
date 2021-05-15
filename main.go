package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/jffp113/Thesis_Client/Client"
	"github.com/jffp113/Thesis_Client/Handlers/Algorand"
	"github.com/jffp113/Thesis_Client/Handlers/SawtoothBaseIntKey"
	SawtoothExtendedIntKey "github.com/jffp113/Thesis_Client/Handlers/SawtoothExtendedIntKey"
	"github.com/jffp113/Thesis_Client/Handlers/SignerNode"
	"github.com/jffp113/Thesis_Client/Handlers/SimpleHttp"
	"os"
	"time"
)

type Opts struct {
	ConcurrentClient int    `short:"c" long:"concurrent" default:"1" description:"Number Of Concurrent Clients"`
	Duration         int    `short:"d" long:"duration"   default:"10" description:"Duration in Seconds"`
	Handler          string `short:"a" long:"handler"  default:"http" description:"Handler to be executed"`
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error:", r)
		}
	}()

	var opts Opts

	parser := flags.NewParser(&opts, flags.Default)
	remaining, err := parser.Parse()

	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Printf("Failed to parse args: %v\n", err)
			os.Exit(2)
		}
	}

	if len(remaining) > 0 {
		fmt.Printf("Error: Unrecognized arguments passed: %v\n", remaining)
		os.Exit(2)
	}

	reqCli := Client.NewRequester()
	reqCli.SetConcurrentClients(opts.ConcurrentClient)
	reqCli.SetConfigFilePath("conf.yaml")
	reqCli.SetDuration(time.Second * time.Duration(opts.Duration))

	reqCli.AddHandler("http", SimpleHttp.NewHandler())
	reqCli.AddHandler("sawtooth", SawtoothBaseIntKey.NewHandler())
	reqCli.AddHandler("sawtoothX", SawtoothExtendedIntKey.NewHandler())
	reqCli.AddHandler("signernode", SignerNode.NewHandler())
	reqCli.AddHandler("algorand", Algorand.NewHandler())
	err = reqCli.Start(opts.Handler)

	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(2)
	}

}
