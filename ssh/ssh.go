package ssh

import (
	"golang.org/x/crypto/ssh"
	"time"
	"fmt"
	"bytes"
	"strings"
)

var config *ssh.ClientConfig

// returns hostname from ip through ssh, if fails then returns ip
func GetHostname(ip string) string {
	if config == nil {
		return ip
	}
	fmt.Println("Resolving host for " + ip)
	client, err := ssh.Dial("tcp", ip + ":22", config)
	if err != nil {
		fmt.Println("error resolving hostname. Unable to ssh to " + ip)
		return ip
	}
	return strings.Trim(executeCmd(client, "/bin/hostname"), "\t\n\r")
}

// initializing ssh config
func Config(username, password string)  {
	config = &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		Timeout:         5 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO fix me, should not be used in prod
	}
}

// returns output of shell command running over ssh
func executeCmd(conn *ssh.Client, cmd string) string {
	session, _ := conn.NewSession()
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Run(cmd)

	return stdoutBuf.String()
}
