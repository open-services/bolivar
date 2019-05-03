package update

import (
	"log"

	"github.com/open-services/bolivar/cli"
)

// Updater is a mechanism for receiving a new root hash
type Updater interface {
	// Start(interval time.Duration)
	Update() string
}

// checks if a string in a list of strings exists
// (how is this not in the standard library???)
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// valid update types
var updateTypes = []string{"http", "dns", "ipns", "pubsub"}

// NewUpdater returns a updater based on the `update-type` in config
func NewUpdater(config cli.Config) Updater {
	updateType := config.UpdateType
	if stringInSlice(updateType, updateTypes) {
		// Currently only supports http
		if updateType != "http" {
			log.Fatal("Update type " + updateType + " is not yet supported")
		}
	} else {
		log.Fatal("Update type " + updateType + " is not supported")
	}

	updater := NewHTTPUpdater(config)
	return updater
}
