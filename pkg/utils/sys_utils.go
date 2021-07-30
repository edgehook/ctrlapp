package utils

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/exp/rand"
)

/*
* GetLocalMACs
* get the local host's macaddress.
 */
func GetLocalMACs() []string {

	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	macAddrs := make([]string, 0)

	for _, ni := range netInterfaces {
		if !strings.HasPrefix(ni.Name, "e") &&
			!strings.HasPrefix(ni.Name, "w") &&
			!strings.HasPrefix(ni.Name, "p") {
			continue
		}
		macAddr := ni.HardwareAddr.String()
		if len(macAddr) > 6 {
			macAddrs = append(macAddrs, strings.ToUpper(macAddr))
		}
	}

	return macAddrs
}

/*
* GetLocalIPs
* get all IP of all network interfaces.
 */
func GetLocalIPs() []string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}

	ips := make([]string, 0)

	for _, addr := range addrs {
		ipNet, isThisType := addr.(*net.IPNet)
		if isThisType && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}

	return ips
}

/*
* GetOsType:
* return lowcase ostype.
 */
func GetOsType() string {
	return runtime.GOOS
}

func RandomInt(min, max int) int {
	if min >= max || min == 0 || max == 0 {
		return max
	}

	rand.Seed(uint64(time.Now().Unix()))
	return rand.Intn(max-min) + min
}

func FileCopy(src, dst string) error {
	isWindows := strings.Contains(GetOsType(), "windows")
	if isWindows {
		fileName := filepath.Base(src)
		targetDir := filepath.Dir(dst)

		cmd := exec.Command("xcopy", src, targetDir, "/I", "/E", "/F", "/O", "/Y")
		_, err := cmd.Output()
		if err != nil {
			return err
		}

		return os.Rename(filepath.Join(targetDir, fileName), dst)
	}

	// default for linux.
	cmd := exec.Command("cp", "-r", src, dst)
	_, err := cmd.Output()

	return err
}

/*
* Execute:
* Execute the command without paramenta.
 */
func Execute(command string) (string, error) {
	if strings.Contains(GetOsType(), "windows") {
		cmd := exec.Command("cmd", "/C", command)
		cmd.Env = os.Environ()
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}

		return string(output), nil
	}

	// default for unix.
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

/*
* Execute1:
* Execute1 the command with paramenta.
 */
func Execute1(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

/*
* Execute2:
* Execute the command realtime.
 */
func Execute2(command string, callback func(io.ReadCloser, io.ReadCloser)) error {
	cmd := exec.Command("/bin/bash", "-c", command)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	err := cmd.Start()
	if err != nil {
		return err
	}

	if callback != nil {
		callback(stdout, stderr)
	}

	return cmd.Wait()
}

/*
* XmlParser
* parse the xml file.
 */
func XmlParser(fileName string, v interface{}) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	return xml.Unmarshal(data, v)
}

//read file from filepath
func GetFileContent(filePath string) ([]byte, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func WriteFileContent(filePath string, content []byte) error {
	return ioutil.WriteFile(filePath, content, 0755)
}

/*
* JsonParser
* parse the xml file.
 */
func JsonParser(fileName string, v interface{}) error {
	data, err := GetFileContent(fileName)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}
