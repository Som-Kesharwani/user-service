package helper

import (
	"flag"
	"github.com/Som-Kesharwani/shared-service/logger"
	"os"
)

var Secretary []byte

func init() {
	// Load the private key from the file specified by the command line flag
	keyPath := flag.String("Secretary", "F:\\Golang\\user-service\\static\\private_key.pem", "path to the private key PEM file")
	flag.Parse()

	// Read the private key filenil

	var err error
	Secretary, err = os.ReadFile(*keyPath)
	if err != nil {
		logger.Error.Printf("Failed to read private key: %v", err)
		panic(err)
	}

}
