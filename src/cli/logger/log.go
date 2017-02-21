package logger

import (
	"fmt"
	"github.com/fatih/color"
	"os"
)

func Attn(v ...interface{}) {
	color.Set(color.FgBlue)
	fmt.Println(v...)
	color.Unset()
}

func Info(v ...interface{}) {
	fmt.Println(v...)
}

func Warn(v ...interface{}) {
	color.Set(color.FgYellow)
	fmt.Print("Warn: ")
	fmt.Println(v...)
	color.Unset()
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
