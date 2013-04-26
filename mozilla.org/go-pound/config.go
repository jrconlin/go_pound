package go_pound

import (
    "log"
    "os"
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

