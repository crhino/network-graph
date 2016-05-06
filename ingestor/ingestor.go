package ingestor

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/crhino/network-graph/network"
	"github.com/pivotal-golang/lager"
)

const firehoseSubscriptionId = "firehose-a"

type FirehoseIngestor struct {
	dopplerAddr string
	authToken   string
	network     network.Graph
	logger      lager.Logger
}

func NewFirehoseIngestor(logger lager.Logger, addr string, token string, network network.Graph) *FirehoseIngestor {
	return &FirehoseIngestor{
		logger:      logger,
		dopplerAddr: addr,
		authToken:   token,
		network:     network,
	}
}

func (fi *FirehoseIngestor) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	consumer := consumer.New(fi.dopplerAddr, &tls.Config{InsecureSkipVerify: true}, nil)
	consumer.SetDebugPrinter(ConsoleDebugPrinter{})
	msgChan, errorChan := consumer.Firehose(firehoseSubscriptionId, fi.authToken)

	close(ready)

	for {
		select {
		case msg := <-msgChan:
			var remoteAddr string
			var peerType events.PeerType
			ip := network.IP(msg.GetIp())
			if ip == "" {
				continue
			}

			switch msg.GetEventType() {
			case events.Envelope_HttpStart:
				event := msg.GetHttpStart()
				if event == nil {
					continue
				}
				remoteAddr = event.GetRemoteAddress()
				peerType = event.GetPeerType()
			case events.Envelope_HttpStartStop:
				event := msg.GetHttpStartStop()
				if event == nil {
					continue
				}
				remoteAddr = event.GetRemoteAddress()
				peerType = event.GetPeerType()
			default:
				continue
			}

			if remoteAddr == "" {
				continue
			}

			remoteIp := network.IP(strings.Split(remoteAddr, ":")[0])

			fi.logger.Debug("adding-edge", lager.Data{"source": ip, "target": remoteIp})

			fi.network.AddNode(ip, fmt.Sprintf("%s/%s", msg.GetJob(), msg.GetIndex()))
			fi.network.AddNode(remoteIp, "")

			switch peerType {
			case events.PeerType_Client:
				fi.network.AddEdge(ip, remoteIp)
			case events.PeerType_Server:
				fi.network.AddEdge(remoteIp, ip)
			}

		case err := <-errorChan:
			return err
		case <-signals:
			return nil
		}
	}
}

type ConsoleDebugPrinter struct{}

func (c ConsoleDebugPrinter) Print(title, dump string) {
	println(title)
	println(dump)
}
