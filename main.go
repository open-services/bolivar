package main

import (
	"fmt"
	"log"
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
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

	err := p2p.ConnectToPeer(libp2pNode, h, appConfig.FederateAddr)
	if err != nil {
		panic(err)
	}
	// make sure we're connected to federate addr always
	go func() {
		// TODO should check the peer id from the federate addr
		orID := "QmYPJrFYfohS7zcT6aVjbGfWxMfm5GhN6qChVrUdLDjaEH"
		for {
			time.Sleep(5 * time.Second)
			conns := h.Network().Conns()

			foundOR := false
			for _, conn := range conns {
				if conn.RemotePeer().String() == orID {
					foundOR = true
				}
			}

			if !foundOR {
				log.Println("Lost connection to OR for some reason")
				log.Println("reconnecting")
				err := p2p.ConnectToPeer(libp2pNode, h, appConfig.FederateAddr)
				if err != nil {
					panic(err)
				}
				log.Println("reconnected")
			}
		}
	}()

	log.Fatal(http.StartServer(appConfig, libp2pNode))
}
