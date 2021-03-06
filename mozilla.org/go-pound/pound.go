package main

import (
	"code.google.com/p/go.net/websocket"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"time"
)

const (
	VERSION = "0.0.1"
)

// config cruft
type Config struct {
	Target  string
	Clients int
	Sleep   string
}

func parseConfig(filename string) (config *Config) {
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

func genToken() (string) {
	uuid := make([]byte, 16)
	n, err := rand.Read(uuid)
	if n != len(uuid) || err != nil {
		return ""
	}

	uuid[8] = 0x80
	uuid[4] = 0x40

	return hex.EncodeToString(uuid)
}

func poundSock(target string, config *Config, cmd, ctrl chan int, id int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf(".")
		}
	}()
	hostname := os.Getenv("HOST")
	if hostname == "" {
		hostname = "localhost"
	}
	targ := target // + fmt.Sprintf("#id=%d", id)
	log.Printf("INFO : (%d) Connecting from %s to %s\n", id,
		"ws://"+hostname, targ)
	//ws, err := websocket.Dial(targ, "push-notification", targ)
	ws, err := websocket.Dial(targ, "", targ)
	err = ws.SetDeadline(time.Now().Add(time.Second * 30))
	if err != nil {
		log.Printf("ERROR: (%d) Unable to open websocket: %s\n",
			id, err.Error())
		cmd <- id
		return err
	}
	duration, err := time.ParseDuration(config.Sleep)
	tc := time.NewTicker(duration)
		msg := fmt.Sprintf("{\"messageType\": \"hello\", "+
			"\"uaid\": \"%s\", \"channelIDs\":[]}", genToken())
		_, err = ws.Write([]byte(msg))
	websocket.Message.Receive(ws, &msg)
	for {
		err = ws.SetDeadline(time.Now().Add(time.Second * 5))
		if err != nil {
			log.Printf("ERROR: (%d) Unable to write ping to websocket %s\n",
				id, err.Error())
			cmd <- id
			return err
		}
        ws.Write([]byte("{}"))
		// do a raw receive from the socket.
		// Note: ws.Read doesn't like pulling data.
		var msg string
		websocket.Message.Receive(ws, &msg)

		//if _, err = ws.Read(msg); err != nil {
		//
		//	log.Printf("WARN : (%d) Bad response %s\n", id, err)
		//	cmd <- id
		//	return
		//}
		select {
		case cc := <-ctrl:
			if cc == 0 {
				break
			}
		case <-tc.C:
			continue
		}
	}
	log.Printf("INFO : (%d) Shutting down...\n", id)
	return err
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC! %s\n", r)
		}
	}()
	configFile := flag.String("config", "config.json", "Config file")
	flag.Parse()

	config := parseConfig(*configFile)
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

	chans := make(map[int]chan int)
	cmd := make(chan int, config.Clients)

	// run as many clients as specified
	totalClients := config.Clients
	gov := time.NewTicker(time.Duration(time.Second * 2))
	for spawned := 0; spawned < totalClients; {
		select {
		case <-gov.C:
			log.Printf("Spawning %d\n", spawned)
			if spawned < totalClients {
				eggs := int(math.Min(1000, float64(totalClients-spawned)))
				for cli := 0; cli < eggs; cli++ {
					spawn := spawned + cli
					ctrl := make(chan int)
					chans[spawn] = ctrl
					// open a socket to the Target
					log.Printf("Spawning %d\n", spawn)

					go func(spawn int) {
						poundSock(config.Target, config, cmd, ctrl, spawn)
					}(spawn)
				}
				spawned = spawned + eggs
			}
		}

	}
	gov.Stop()
	tc := time.NewTicker(time.Duration(time.Second * 5))
	for {
		select {
		case x := <-cmd:
			log.Printf("Exiting %d \n", x)
			totalClients = runtime.NumGoroutine()
		case <-tc.C:
			log.Printf("Info: Active Clients: %d \n", totalClients)
		}
	}
}
