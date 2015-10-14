package main
import (
	"os"
	"fmt"
	"strconv"
	"path/filepath"
	"io/ioutil"
	"encoding/json"
)

type Server struct {
	Name     string `json:"name"`
	Ip       string `json:"ip"`
	Port     int `json:"port"`
	Mod      string `json:"mod"`
	BasePath string `json:"basePath"`
	HomePath string `json:"homePath"`
	Configs  []string `json:"configs"`
	User     string `json:"user"`
	Running  bool `json:"running"`
	Pid      int `json:"pid"`
}

type Configuration struct {
	Servers        []Server `json:"servers"`
	ExecutablePath string `json:"executablePath"`
	ScreenPath     string `json:"screenPath"`
	PgrepPath      string `json:"pgrepPath"`
	KillPath       string `json:"killPath"`
	SuPath         string `json:"suPath"`
}

type OldServer struct {
	Name     string `json:"name"`
	Ip       string `json:"ip"`
	BasePath string `json:"basePath"`
	Running  bool `json:"running"`
	User     string `json:"user"`
	HomePath string `json:"homePath"`
	Configs  []string `json:"configs"`
	Port     string `json:"port"`
	Mod      string `json:"mod"`
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("usage: ./config-converter <old config> <new config>")
		return
	}

	var oldConfig []OldServer

	bytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	json.Unmarshal(bytes, &oldConfig)

	newConfig := Configuration{}
	newConfig.ExecutablePath = "/path/to/etded"
	newConfig.ScreenPath = "/usr/bin/screen"
	newConfig.PgrepPath = "/usr/bin/pgrep"
	newConfig.KillPath = "/bin/kill"
	newConfig.SuPath = "/bin/su"

	for _, server := range oldConfig {
		newServer := Server{}
		newServer.Name = server.Name
		if len(server.Ip) > 0 {
			newServer.Ip = server.Ip
		} else {
			newServer.Ip = "localhost"
		}
		newServer.Port, _ = strconv.Atoi(server.Port)
		newServer.Mod = server.Mod
		newServer.BasePath = filepath.Clean(server.BasePath)
		newServer.HomePath = filepath.Clean(server.HomePath)
		newServer.Configs = server.Configs
		newServer.User = server.User
		newServer.Running = false
		newServer.Pid = -1
		newConfig.Servers = append(newConfig.Servers, newServer)
	}

	newConfigInJson, _ := json.MarshalIndent(newConfig, "", "  ")
	ioutil.WriteFile(os.Args[2], newConfigInJson, 0644)
}