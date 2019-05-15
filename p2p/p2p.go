package p2p

import (
	"context"
	"fmt"
	"strconv"
	"time"

	humanize "github.com/dustin/go-humanize"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	golog "github.com/ipfs/go-log"
	libp2p "github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	ipnet "github.com/libp2p/go-libp2p-interface-pnet"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	metrics "github.com/libp2p/go-libp2p-metrics"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	pnet "github.com/libp2p/go-libp2p-pnet"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/open-services/bolivar/cli"
)

// CurrentStats is the latest received metrics for current/total bandwidth
var CurrentStats *metrics.Stats

func setupConnMgr() *connmgr.BasicConnMgr {
	cmgr := connmgr.NewConnManager(10, 50, 1*time.Second)
	go func() {
		nctx, cancel := context.WithCancel(context.Background())
		for {
			time.Sleep(10 * time.Second)
			if nctx == nil {
				nctx, cancel = context.WithCancel(context.Background())
			}
			cmgr.TrimOpenConns(nctx)
			cancel()
			cancel = nil
			nctx = nil
		}
	}()
	return cmgr
}

// SetupLibp2p ...
func SetupLibp2p(
	ctx context.Context,
	hostKey crypto.PrivKey,
	secret []byte,
	listenAddrs []multiaddr.Multiaddr,
) (host.Host, *dht.IpfsDHT, error) {

	var prot ipnet.Protector
	var err error

	// Create protector if we have a secret.
	if len(secret) > 0 {
		var key [32]byte
		copy(key[:], secret)
		prot, err = pnet.NewV1ProtectorFromBytes(&key)
		if err != nil {
			return nil, nil, err
		}
	}

	rep := metrics.NewBandwidthCounter()

	h, err := libp2p.New(
		ctx,
		libp2p.Identity(hostKey),
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.PrivateNetwork(prot),
		libp2p.NATPortMap(),
		libp2p.ConnectionManager(setupConnMgr()),
		libp2p.BandwidthReporter(rep),
	)

	if err != nil {
		return nil, nil, err
	}

	firstResult := false

	// reporting
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			// bandwidth
			stats := rep.GetBandwidthTotals()
			CurrentStats = &stats
			firstResult = true
		}
	}()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			// peers
			conns := h.Network().Conns()
			numPeers := strconv.Itoa(len(conns))

			if firstResult {
				stats := CurrentStats

				// spew.Dump(stats)
				bwMessage := "%s|%s [%s/s|%s/s] (Peers: %s)"
				fmt.Println(fmt.Sprintf(bwMessage,
					humanize.Bytes(uint64(stats.TotalIn)),
					humanize.Bytes(uint64(stats.TotalOut)),
					humanize.Bytes(uint64(stats.RateIn)),
					humanize.Bytes(uint64(stats.RateOut)),
					numPeers,
				))
			}
		}
	}()

	idht, err := dht.New(ctx, h)
	if err != nil {
		nerr := h.Close()
		if nerr != nil {
			return nil, nil, nerr
		}
		return nil, nil, err
	}

	rHost := routedhost.Wrap(h, idht)
	return rHost, idht, nil
}

type discoveryHandler struct {
	host host.Host
}

const discoveryConnTimeout = time.Second * 30

func (dh *discoveryHandler) HandlePeerFound(p peerstore.PeerInfo) {
	// fmt.Println("found local peer")
	// // log.Warning("trying peer info: ", p)
	// spew.Dump(p.Addrs)
	ctx, cancel := context.WithTimeout(context.Background(), discoveryConnTimeout)
	defer cancel()
	// important connection!
	if err := dh.host.Connect(ctx, p); err != nil {
		fmt.Println("Failed to connect to peer found by discovery: ", err)
	}
	dh.host.ConnManager().Protect(p.ID, "local")
	fmt.Println("Think we protected a bootstrap peer from being killed too")
}

// StartLibp2p ...
// TODO needs to have a connmgr as we're connecting to all peers currently
// TODO also include bandwidth reporter for nicer ui
func StartLibp2p(config cli.Config) (*ipfslite.Peer, host.Host) {
	fmt.Println("Starting libp2p host")
	ctx := context.Background()

	err := golog.SetLogLevel("*", "notice")
	// err := golog.SetLogLevel("*", "info")
	if err != nil {
		panic(err)
	}

	ds, err := ipfslite.BadgerDatastore(config.RepoPath)
	if err != nil {
		panic(err)
	}
	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		panic(err)
	}

	// TODO configurable
	listen, _ := multiaddr.NewMultiaddr(config.ListenAddr)

	h, dht, err := SetupLibp2p(
		ctx,
		priv,
		nil,
		[]multiaddr.Multiaddr{listen},
	)

	if err != nil {
		panic(err)
	}

	srv, err := discovery.NewMdnsService(context.Background(), h, 5*time.Second, discovery.ServiceTag)

	if err != nil {
		panic(err)
	}

	n := &discoveryHandler{h}

	srv.RegisterNotifee(n)

	lite, err := ipfslite.New(ctx, ds, h, dht, nil)
	if err != nil {
		panic(err)
	}
	return lite, h
}

// ConnectToPeer bootstraps with a peer to keep the connection open for as long as possible
func ConnectToPeer(node *ipfslite.Peer, h host.Host, ma string) error {
	fmt.Println("Connecting to open-registry.dev libp2p node...")
	openRegistryAddr, err := multiaddr.NewMultiaddr(ma)
	if err != nil {
		return err
	}
	orPeerInfo, err := peerstore.InfoFromP2pAddr(openRegistryAddr)
	if err != nil {
		return err
	}
	peers := []peerstore.PeerInfo{}
	peers = append(peers, *orPeerInfo)
	// mark as important so connmgr doesn't kill connection
	node.Bootstrap(peers)
	fmt.Println("Connected")
	h.ConnManager().Protect(orPeerInfo.ID, "bootstrap")
	fmt.Println("Think we protected a bootstrap peer from being killed too")
	return nil
}
