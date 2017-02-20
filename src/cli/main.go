package main

import (
  "github.com/kardianos/osext"
  "crypto/sha512"
	"archive/tar"
	"compress/gzip"
  "log"
	"os"
  "io"
)

func main() {
  defer os.Exit(0)
  unpack()
}

func unpack() {
  // Our marker string indicates when we have reached the beginning of the tar
  // archive in our file
	const markerString = "StorjSDKBinaryPackedDataMarker"
  var (
    binaryPath string
    f *os.File
    err error
		mBytes     []byte
		marker     []byte
		bytes      []byte
    markerIndex int
		tarReader  *tar.Reader
		header     *tar.Header
		gzipReader *gzip.Reader
  )

  // Get a reference to the executable and open it
  binaryPath, err = osext.Executable()

  if err != nil {
    log.Println("Unable to resolve path to binary file!")
    log.Fatal(err)
  }

  f, err = os.OpenFile(binaryPath, os.O_RDONLY, 0755)

  if err != nil {
    log.Println("Unable to open executable!")
    log.Printf("Resolved executable to: %s", binaryPath)
    log.Fatal(err)
  }

  // Begin looking for our marker
	mBytes = make([]byte, len(markerString))
	copy(mBytes[:], markerString)
  mHash := sha512.Sum512(mBytes)
  marker = make([]byte, len(mHash))
  copy(marker[:], mHash[:])
  markerIndex = 0

  index := 1
  _, err = f.Read(bytes)
  bytes = make([]byte, 1)
  Reader: for err == nil {
    // Iterate through each byte of the file looking for our marker
    for _, value := range bytes {
      if value != marker[markerIndex] {
        markerIndex = -1
      }
      markerIndex += 1
      if markerIndex == len(marker) {
        break Reader
      }
    }
    index += 1
    if index % 1000 == 0 {
      log.Println("Read %v bytes", index)
    }
    _, err = f.Read(bytes)
  }

  if err != io.EOF && err != nil {
    log.Println("Failed to parse executable!")
    log.Fatal(err)
  }

  if markerIndex != len(marker) {
    log.Fatal("Did not find marker!")
  }

  log.Println("Found marker at position: ", index)

	gzipReader, err = gzip.NewReader(f)
	if err != nil {
		log.Fatal("Unable to parse gzip:", err)
		os.Exit(1)
	}

	tarReader = tar.NewReader(gzipReader)
	header, err = tarReader.Next()
	for ; err == nil; header, err = tarReader.Next() {
		log.Println("Header:", header.Name)
	}

	if err != io.EOF {
		log.Fatal("Encountered error while unpacking:", err)
		os.Exit(1)
	}

	os.Exit(0)
}
