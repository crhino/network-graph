package handlers

import (
	"net/http"

	"github.com/crhino/network-graph/network"
	"github.com/pivotal-golang/lager"
)

type dotHandler struct {
	logger  lager.Logger
	network network.Graph
}

func NewDOTHandler(logger lager.Logger, network network.Graph) http.Handler {
	return &dotHandler{
		logger:  logger,
		network: network,
	}
}

func (n *dotHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	data, err := n.network.EncodeDOT()
	if err != nil {
		n.logger.Error("error-sending-network-state", err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
	}

	resp.Write(data)
}
