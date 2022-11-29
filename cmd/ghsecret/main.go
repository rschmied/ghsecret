package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"log"

	"golang.org/x/crypto/nacl/box"
)

const (
	GH_KEY    string = "GH_KEY"
	GH_KEY_ID string = "GH_KEY_ID"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "dev"

	// goreleaser can also pass the specific commit if you want
	// commit string = ""
)

type Secret struct {
	KeyID string `json:"key_id"`
	Value string `json:"encrypted_value"`
}

func main() {
	flag.Usage = func() {
		cmd := filepath.Base(os.Args[0])
		fmt.Printf("%s [-key ENV_NAME][-key-id ENV_NAME] [value]\n\n", cmd)
		fmt.Printf("Version: %s\n\n", version)
		fmt.Println("Encrypts the given value or stdin if no value argument is")
		fmt.Println("provided using the key stored in GH_KEY with the key ID")
		fmt.Println("GH_KEY_ID. The result is a JSON object that can be consumed")
		fmt.Println("by the Github secrets API via the e.g. gh CLI client.")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
	}

	ghKeyEnv := flag.String("key", GH_KEY, "Name of the env var")
	ghKeyIDenv := flag.String("key-id", GH_KEY_ID, "Name of the env var")
	flag.Parse()

	base64key, ok := os.LookupEnv(*ghKeyEnv)
	if !ok {
		log.Printf("Required env var \"%s\" with key data not found\n", *ghKeyEnv)
		os.Exit(1)
	}

	keyID, ok := os.LookupEnv(*ghKeyIDenv)
	if !ok {
		log.Printf("Required env var \"%s\" with key ID not found\n", *ghKeyIDenv)
		os.Exit(1)
	}

	gh_key, err := base64.StdEncoding.DecodeString(base64key)
	if err != nil {
		log.Println("Key did not decode (base64)")
		os.Exit(1)
	}

	var value []byte
	switch flag.NArg() {
	case 0:
		value, err = io.ReadAll(io.Reader(os.Stdin))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Reading standard input:", err)
		}
	case 1:
		var ok bool
		valueString, ok := os.LookupEnv(flag.Arg(0))
		if !ok {
			log.Printf("Provided env var \"%s\" with value not found\n", flag.Arg(0))
			os.Exit(1)
		}
		value = []byte(valueString)
	default:
		fmt.Fprintln(os.Stderr, "Either no or exactly one argument is required")
		os.Exit(1)
	}

	var secretKey [32]byte
	copy(secretKey[:], gh_key)

	encMsg := []byte{}
	encMsg, err = box.SealAnonymous(encMsg, value, &secretKey, rand.Reader)
	if err != nil {
		log.Printf("Encrypt didn't work: %s\n", err)
		os.Exit(1)
	}
	// print the JSON object to stdout, the encrypted secret is base64 encoded
	secret := Secret{keyID, base64.StdEncoding.EncodeToString(encMsg)}
	data, err := json.Marshal(secret)
	if err != nil {
		log.Printf("Couldn't encode secret: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}
