package main

import (
	"fmt"
	"smartRedis/cluster"
	"smartRedis/status"
	"smartRedis/statsd"
	"smartRedis/flags"
)

func main() {
	flags.Init()
	if flags.Action == "status" {
		go status.Status()
	} else if flags.Action == "statsd" {
		statsd.Statsd()
	} else if flags.Action == "create-cluster" {
		cluster.ClusterCreate()
	} else {
		fmt.Println("Usage: ./smartRedis -action=[status|statsd|cluster]")
		fmt.Println("status          -- Show Cluster status.")
		fmt.Println("create-cluster  -- Launch Redis Cluster instances.")
		fmt.Println("statsd          -- Push metrics to statsd.")
	}
}