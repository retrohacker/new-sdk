package main

import (
	"./cmd"
	"./logger"
	"archive/tar"
	"compress/gzip"
	"crypto/sha512"
	"github.com/alecaivazis/survey"
	"github.com/gosuri/uiprogress"
	"github.com/kardianos/osext"
	"io"
	"os"
)

func main() {
	// First a quick heuristic to verify we have expanded our files into the CWD
	if !checkDir() {
		expandFiles()
	}

	err := cmd.RootCmd.Execute()
	if err != nil {
		logger.Fatal(err)
	}
}

func checkDir() bool {
	var (
		err   error
		cwd   string
		dir   *os.File
		files []os.FileInfo
	)

	cwd, err = os.Getwd()
	if err != nil {
		logger.Error(err)
		logger.Fatal("Unable to find current working directory")
	}

	dir, err = os.Open(cwd)
	if err != nil {
		logger.Error(err)
		logger.Fatal("Unable to open working directory: ", cwd)
	}

	files, err = dir.Readdir(100)
	if err != nil && err != io.EOF {
		logger.Error(err)
		logger.Fatal("Unable to read working directory: ", cwd)
	}

	// We use a quick heuristic to see if we have already expanded the SDK files
	// If we see each of these files in the directory and they are of the correct
	// type, we assume the sdk has expanded
	expectedFiles := []string{
		"docker-compose.yml",
		"bridge",
		"cli",
		"vpn",
		"complex",
		"mongodb",
		"share",
	}

	// true = dir, false = file
	expectedTypes := []bool{
		false, // docker-compose.yml
		true,  // bridge
		true,  // cli
		true,  // vpn
		true,  // complex
		true,  // mongodb
		true,  // share
	}

	expectedSeen := []bool{
		false, // docker-compose.yml
		false, // bridge
		false, // cli
		false, // vpn
		false, // complex
		false, // mongodb
		false, // share
	}

	for _, file := range files {
		for index, name := range expectedFiles {
			if name == file.Name() && expectedTypes[index] == file.IsDir() {
				expectedSeen[index] = true
			}
		}
	}

	for _, seen := range expectedSeen {
		if !seen {
			return false
		}
	}

	return true
}

func expandFiles() {
	// We expect a clean working directory to expand into
	expectClean()

	// First prompt to see if the user wants to expand here
	expandPrompt()

	// We now know we have a clean working directory and permission to expand,
	// let's do this!
	unpack()
}

func expectClean() {
	var (
		cwd          string
		err          error
		dir          *os.File
		files        []os.FileInfo
		binaryFolder string
	)

	cwd, err = os.Getwd()
	if err != nil {
		logger.Error(err)
		logger.Fatal("Unable to find current working directory")
	}

	dir, err = os.Open(cwd)
	if err != nil {
		logger.Error(err)
		logger.Fatal("Unable to open working directory: ", cwd)
	}

	files, err = dir.Readdir(100)
	if err != nil && err != io.EOF {
		logger.Error(err)
		logger.Fatal("Unable to read working directory: ", cwd)
	}

	// Get a reference to the executable and open it
	binaryFolder, err = osext.ExecutableFolder()
	if err != nil {
		logger.Error("Unable to resolve path to binary file!")
		logger.Fatal(err)
	}

	// We don't expect any files in the cwd unless the binary is here, then we
	// expect 1. We do this incase they put the binary in their path.
	expectedLen := 0
	if binaryFolder == cwd {
		expectedLen++
	}

	if len(files) > expectedLen {
		logger.Error("This tool expects a clean working directory to expand into.")
		logger.Info()
		logger.Info("Please create a new directory and rerun the sdk from there.")
		logger.Info(
			"If you have already expanded the sdk into a folder, please change back")
		logger.Info("to that directory and re-run the sdk.")
		logger.Info()
		logger.Fatal("Refusing to continue")
	}
}

func expandPrompt() {
	var qs = []*survey.Question{
		{
			Name: "expand",
			Prompt: &survey.Choice{
				Message: "Would you like to expand here?",
				Choices: []string{"yes", "no"},
				Default: "yes",
			},
		},
	}

	cwd, err := os.Getwd()
	if err != nil {
		logger.Error(err)
		logger.Fatal("Unable to find current working directory")
	}

	logger.Attn("We noticed you haven't expanded the binary into this directory")
	logger.Info()
	msg :=
		`We can automatically expand all of the SDK files into this directory for you.

The directory everything will be put into is:
`
	logger.Info(msg, cwd)
	logger.Info()

	answers, err := survey.Ask(qs)
	if err != nil {
		logger.Error(err)
		logger.Fatal("Unable to parse user input")
	}

	if answers["expand"] == "no" {
		logger.Fatal("This tool requires the SDK files. Refusing to continue.")
	}
}

func unpack() {
	// Our marker string indicates when we have reached the beginning of the tar
	// archive in our file
	const markerString = "StorjSDKBinaryPackedDataMarker"

	var (
		index       int64
		headers     int64
		hcount      int64
		binaryPath  string
		f           *os.File
		fstat       os.FileInfo
		err         error
		mBytes      []byte
		marker      []byte
		bytes       []byte
		markerIndex int
		tarReader   *tar.Reader
		header      *tar.Header
		gzipReader  *gzip.Reader
	)

	// Get a reference to the executable and open it
	binaryPath, err = osext.Executable()

	if err != nil {
		logger.Error("Unable to resolve path to binary file!")
		logger.Fatal(err)
	}

	f, err = os.OpenFile(binaryPath, os.O_RDONLY, 0755)

	if err != nil {
		logger.Error("Unable to open executable!")
		logger.Error("Resolved executable to: %s", binaryPath)
		logger.Fatal(err)
	}

	fstat, err = f.Stat()

	if err != nil {
		logger.Error("Unable to stat executable")
		logger.Fatal(err)
	}

	// Create a progress bar
	uiprogress.Start()                 // start rendering
	progress := uiprogress.AddBar(100) // Add a new bar
	stage := "  Seeking"
	progress.AppendCompleted()
	progress.PrependFunc(func(b *uiprogress.Bar) string {
		return stage + " "
	})

	// Begin looking for our marker
	mBytes = make([]byte, len(markerString))
	copy(mBytes[:], markerString)
	mHash := sha512.Sum512(mBytes)
	marker = make([]byte, len(mHash))
	copy(marker[:], mHash[:])
	markerIndex = 0

	_, err = f.Read(bytes)
	index = 0
	bytes = make([]byte, 1)
Reader:
	for err == nil {
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

		// Do we need to increment the progress bar?
		index++
		if index%(fstat.Size()/100) == 0 {
			progress.Incr()
		}
		_, err = f.Read(bytes)
	}

	if err != io.EOF && err != nil {
		logger.Info("Failed to parse executable!")
		logger.Fatal(err)
	}

	if markerIndex != len(marker) {
		logger.Fatal("Did not find marker!")
	}

	// File has been advanced to the beginning of the archive so we can begin
	// reading the .tar.gz and we can count the headers
	stage = "Headers"
	headers = 0

	gzipReader, err = gzip.NewReader(f)
	if err != nil {
		logger.Fatal("Unable to parse gzip:", err)
	}

	tarReader = tar.NewReader(gzipReader)
	_, err = tarReader.Next()
	for ; err == nil; _, err = tarReader.Next() {
		headers += 1
	}

	if err != io.EOF {
		logger.Fatal("Encountered error while unpacking:", err)
	}

	// Go back to the beginning of the tar.gz
	_, err = f.Seek(index, 0)

	if err != nil {
		logger.Error("Unable to seek in file")
		logger.Fatal(err)
	}

	stage = "Unpacking"
	gzipReader, err = gzip.NewReader(f)
	if err != nil {
		logger.Fatal("Unable to parse gzip:", err)
	}

	tarReader = tar.NewReader(gzipReader)
	header, err = tarReader.Next()
	for ; err == nil; header, err = tarReader.Next() {
		if header.FileInfo().IsDir() {
			err := os.MkdirAll(header.Name, 0777)
			if err != nil {
				logger.Error("Unable to create folder: ", header.Name)
				logger.Fatal(err)
			}
			continue
		}

		tarFile, err := os.OpenFile(header.Name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, header.FileInfo().Mode())
		if err != nil {
			logger.Error("Failed to create file")
			logger.Fatal(err)
		}

		_, err = io.Copy(tarFile, tarReader)
		if err != nil {
			logger.Error("Unable to write file to disk")
			logger.Fatal(err)
		}

		// Update progress bar
		hcount += 1
		if hcount%(headers/(100-index/(fstat.Size()/100))) == 0 {
			progress.Incr()
		}
	}

	if err != io.EOF {
		logger.Fatal("Encountered error while unpacking:", err)
	}
}
