package utils

import (
	"strconv"
	"log"
	"os/exec"
	"strings"
)

func ReadableMemory(mem int) string {
	sizeSuffix := "B"
	size := mem
	if size > 1024 {
		sizeSuffix = "KB"
		size /= 1024
	}
	if size > 1024 {
		sizeSuffix = "MB"
		size /= 1024
	}
	if size > 1024 {
		sizeSuffix = "GB"
		size /= 1024
	}
	return strconv.Itoa(size) + sizeSuffix
}

func ExecCmd() string {
	out, err := exec.Command("/usr/bin/whoami").Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Trim(string(out), "\n\t\r")
}