package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/crhino/network-graph/handlers"
	"github.com/crhino/network-graph/ingestor"
	"github.com/crhino/network-graph/network"
	"github.com/gorilla/mux"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var (
	dopplerAddress = os.Getenv("DOPPLER_ADDR")
	authToken      = os.Getenv("CF_ACCESS_TOKEN")
)

var port = flag.Uint("port", 8080, "Port to run server on")

type IP string

func main() {
	cf_lager.AddFlags(flag.CommandLine)
	flag.Parse()
	logger, _ := cf_lager.New("network-graph")

	network := network.NewGraph()
	firehoseIngestor := ingestor.NewFirehoseIngestor(logger.Session("ingestor"), dopplerAddress, authToken, network)

	networkHandler := handlers.NewNetworkHandler(logger.Session("network-handler"), network)
	dotHandler := handlers.NewDOTHandler(logger.Session("dot-handler"), network)
	r := mux.NewRouter()
	r.Handle("/network", networkHandler)
	r.Handle("/network/dot", dotHandler)
	server := http_server.New(":"+strconv.Itoa(int(*port)), r)

	members := grouper.Members{
		{"ingestor", firehoseIngestor},
		{"server", server},
	}

	group := grouper.NewOrdered(os.Interrupt, members)
	process := ifrit.Invoke(sigmon.New(group))

	logger.Info("started")

	errChan := process.Wait()
	err := <-errChan
	if err != nil {
		logger.Error("shutdown-error", err)
		os.Exit(1)
	}
	logger.Info("exited")
}
