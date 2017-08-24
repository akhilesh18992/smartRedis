package utils

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
)

// convert memory in Byte to human readable
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

// execute shell command
func ExecCmd(cmd string) string {
	out, err := exec.Command(cmd).Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Trim(string(out), "\n\t\r")
}
