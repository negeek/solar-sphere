// Command keygen prints a fresh SIGNING_KEY/VERIFICATION_KEY pair for
// signing and verifying solar-sphere access keys.
//
//	go run ./solar-spectrum/cmd/keygen
package main

import (
	"fmt"

	"github.com/negeek/solar-sphere/solar-spectrum/accesskey"
)

func main() {
	verificationKey, signingKey, err := accesskey.GenerateKeyPair()
	if err != nil {
		panic(err)
	}
	fmt.Println("SIGNING_KEY=" + signingKey)
	fmt.Println("VERIFICATION_KEY=" + verificationKey)
}
