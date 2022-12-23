package log

import (
	"io"
	"log"
	"os"
)

var (
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Fatal   *log.Logger
	LogFile *os.File
)

func init() {
	Debug = log.New(os.Stdout,
		"DEBUG: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(os.Stdout,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(os.Stdout,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(os.Stderr,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Fatal = log.New(os.Stderr,
		"FATAL: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func SetLogFile(f string) {
	of, err := os.OpenFile(f,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	LogFile = of
	ioo := io.MultiWriter(os.Stdout, of)
	ioe := io.MultiWriter(os.Stderr, of)
	Debug.SetOutput(ioo)
	Info.SetOutput(ioo)
	Warning.SetOutput(ioe)
	Error.SetOutput(ioe)
	Fatal.SetOutput(ioe)
}

func SetLogLevel(l string) {
	switch l {
	case "DEBUG":
	case "INFO":
	case "WARNING":
	case "ERROR":
	case "FATAL":
	}
}

func Close() {
	if LogFile != nil {
		LogFile.Close()
	}
}
