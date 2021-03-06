package main

import (
	"fmt"
	"log"
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	host "github.com/libp2p/go-libp2p-host"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/open-services/bolivar/cli"
	http "github.com/open-services/bolivar/http"
	"github.com/open-services/bolivar/p2p"
	"github.com/open-services/bolivar/update"
)

// split to packages
var appConfig cli.Config

// What the version of the application is
var appVersion = "0.1.0"

// What the author of the application is
var appAuthor = "Open-Registry"

// libp2p peer to use for transfer
var libp2pNode *ipfslite.Peer

func main() {

	appConfig = cli.Init(appVersion, appAuthor)

	if appConfig.Verbose {
		cli.PrintConfig(appConfig)
	}

	updater := update.NewUpdater(appConfig)

	go func() {
		for {
			time.Sleep(1 * time.Second)
			http.SetRootHash(updater.Update())
		}
	}()
	http.SetRootHash(updater.Update())

	libp2p, h := p2p.StartLibp2p(appConfig)
	libp2pNode = libp2p

	addrs := h.Addrs()
	fmt.Println("libp2p listening on the following addresses:")
	for _, addr := range addrs {
		fmt.Println(addr)
	}

	connectAndReconnect(libp2pNode, h, appConfig.FederateAddr)

	// TODO should start immediately and hang requests until we have the libp2p
	// node ready and connected
	log.Fatal(http.StartServer(appConfig, libp2pNode))
}

// Ensures we stay connected to this peer.
// In case where peer got disconnected, try to reconnect
func connectAndReconnect(libp2pNode *ipfslite.Peer, host host.Host, ma string) {
	err := p2p.ConnectToPeer(libp2pNode, host, ma)
	if err != nil {
		panic(err)
	}
	// make sure we're connected to federate addr always
	go func() {
		fedID, err := getPeerIDFromMultiaddr(ma)
		if err != nil {
			panic(err)
		}
		for {
			time.Sleep(5 * time.Second)
			conns := host.Network().Conns()

			foundOR := false
			for _, conn := range conns {
				if conn.RemotePeer().String() == fedID {
					foundOR = true
				}
			}

			if !foundOR {
				log.Println("Lost connection to OR for some reason")
				log.Println("reconnecting")
				err := p2p.ConnectToPeer(libp2pNode, host, appConfig.FederateAddr)
				if err != nil {
					panic(err)
				}
				log.Println("reconnected")
			}
		}
	}()

}

func getPeerIDFromMultiaddr(maStr string) (string, error) {
	ma, err := multiaddr.NewMultiaddr(maStr)
	if err != nil {
		return "", err
	}
	// Protocol ID from https://github.com/multiformats/go-multiaddr/blob/e1825f7b50d1dcebdaa28bc31a310fa2be4c00ee/protocols.go#L17

	fedID, err := ma.ValueForProtocol(0x01A5)
	if err != nil {
		return "", err
	}
	return fedID, nil
}
