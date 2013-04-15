package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"log"
    "io"
    "net/http"
	"os"
	"runtime"
    "encoding/json"
)

type Config struct {
    Target  string
    Clients int
    Sleep string
}

func ParseConfig(filename string) (config *Config) {
	config = new(Config)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("unable to open file: " + err.Error())
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	fsize := fileInfo.Size()
	rawBytes := make([]byte, fsize)
	_, err = file.Read(rawBytes)
	if err != nil {
		log.Fatal("unable to read file: " + err.Error())
	}
	if err = json.Unmarshal(rawBytes, config); err != nil {
		log.Fatal("Unable to parse file: " + err.Error())
	}

	return config
}

const (
	VERSION = "0.0.1"
)

func pongServer(ws *websocket.Conn) {
    io.Copy(ws, ws)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC! %s\n", r)
		}
	}()
	configFile := flag.String("config", "config.json", "Config file")
	flag.Parse()

	config := ParseConfig(*configFile)
	if config == nil {
		log.Fatal("No config")
		return
	}

	// This is an odd value full of voodoo.
	// The docs say that this should match the number of CPUs, only if you
	// set it to 1, go appears to not actually spawn any threads. (None of
	// the poundSock() calls are made.) If you give it something too excessive,
	// the scheduler blows chunks. 8 per CPU, while fairly arbitrary, seems
	// to provide the greatest stability.
	//
	// Go is a fun toy, but this is why you don't build hospitals out of lego.
	runtime.GOMAXPROCS(runtime.NumCPU() * 8)

    http.Handle("/ws", websocket.Handler(pongServer))
    log.Printf("Starting server")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        panic (err.Error())
    }
}
