package update

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/open-services/bolivar/cli"
)

// HTTPUpdater receives root hash updates via polling a HTTP endpoint
type HTTPUpdater struct {
	Config cli.Config
}

type httpRes struct {
	Hash           string `json:"hash"`
	CumulativeSize int    `json:"cumulativesize"`
	Blocks         int    `json:"blocks"`
}

// helper function to receive json from endpoint
func getJSON(userAgent, url string, target interface{}) error {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	r.Header.Set("User-Agent", userAgent)
	resp, err := myClient.Do(r)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	return json.NewDecoder(resp.Body).Decode(target)
}

// NewHTTPUpdater returns a new http updater...
func NewHTTPUpdater(config cli.Config) *HTTPUpdater {
	hu := new(HTTPUpdater)
	hu.Config = config
	return hu
}

var myClient = &http.Client{Timeout: 10 * time.Second}

// Update receives latest root hash from open-registry
// TODO write to repo path so the next time we start it, we have a hash
// to go by initially
func (hu *HTTPUpdater) Update() string {
	endpoint := hu.Config.HTTPEndpoint
	isVerbose := hu.Config.Verbose

	if isVerbose {
		log.Println("Fetching latest root hash from " + endpoint)
	}

	userAgent := "bolivar/0.1.0"

	res := new(httpRes)
	err := getJSON(userAgent, endpoint, res)

	if err != nil {
		log.Println("Seems there are issues contacting to " + endpoint + " right now")
		if isVerbose {
			log.Println(err)
		}
	}

	if isVerbose {
		log.Println("Fetched, latest hash is " + res.Hash)
	}

	return res.Hash
}
