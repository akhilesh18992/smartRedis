package statsd

import (
	"gopkg.in/alexcesaro/statsd.v2"
	"fmt"
	"smartRedis/status"
	"smartRedis/model"
	"time"
	"smartRedis/flags"
	"strings"
)

func Statsd()  {
	fmt.Println(flags.StatsdHostPort, flags.Prefix, flags.RedisPort, flags.RedisHost)

	statsdClient, err := statsd.New(
		statsd.Address(flags.StatsdHostPort),
		statsd.Prefix(flags.Prefix),
	)

	if err != nil {
		panic("Error Connection Statsd: " +  err.Error())
	} else {
		fmt.Println("Statsd Connected")
	}
	// choose between node and machine stats
	for _ = range time.NewTicker(time.Duration(20)*time.Second).C {
		go publishMetrics(statsdClient, flags.RedisHost, flags.RedisPort)

	}

}

func publishMetrics(statsdClient *statsd.Client, hostInput, portInput string)  {
	var nodesInfo model.NodesInfo
	if strings.Contains(portInput, ",") {
		nodesInfo = status.GetNodeInfo(hostInput, portInput)
	} else {
		nodesInfo, _, _ = status.GetClusterNodesInfo(hostInput, portInput, false)
	}

	result := make(chan string, len(nodesInfo))
	nodeList := make(chan model.NodeInfo, len(nodesInfo))

	for w := 1; w <= 20; w++ {
		go publishStatsWorker(statsdClient, nodeList, result)
	}

	for _, node := range nodesInfo {
		nodeList <- node
	}
	close(nodeList)
	for i := 1; i <= len(nodesInfo); i++ {
		fmt.Println(<-result)
	}
}

func publishStatsWorker(client *statsd.Client, nodeList chan model.NodeInfo, result chan string) {
	for node := range nodeList {
		client.Count(node.Host + "." + node.Role + "." + node.Port + ".mem.used", node.UsedMemory)
		client.Count(node.Host + "." + node.Role + "." + node.Port + ".mem.usedPeak", node.UsedMemoryPeak)
		client.Count(node.Host + "." + node.Role + "." + node.Port + ".keyspace_hits", node.Hits)
		client.Count(node.Host + "." + node.Role + "." + node.Port + ".network.input", node.InstantaneousInputKbps)
		client.Count(node.Host + "." + node.Role + "." + node.Port + ".network.output", node.InstantaneousOutputKbps)
		client.Count(node.Host + "." + node.Role + "." + node.Port + ".ops.output", node.InstantaneousOpsPerSec)
		client.Count(node.Host + "." + node.Role + "." + node.Port + ".nonexpirykeys", node.NonExpiryKeys)
		result <- "SuccessFully pushed to statsd " + node.Host + ":" + node.Port
	}
	client.Flush()
}