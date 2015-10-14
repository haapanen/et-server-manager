package main
import (
	"os/user"
	"strconv"
	"os"
	"net"
	"time"
	"fmt"
	"bufio"
	"io/ioutil"
	"errors"
	"strings"
	"os/exec"
	"syscall"
	"path/filepath"
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

//
// Returns the status response of the server in a parsed format
func (s *Server) GetStatus() (*StatusResponse, error) {
	conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:%d", s.Ip, s.Port), 1000 * time.Millisecond)
	if err != nil {
		return &StatusResponse{}, err
	}
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	defer conn.Close()

	if err != nil {
		return &StatusResponse{}, err
	}

	fmt.Fprintf(conn, "\xff\xff\xff\xffgetstatus")
	p := make([]byte, 1024)
	_, err = bufio.NewReader(conn).Read(p)

	if err != nil {
		return &StatusResponse{}, err
	}

	sr := ParseStatusResponse(string(p))
	return sr, nil
}

//
// Prints the server status
func (s *Server) PrintStatus(verbose bool) {
	if !s.Running {
		fmt.Printf("Server \"%s\" is not running.\n", s.Name)
		return
	}

	if (verbose) {
		status, err := s.GetStatus()
		if err != nil {
			fmt.Printf("Could not get server status: %s\n", err)
			return
		}

		fmt.Println("----------------------------------------")
		fmt.Printf("Server:\t\t%s\n", s.Name)
		fmt.Println("----------------------------------------")
		fmt.Printf("Address:\t%s:%d\n", s.Ip, s.Port)
		fmt.Println("----------------------------------------")
		fmt.Printf("Hostname:\t%s\n", StripColors(status.Keys["sv_hostname"]))
		fmt.Printf("Players:\t%d/%s\n", len(status.Players), status.Keys["sv_maxclients"])
		fmt.Printf("Map:\t\t%s\n", status.Keys["mapname"])
		fmt.Println("----------------------------------------")
		if len(status.Players) > 0 {
			fmt.Printf("Players:\t%s\n", status.Players[0])
			for _, player := range status.Players[1:] {
				fmt.Printf("\t\t%s\n", StripColors(player))
			}
			fmt.Println("----------------------------------------")
		}

	} else {
		fmt.Println("----------------------------------------")
		fmt.Printf("Server:\t\t%s\n", s.Name)
		fmt.Println("----------------------------------------")
		fmt.Printf("Address:\t%s:%d\n", s.Ip, s.Port)
		fmt.Printf("Running:\t%t\n", s.Running)
		fmt.Println("----------------------------------------")
	}
}

//
// Checks if the server is running and starts the server process if not
// Informs the user of the current server status
func (s *Server) Start() (error) {
	contains, err := s.basePathContainsNecessaryFiles()
	if !contains {
		return errors.New(fmt.Sprintf("Server \"%s\" could not be started: %s\n", s.Name, err))
	}

	if !s.homePathExistsInFS() {
		return errors.New(fmt.Sprintf("Server \"%s\" has an invalid homepath\n", s.Name))
	}

	if s.Running {
		if !s.CheckServer() {
			fmt.Printf("Server \"%s\" should be running, but it is not. Restarting server.\n", s.Name)
			_, err := s.StartServerProcess()
			return err
		} else {
			fmt.Printf("Server \"%s\" is already running. (Process ID: %d)\n", s.Name, s.Pid)
			fmt.Printf("Run \"screen -drS %s\" to open server console.\n", s.Name + strconv.Itoa(s.Port))
			return nil
		}
	}

	serverProcess, err := s.StartServerProcess()
	if err != nil {
		fmt.Printf("Could not start server process: %s\n", err)
		return err
	}

	err = SaveConfig(configFile, globalConfiguration)
	if err != nil {
		fmt.Printf("Error: could not save config \"%s\": %s\n", err)
	}

	fmt.Printf("Started server \"%s\" with process ID: %d\n", s.Name, serverProcess.Process.Pid + 1)
	fmt.Printf("%v\n", serverProcess)
	return nil
}

//
// Starts the server process if it's not already running
func (s *Server) StartServerProcess() (*exec.Cmd, error) {
	user, err := user.Lookup(s.User)
	if err != nil {
		return &exec.Cmd{}, err
	}

	parameters := []string{
		"-dmS",
		s.Name + strconv.Itoa(s.Port),
		globalConfiguration.ExecutablePath,
		fmt.Sprintf("+set fs_game \"%s\"", s.Mod),
	}

	for _, config := range s.Configs {
		parameters = append(parameters, fmt.Sprintf("exec \"%s\"", config))
	}

	parameters = append(parameters, fmt.Sprintf("+set com_hunkmegs \"128\""))
	parameters = append(parameters, fmt.Sprintf("+set fs_basepath \"%s\"", s.BasePath))
	parameters = append(parameters, fmt.Sprintf("+set fs_homepath \"%s\"", s.HomePath))
	parameters = append(parameters, fmt.Sprintf("+set net_ip \"%s\"", s.Ip))
	parameters = append(parameters, fmt.Sprintf("+set net_port \"%d\"", s.Port))
	parameters = append(parameters, fmt.Sprintf("+map oasis"))

	serverProcess := exec.Command(globalConfiguration.ScreenPath, parameters...)
	serverProcess.Dir = filepath.Dir(globalConfiguration.ExecutablePath)
	// Needed for ET Process, saves .etwolf folder here
	serverProcess.Env = []string{fmt.Sprintf("HOME=%s", user.HomeDir)}

	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return serverProcess, err
	}
	gid, err := strconv.Atoi(user.Gid)
	if err != nil {
		return serverProcess, err
	}

	serverProcess.SysProcAttr = &syscall.SysProcAttr{}
	serverProcess.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	err = serverProcess.Run()
	if err != nil {
		return serverProcess, err
	}
	s.Running = true
	s.Pid = serverProcess.Process.Pid + 1
	return serverProcess, nil

}

//
// Stops the server if it's running
func (s *Server) Stop() (err error) {
	err = nil
	if s.Running {
		fmt.Printf("Stopping server \"%s\"\n", s.Name)

		proc := exec.Command(globalConfiguration.KillPath, strconv.Itoa(s.Pid))
		err := proc.Run()
		fmt.Printf("Stopping server with pid: %d\n", s.Pid)
		if err != nil {
			fmt.Printf("Could not stop server: %s\n", err)
			return err
		}
		s.Running = false
		err = SaveConfig(configFile, globalConfiguration)
		fmt.Printf("Stopped server \"%s\"\n", s.Name)
	} else {
		fmt.Printf("Server \"%s\" is not running.\n", s.Name)
	}
	return err
}

//
// Restarts the server
func (s *Server) Restart() {
	s.Stop()
	fmt.Printf("Waiting 2 seconds")

	fmt.Printf(".")
	time.Sleep(700 * time.Millisecond)
	fmt.Printf(".")
	time.Sleep(700 * time.Millisecond)
	fmt.Printf(".\n")
	time.Sleep(600 * time.Millisecond)

	s.Start()
}

//
// Checks that the server is still running
func (s *Server) CheckServer() (bool) {
	if s.Running {
		pgrep := exec.Command(globalConfiguration.PgrepPath, "-f", s.Name + strconv.Itoa(s.Port))

		_, err := pgrep.Output()
		if err != nil {
			return false
		}
	}
	return true
}

//
// Checks that the calling user is allowed to execute the script
func (s *Server) SecurityCheck() (bool, error) {
	owner, err := user.Lookup(s.User)
	if err != nil {
		return false, err
	}

	serverOwnerUid, err := strconv.Atoi(owner.Uid)
	if err != nil {
		return false, err
	}

	currentUid := os.Geteuid()
	if currentUid == serverOwnerUid || currentUid == 0 {
		return true, nil
	}
	return false, nil
}

//
// Checks if the basepath contains necessary files to host an ET server
// pak0.pk3, pak1.pk3, pak2.pk3 and mp_bin.pk3
func (s *Server) basePathContainsNecessaryFiles() (bool, error) {
	fileInfos, err := ioutil.ReadDir(filepath.Join(s.BasePath, "etmain"))
	if err != nil {
		return false, errors.New(fmt.Sprintf("Basepath check failed: %s", err))
	}

	requiredFiles := map[string]bool{
		"pak0.pk3": false,
		"pak1.pk3": false,
		"pak2.pk3": false,
		"mp_bin.pk3": false,
	}

	// Check that required files can be found at the basepath etmain
	for _, fi := range fileInfos {
		if _, ok := requiredFiles[fi.Name()]; ok {
			requiredFiles[fi.Name()] = true
		}
	}

	missingFiles := []string{}
	for file, exists := range requiredFiles {
		if !exists {
			missingFiles = append(missingFiles, file)
		}
	}

	if len(missingFiles) > 0 {
		return false, errors.New(fmt.Sprintf("Missing the following files: %s", strings.Join(missingFiles, ", ")))
	}
	return true, nil
}

//
// Checks if server's homepath actually exists in file system
func (s *Server) homePathExistsInFS() (bool) {
	fileInfo, err := os.Stat(s.HomePath)

	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}