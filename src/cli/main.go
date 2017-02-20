package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	var (
		tarball    string
		err        error
		f          *os.File
		tarReader  *tar.Reader
		header     *tar.Header
		gzipReader *gzip.Reader
	)

	tarball, err = filepath.Abs("example.tar.gz")
	if err != nil {
		fmt.Println("Unable to resolve path to example.tar.gz:", err)
		os.Exit(1)
	}

	f, err = os.Open(tarball)
	if err != nil {
		fmt.Println("Unable to open file:", err)
		os.Exit(1)
	}

	gzipReader, err = gzip.NewReader(f)
	if err != nil {
		fmt.Println("Unable to parse gzip:", err)
		os.Exit(1)
	}

	tarReader = tar.NewReader(gzipReader)
	header, err = tarReader.Next()
	for ; err == nil; header, err = tarReader.Next() {
		fmt.Println("Header:", header.Name)
	}

	if err != io.EOF {
		fmt.Println("Encountered error while unpacking:", err)
		os.Exit(1)
	}

	os.Exit(0)
}
