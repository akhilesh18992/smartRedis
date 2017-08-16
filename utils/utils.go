package utils

import "strconv"

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
