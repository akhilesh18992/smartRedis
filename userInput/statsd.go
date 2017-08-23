package userInput

import (
	"fmt"
	"log"
	"strings"
)

func AskForStatsdHost() (host string) {
	fmt.Print("Enter statsd host:")
	_, err := fmt.Scanln(&host)
	if err != nil {
		if err.Error() == "unexpected newline" {

		} else {
			log.Fatal(err)
		}
	}
	if len(host) > 0 {
		return strings.TrimSpace(host)
	} else {
		fmt.Println("Error: Empty host")
		return AskForStatsdHost()
	}
}

func AskForStatsdPort() (port string) {
	fmt.Print("Enter statsd port:")
	_, err := fmt.Scanln(&port)
	if err != nil {
		if err.Error() == "unexpected newline" {

		} else {
			log.Fatal(err)
		}
	}
	if len(port) > 0 {
		return strings.TrimSpace(port)
	} else {
		fmt.Println("Error: Empty port")
		return AskForStatsdHost()
	}
}

func AskForStatsdPrefix() (prefix string) {
	fmt.Print("Enter statsd prefix:")
	_, err := fmt.Scanln(&prefix)
	if err != nil {
		if err.Error() == "unexpected newline" {

		} else {
			log.Fatal(err)
		}
	}
	if len(prefix) > 0 {
		return strings.TrimSpace(prefix)
	} else {
		fmt.Println("Error: Empty prefix")
		return AskForStatsdHost()
	}
}
