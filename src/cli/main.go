package main

import (
  "./cmd"
  "./logger"
  "os"
)

func main() {
  if !checkDir() {
    logger.Fatal("Directory not initalized!");
  }

  // First a quick heuristic to see if we have expanded our files into the CWD
  err := cmd.RootCmd.Execute()
  if err != nil {
    logger.Fatal(err)
  }
}

func checkDir() bool {
  var (
    err error
    cwd string
    dir *os.File
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
  if err != nil {
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
    true, // bridge
    true, // cli
    true, // vpn
    true, // complex
    true, // mongodb
    true, // share
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
