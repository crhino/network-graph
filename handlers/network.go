package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/crhino/network-graph/network"
	"github.com/pivotal-golang/lager"
)

type networkHandler struct {
	logger  lager.Logger
	network network.Graph
}

func NewNetworkHandler(logger lager.Logger, network network.Graph) http.Handler {
	return &networkHandler{
		logger:  logger,
		network: network,
	}
}

func (n *networkHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	data, err := json.Marshal(n.network)
	if err != nil {
		n.logger.Error("error-sending-network-state", err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
	}

	resp.Write(data)
}
