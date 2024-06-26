package vectorize

import (
	"crypto/rand"
	"math/big"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/TheHackerDev/cartograph/internal/analyzer/vectorize/bagofwords"
)

// serverheader.go contains the code to vectorize HTTP server headers.

// serverHeadersMap is a map of server headers to their index in the vector.
var serverHeadersMap map[string]int

// init initializes the serverHeadersMap, and runs before the main function.
func init() {
	serverHeadersMap = make(map[string]int)
	for _, header := range bagofwords.ServerHeaders {
		serverHeadersMap[header] = 1
	}
}

// ServerHeader returns a vectorized representation of the server headers.
func ServerHeader(serverHeaderValues []string) []float32 {
	vector := make([]float32, len(bagofwords.ServerHeaders))

	for _, serverHeader := range serverHeaderValues {
		if _, ok := serverHeadersMap[strings.ToLower(serverHeader)]; ok {
			vector[serverHeadersMap[strings.ToLower(serverHeader)]] = 1
		}
	}

	return vector
}

// GenerateServerHeaderVector generates a vector of server headers for testing purposes.
// The vector is generated by randomly selecting one of the server headers from the list of server headers.
func GenerateServerHeaderVector() []float32 {
	vector := make([]float32, len(bagofwords.ServerHeaders))

	// Randomly select a server header to add to the vector
	index, randErrBig := rand.Int(rand.Reader, big.NewInt(int64(len(bagofwords.ServerHeaders))))
	if randErrBig != nil {
		log.Fatal("Error generating random number for server header vector generation")
	}
	vector[index.Int64()] = 1

	return vector
}
