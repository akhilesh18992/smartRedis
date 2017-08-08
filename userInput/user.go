package userInput

import (
	"log"
	"fmt"
	"strconv"
	"strings"
	"golang.org/x/crypto/ssh/terminal"
)

var REDIS_VERSIONS = []string{"3.2.10", "4.1"}

const MAX_NODES  = 	500

func AskForUsername() string {
	var username string
	fmt.Print("username:")
	_, err := fmt.Scanln(&username)
	if err != nil {
		if err.Error() == "unexpected newline" {

		} else {
			log.Fatal(err)
		}
	}
	username = strings.TrimSpace(username)
	if username != "" {
		return username
	} else {
		fmt.Println("Error: Empty username")
		return AskForUsername()
	}
}
func AskForPassword() string {
	var pass []byte
	fmt.Print("password:")
	pass, err := terminal.ReadPassword(0)
	if err != nil {
		if err.Error() == "unexpected newline" {

		} else {
			log.Fatal(err)
		}
	}
	fmt.Println()
	if len(pass) > 0 {
		return string(pass)
	} else {
		fmt.Println("Error: Empty Password")
		return AskForPassword()
	}
}

func ChooseRedisVersion() string {
	var redisVersionId int
	fmt.Println("Supported Redis Versions:")
	for i, version := range REDIS_VERSIONS {
		fmt.Println(strconv.Itoa(i + 1) + ". " + version)
	}
	fmt.Println("Please choose redis version(1/2/3..):")
	_, err := fmt.Scanln(&redisVersionId)
	if err != nil {
		if err.Error() == "unexpected newline" {

		} else {
			log.Fatal(err)
		}
	}
	if redisVersionId < 1 || redisVersionId > len(REDIS_VERSIONS) {
		return ChooseRedisVersion()
	} else {
		return REDIS_VERSIONS[redisVersionId - 1]
	}
}

func AskForHostPort() (string, string) {
	var redisNodeInfo string
	fmt.Print("enter redis node to monitor(host:port):")
	_, err := fmt.Scanln(&redisNodeInfo)
	if err != nil {
		if err.Error() == "unexpected newline" {

		} else {
			log.Fatal(err)
		}
	}
	redisNodeInfo = strings.TrimSpace(redisNodeInfo)
	parts := strings.Split(redisNodeInfo, ":")
	if len(parts) != 2  {
		fmt.Println("Error: Pass host:port to fetch status")
		return AskForHostPort()
	} else {
		return parts[0], parts[1]
	}
}

