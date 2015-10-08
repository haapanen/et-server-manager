package main
import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type Configuration struct {
	Servers []Server `json:"servers"`
	ExecutablePath string `json:"executablePath"`
	ScreenPath string `json:"screenPath"`
	PgrepPath string `json:"pgrepPath"`
	KillPath string `json:"killPath"`
	SuPath string `json:"suPath"`
}

//
// Reads a configuration from fileName and returns it
func ReadConfig(fileName string) (config *Configuration, err error) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Could not open config.json. Creating a new one.")
		config = createDefaultConfig()
		err = SaveConfig(fileName, config)
		return config, err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("Could not read config from config.json: %s\n", err)
		return config, err
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		fmt.Println("Could not parse json configuration: %s\n", err)
		return config, err
	}

	return config, nil
}

//
// Saves the config to a file
func SaveConfig(fileName string, config *Configuration) (error) {
	bytes, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fileName, bytes, 0644)
	return err
}

//
// Adds a default server to the configuration
func AddServer(config *Configuration) {
	server := Server{}

	server.Name = "etserver1"
	server.Ip = "127.0.0.1"
	server.Port = 27960
	server.Mod = "etjump"
	server.BasePath = "/home/server/et/"
	server.HomePath = "/home/server/et/"
	server.Configs = []string{"server.cfg"}
	server.User = "root"
	server.Running = false
	server.Pid = -1

	config.Servers = append(config.Servers, server)
}

//
// Creates a default configuration
func createDefaultConfig() (config *Configuration) {
	AddServer(config)
	config.ExecutablePath = "/path/to/etded"
	config.ScreenPath = "/usr/bin/screen"
	config.PgrepPath = "/usr/bin/pgrep"
	config.KillPath = "/bin/kill"
	config.SuPath = "/bin/su"

	return config
}