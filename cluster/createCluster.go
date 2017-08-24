package cluster

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net/http"
	"os"
	"smartRedis/userInput"
	"strings"
	"time"
)

func copyRedisTar(client *ssh.Client, redisVersion string) {
	sftp, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("SFTP connectiong created")
	}
	defer sftp.Close()

	srcPath := "/home/akhilesh.singh/"
	dstPath := "/home/akhilesh.singh/"
	filename := redisVersion
	// Open the source file
	srcFile, err := os.Open(srcPath + filename)
	r := io.Reader(srcFile)
	if err != nil {
		log.Fatal(err.Error() + srcPath + filename)
	} else {
		log.Println("Opening source file: " + srcPath + filename)
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := sftp.Create(dstPath + filename)
	if err != nil {
		log.Fatal(err.Error() + dstPath + filename)
	} else {
		log.Println("Destination File created: " + dstPath + filename)
	}
	defer dstFile.Close()

	// Copy the file
	dstFile.ReadFrom(r)
	log.Println("COPIED...DONE")
}

func getHost(hostPort []string) (hosts map[string]bool, err error) {
	for _, hp := range hostPort {
		parts := strings.Split(hp, ":")
		if len(parts) != 2 {
			err = errors.New("Wrong Host Port Configuration")
			return
		}
	}
	return
}

func ClusterCreate() {
	username := userInput.AskForUsername()
	pass := userInput.AskForPassword()
	redisVersion := userInput.ChooseRedisVersion()
	//hostPort := userInput.AskForHostsPort()
	//hosts, err := getHost(hostPort)
	//if err != nil {
	//
	//}
	//log.Println(hosts)
	//results := make(chan string, 10)
	//timeout := time.After(5 * time.Second)
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		Timeout:         5 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO fix me, should not be used in prod
	}
	fmt.Println("Dialing......")
	client, err := ssh.Dial("tcp", "nmc-idp4:22", config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	fmt.Println("Dialed success")
	//session, err := client.NewSession()
	//if err != nil {
	//	panic("Failed to create session: " + err.Error())
	//}
	//defer session.Close()
	//var b bytes.Buffer
	//session.Stdout = &b
	//fmt.Println("Running command")
	//if err := session.Run("/usr/bin/whoami"); err != nil {
	//	panic("Failed to run: " + err.Error())
	//}
	//fmt.Println(b.String())
	log.Println(executeCmd(client, "/usr/bin/whoami"))
	redisTar := "redis-" + redisVersion + ".tar.gz"
	out, err := os.Create("/home/akhilesh.singh/" + redisTar)
	if err != nil {
		panic("Failed to create file: " + err.Error())
	} else {
		log.Println("File created: " + "/home/akhilesh.singh/" + redisTar)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get("http://download.redis.io/releases/" + redisTar)
	if err != nil {
		panic("Failed to download file: " + err.Error())
	} else {
		log.Println("File downloaded: " + redisTar)
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic("Failed to copy file: " + err.Error())
	} else {
		log.Println("File copied")
	}

	copyRedisTar(client, redisTar)

	//for _, hostname := range hosts {
	//	go func(hostname string) {
	//		results <- executeCmd(cmd, hostname, config)
	//	}(hostname)
	//}
	//
	//for i := 0; i < len(hosts); i++ {
	//	select {
	//	case res := <-results:
	//		fmt.Print(res)
	//	case <-timeout:
	//		fmt.Println("Timed out!")
	//		return
	//	}
	//}
}

func executeCmd(conn *ssh.Client, cmd string) string {
	session, _ := conn.NewSession()
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Run(cmd)

	return conn.User() + ": " + stdoutBuf.String()
}
