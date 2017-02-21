package main

import (
	"crypto/sha512"
	"io"
	"log"
	"os"
)

func main() {
	defer os.Exit(0)

	const markerString = "StorjSDKBinaryPackedDataMarker"

	var (
		tarball    *os.File
		executable *os.File
		mBytes     []byte
		marker     []byte
		err        error
		bytes      []byte
		nr         int
		nw         int
	)

	tarball, err = os.OpenFile("./sdk.tar.gz", os.O_RDONLY, 0755)
	if err != nil {
		log.Print("Error: ", err)
		log.Fatal("Unable to open ./sdk.tar.gz")
	}

	executable, err = os.OpenFile("./sdk", os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		log.Print("Error: ", err)
		log.Fatal("Unable to open ./sdk")
	}

	defer tarball.Close()
	defer executable.Close()

	// Create a marker in the file so we can find the beginning of the tarball
	// when unpacking
	mBytes = make([]byte, len(markerString))
	copy(mBytes[:], markerString)
	mHash := sha512.Sum512(mBytes)
	marker = make([]byte, len(mHash))
	copy(marker[:], mHash[:])

	log.Printf("Using marker: %v", marker)
	nw, err = executable.Write(marker)
	if err != nil {
		log.Print("Error: ", err)
		log.Fatal("Failed to write marker to file")
	}

	bytes = make([]byte, 100)
	nr, err = tarball.Read(bytes)
	for err == nil {
		nw, err = executable.Write(bytes[0:nr])
		if err != nil {
			break
		}
		nr, err = tarball.Read(bytes)
	}

	if nr == 0 && err != io.EOF {
		log.Print("Error: ", err)
		log.Fatal("Failed to read entire tarball")
	}

	if nw == 0 && err != io.EOF {
		log.Print("Error: ", err)
		log.Fatal("Failed to write to executable")
	}
}
