package main

import (
	"fmt"
	"os"
	"smartRedis/cluster"
	"smartRedis/status"
)

func main() {
	var operation string
	if len(os.Args) > 1 {
		operation = os.Args[1]
	}
	if operation == "status" {
		status.Status()
	} else if operation == "create-cluster" {
		cluster.ClusterCreate()
	} else {
		fmt.Println("Usage: ./redisUtil [status|cluster]")
		fmt.Println("status       	-- Show Cluster status.")
		fmt.Println("create-cluster  -- Launch Redis Cluster instances.")
	}
}
