package cli

import (
	"log"
	"path"

	"github.com/davecgh/go-spew/spew"
	homedir "github.com/mitchellh/go-homedir"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

// TODO should be using dnsaddr (currently hardcoded peerID)
var openRegistryMultiaddr = "/dns4/npm.open-registry.dev/tcp/4001/ipfs/QmdAgn46prB29nfW1fSoKJsehma7DL6C3yygjCaVKYxcFw"

// what path should our libp2p node listen at
var defaultListenAddr = "/ip4/0.0.0.0/tcp/4005"

// what path should we use to store data
// TODO should point to directory in homedir by default
var defaultRepoPath = getDefaultRepoPath()

// what address to set HTTP server to listen to
var defaultHTTPAddress = "localhost"

// what port to set HTTP server to listen to
var defaultHTTPPort = "8080"

// default http update endpoint
var defaultHTTPUpdateEndpoint = "https://npm.open-registry.dev"

// cli args
var (
	federateAddr = kingpin.Flag("federate", "Multiaddr of primary federation server to read from").Default(openRegistryMultiaddr).String()
	share        = kingpin.Flag("share", "Whetever to share downloaded packages to others").Default("true").Short('s').Bool()
	updateType   = kingpin.Flag("update-type", "How to receive updates to root hash.").Default("http").String()
	offline      = kingpin.Flag("offline", "Disable any external connections").Short('o').Bool()
	verbose      = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	listenAddr   = kingpin.Flag("listen-addr", "What address to set the libp2p node to listen at").Default(defaultListenAddr).Short('l').String()
	repoPath     = kingpin.Flag("repo-path", "What path to use for the repository").Default(defaultRepoPath).Short('r').String()
	httpAddress  = kingpin.Flag("http-address", "What address to use for HTTP server").Default(defaultHTTPAddress).Short('a').String()
	httpPort     = kingpin.Flag("http-port", "What port to use for HTTP server").Default(defaultHTTPPort).OverrideDefaultFromEnvar("PORT").Short('p').String()
	httpEndpoint = kingpin.Flag("http-endpoint", "What the endpoint for registry index is").Default(defaultHTTPUpdateEndpoint).String()
)

// Config holds all the configuration variables from parsing the CLI args and flags
type Config struct {
	Offline      bool
	Verbose      bool
	Share        bool
	FederateAddr string
	UpdateType   string
	ListenAddr   string
	RepoPath     string
	HTTPAddress  string
	HTTPPort     string
	HTTPEndpoint string
}

func getDefaultRepoPath() string {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	p := path.Join(home, ".bolivar")
	return p
}

// Init initializes the cli
func Init(version, author string) Config {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version(version).Author(author)
	kingpin.CommandLine.Help = "A federated version of Open-Registry (https://open-registry.dev)"
	kingpin.Parse()

	cfg := Config{
		FederateAddr: *federateAddr,
		Share:        *share,
		UpdateType:   *updateType,
		Offline:      *offline,
		Verbose:      *verbose,
		ListenAddr:   *listenAddr,
		RepoPath:     *repoPath,
		HTTPAddress:  *httpAddress,
		HTTPPort:     *httpPort,
		HTTPEndpoint: *httpEndpoint,
	}

	return cfg
}

// PrintConfig is a helper function to print out all config options
func PrintConfig(config Config) {
	log.Println("CLI flags and args:")
	spew.Dump(config)
}
