package logger

import (
  "os"
  "fmt"
  "github.com/fatih/color"
)

func Info(v ...interface{}) {
  fmt.Println(v...)
}

func Warn(v ...interface{}) {
}

func Error(v ...interface{}) {
  color.Set(color.FgRed)
  fmt.Print("Error: ")
  fmt.Println(v...)
  color.Unset()
}

func Fatal(v ...interface{}) {
  Error(v...)
  os.Exit(-1)
}
