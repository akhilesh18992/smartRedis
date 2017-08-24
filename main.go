package main

import (
	"fmt"
	"smartRedis/cluster"
	"smartRedis/flags"
	"smartRedis/statsd"
	"smartRedis/status"
)

func main() {
	flags.Init()
	if flags.Action == "status" {
		status.Status()
	} else if flags.Action == "statsd" {
		statsd.Statsd()
	} else if flags.Action == "cluster" {
		cluster.ClusterCreate()
	} else {
		fmt.Println("Usage: ./smartRedis -action=[status|statsd|cluster] -redisHost=localhost -redisPort=30001 -resolveHostname=n")
		fmt.Println("status          -- Show Cluster status.")
		fmt.Println("cluster         -- Launch Redis Cluster instances.")
		fmt.Println("statsd          -- Push metrics to statsd.")
	}
}
