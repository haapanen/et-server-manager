package main
import (
	"fmt"
	"os"
	"time"
	"errors"
)

var globalConfiguration *Configuration
var configFile string = "config.json"

func main() {
	var err error
	globalConfiguration, err = ReadConfig(configFile)
	if err != nil {
		fmt.Println("Server manager startup failed: Could not read configuration. %s\n", err)
		return
	}

	argc := len(os.Args)
	if argc < 2 || argc > 3 {
		fmt.Println("usage: ./manager <loop | status>")
		fmt.Printf("usage: ./manager %s %s\n", "<start | stop | restart | status>", "<server>")
		return
	}

	if argc == 2 {
		switch os.Args[1] {
		case "add":
			AddServer(globalConfiguration)
		case "loop":
			loop()
		case "status":
			for _, server := range globalConfiguration.Servers {
				server.PrintStatus(false)
			}
		default:
			fmt.Printf("Unknown command: %s\n", os.Args[1])
		}
	} else if argc == 3 {
		server, err := findServer(os.Args[2])
		if err != nil {
			fmt.Printf("Could not find server: %s\n", err)
			return
		}
		switch os.Args[1] {
		case "start":
			server.Start()
		case "stop":
			server.Stop()
		case "restart":
			server.Restart()
		case "status":
			server.PrintStatus(true)
		default:
			fmt.Printf("Unknown command: %s\n", os.Args[1])
		}
	}
	return
}

func loop() {
	if os.Geteuid() != 0 {
		fmt.Println("You must be root to loop")
		return
	}

	config, err := ReadConfig(configFile)
	if err != nil {
		fmt.Printf("Could not read config.json: %s\n", err)
		return
	}

	fileInfo, err := os.Stat(configFile)
	if err != nil {
		fmt.Printf("Could not check previous modification time of config.json: %s\n", err)
		return
	}

	prevModTime := fileInfo.ModTime()
	for {
		time.Sleep(1 * time.Second)

		fileInfo, err = os.Stat(configFile)
		if fileInfo.ModTime() != prevModTime {
			config, err = ReadConfig(configFile)
			if err != nil {
				fmt.Printf("Could not read config.json while looping: %s\n", err)
				continue
			}
		}

		for idx, _ := range config.Servers {
			if config.Servers[idx].Running {
				if !config.Servers[idx].CheckServer() {
					fmt.Printf("Server \"%s\" should be running but is not. Restarting.\n", config.Servers[idx].Name)
					config.Servers[idx].StartServerProcess()
					SaveConfig(configFile, config)
				}
			}
		}
	}
}

func findServer(name string) (s *Server, err error) {
	for idx, _ := range globalConfiguration.Servers {
		if globalConfiguration.Servers[idx].Name == name {
			return &globalConfiguration.Servers[idx], nil
		}
	}
	return &Server{}, errors.New("No server with name: " + name)
}