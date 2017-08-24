package userInput

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"smartRedis/utils"
	"strconv"
	"strings"
)

var REDIS_VERSIONS = []string{"3.2.10", "4.1"}

const MAX_NODES = 500

func AskForUsername() string {
	var username string
	defaultUsername := utils.ExecCmd("/usr/bin/whoami")
	fmt.Print("username(default " + defaultUsername + "):")
	_, err := fmt.Scanln(&username)
	if err != nil {
		if err.Error() == "unexpected newline" {

		} else {
			log.Fatal(err)
		}
	}
	if username == "" {
		username = defaultUsername
	}
	return strings.TrimSpace(username)
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
		fmt.Println(strconv.Itoa(i+1) + ". " + version)
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
		return REDIS_VERSIONS[redisVersionId-1]
	}
}

func AskForHostPort() (host, port string) {
	// ask for host
	fmt.Print("Enter host(default localhost):")
	_, err := fmt.Scanln(&host)
	if err != nil {
		if err.Error() == "unexpected newline" {

		} else {
			log.Fatal(err)
		}
	}
	host = strings.TrimSpace(host)

	// ask for host
	fmt.Print("Enter port(default 6379):")
	_, err = fmt.Scanln(&port)
	if err != nil {
		if err.Error() == "unexpected newline" {

		} else {
			log.Fatal(err)
		}
	}
	port = strings.TrimSpace(port)
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "6379"
	}
	return
}

func AskForUsernamePassword() (username, password, consent string) {
	// ask for username password
	fmt.Print("Resolve IP to Hostname(y/n):")
	_, err := fmt.Scanln(&consent)
	if err != nil {
		if err.Error() == "unexpected newline" {

		} else {
			log.Fatal(err)
		}
	}
	consent = strings.TrimSpace(consent)
	if consent == "y" || consent == "Y" {
		username = AskForUsername()
		password = AskForPassword()
	}
	return
}

func AskForUsernamePasswordWithoutConsent() (username, password string) {
	// ask for username password
	username = AskForUsername()
	password = AskForPassword()
	return
}
