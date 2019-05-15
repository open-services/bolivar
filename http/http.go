package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	gohttp "net/http"
	"time"

	"github.com/gorilla/mux"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	cid "github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-unixfs/hamt"
	"github.com/open-services/bolivar/cli"
	"github.com/open-services/bolivar/p2p"
)

var libp2pNode *ipfslite.Peer
var currentRootHash = ""

// FetchMetadataViaHTTP fetches metadata for a package from a centralized registry
func FetchMetadataViaHTTP(w io.Writer, packageName string) {
	url := "https://npm.open-registry.dev/" + packageName
	fmt.Println("external metadata fetch " + url)
	res, err := gohttp.Get(url)
	if err != nil {
		panic(err)
	}
	// TODO needs to rewrite to localhost
	_, err = io.Copy(w, res.Body)
	if err != nil {
		panic(err)
	}
}

// FetchTarballViaHTTP fetches tarball for a package + version from a centralized registry
func FetchTarballViaHTTP(w io.Writer, packageName string, tarball string) {
	url := "https://npm.open-registry.dev/" + packageName + "/-/" + tarball
	fmt.Println("external tarball fetch " + url)
	res, err := gohttp.Get(url)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(w, res.Body)
	if err != nil {
		panic(err)
	}
}

// GetNode returns data from a hash
func GetNode(node *ipfslite.Peer, hash string) format.Node {
	c, err := cid.Decode(hash)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	dnode, err := node.Get(ctx, c)

	if err != nil {
		panic(err)
	}
	return dnode
}

// FindLink takes a link name and a hash, returns the hash of that link or ""
func FindLink(pkg string, hash string) string {
	dnode := GetNode(libp2pNode, hash)

	// TODO should only do this once from each root hash
	shard, err := hamt.NewHamtFromDag(libp2pNode.DAGService, dnode)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	link, err := shard.Find(ctx, pkg)
	if err != nil {
		return ""
	}
	return link.Cid.String()
}

// MetricsHandler returns total and current bandwidth
func MetricsHandler(w gohttp.ResponseWriter, r *gohttp.Request) {
	err := json.NewEncoder(w).Encode(p2p.CurrentStats)
	if err != nil {
		panic(err)
	}
}

// MetadataHandler returns metadata about a package
func MetadataHandler(w gohttp.ResponseWriter, r *gohttp.Request) {
	vars := mux.Vars(r)
	rootHash := currentRootHash
	if len(vars["scope"]) > 0 {
		rootHash = FindLink(vars["scope"], rootHash)
		if rootHash == "" {
			// fetch via http as we're missing the scope + package fully
			fmt.Println("Couldnt find that scope")
			// npm client uses %2f instead of /
			FetchMetadataViaHTTP(w, vars["scope"]+"%2f"+vars["package"])
			return
		}
	}
	packageName := vars["package"]

	packageHash := FindLink(packageName, rootHash)
	if packageHash == "" {
		// fetch via http as we're missing the package fully
		FetchMetadataViaHTTP(w, packageName)
		return
	}

	// fmt.Println("finding metadata for " + packageHash)
	metadataHash := FindLink("metadata.json", packageHash)
	if metadataHash == "" {
		// fetch via http as we're missing the metadata
		FetchMetadataViaHTTP(w, packageName)
		return
	}

	c, err := cid.Decode(metadataHash)
	if err != nil {
		panic(err)
	}

	res, err := libp2pNode.GetFile(context.Background(), c)
	if err != nil {
		panic(err)
	}

	buf, err := ioutil.ReadAll(res)
	if err != nil {
		panic(err)
	}

	newAddress := []byte("http://" + r.Host)

	result := bytes.Replace(buf,
		[]byte("https://npm.open-registry.dev"),
		newAddress,
		-1)

	_, err = w.Write(result)
	if err != nil {
		panic(err)
	}
}

// TarballHandler returns the tarball for a package
// TODO missing fetch from http when tarball is missing
func TarballHandler(w gohttp.ResponseWriter, r *gohttp.Request) {
	vars := mux.Vars(r)
	packageName := vars["package"]
	tarball := vars["tarball"]

	rootHash := currentRootHash
	if len(vars["scope"]) > 0 {
		rootHash = FindLink(vars["scope"], rootHash)
	}

	packageHash := FindLink(packageName, rootHash)
	if packageHash == "" {
		panic("Could not find package " + packageName)
	}

	tarballHash := FindLink(tarball, packageHash)

	if tarballHash == "" {
		// fetch via http as we're missing the metadata
		FetchTarballViaHTTP(w, packageName, tarball)
		return
	}

	c, err := cid.Decode(tarballHash)
	if err != nil {
		panic(err)
	}
	res, err := libp2pNode.GetFile(context.Background(), c)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(w, res)
	if err != nil {
		panic(err)
	}
}

// StartServer starts the local http server
func StartServer(config cli.Config, newLibp2pNode *ipfslite.Peer) error {
	libp2pNode = newLibp2pNode

	r := mux.NewRouter()
	r.HandleFunc("/_api/metrics", MetricsHandler)
	r.HandleFunc("/{package}", MetadataHandler)
	r.HandleFunc("/{scope}/{package}", MetadataHandler)
	r.HandleFunc("/{package}/-/{tarball}", TarballHandler)
	r.HandleFunc("/{scope}/{package}/-/{tarball}", TarballHandler)
	gohttp.Handle("/", r)

	address := config.HTTPAddress + ":" + config.HTTPPort

	srv := &gohttp.Server{
		Handler:      r,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("Starting server " + address + "...")
	return srv.ListenAndServe()
}

// SetRootHash sets the hash to use for metadata + tarball resolution
// TODO find a better way than globals for this
func SetRootHash(newRootHash string) {
	currentRootHash = newRootHash
}
