package statsd

import (
	"fmt"
	"gopkg.in/alexcesaro/statsd.v2"
	"smartRedis/flags"
	"smartRedis/model"
	"smartRedis/status"
	"strings"
	"time"
	"os"
)

func Statsd() {
	fmt.Println(flags.Prefix)
	statsdClient, err := statsd.New(
		statsd.Address(flags.StatsdHostPort),
		statsd.Prefix(flags.Prefix),
	)

	if err != nil {
		fmt.Println("Error Connection Statsd: " + err.Error())
		os.Exit(1)
	} else {
		fmt.Println("Statsd Connected")
	}
	// choose between node and machine stats
	for _ = range time.NewTicker(time.Duration(flags.StatsdPushInterval) * time.Second).C {
		go publishMetrics(statsdClient, flags.RedisHost, flags.RedisPort)

	}

}

func publishMetrics(statsdClient *statsd.Client, hostInput, portInput string) {
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
		//fmt.Println(node.Host,node.Role,node.Port, node.UsedMemory, node.UsedMemoryPeak, node.Hits, node.InstantaneousInputKbps, node.InstantaneousOutputKbps, node.InstantaneousOpsPerSec, node.NonExpiryKeys)
		client.Count(node.Host+"."+node.Role+"."+node.Port+".mem.used", node.UsedMemory)
		client.Count(node.Host+"."+node.Role+"."+node.Port+".mem.peak", node.UsedMemoryPeak)
		client.Count(node.Host+"."+node.Role+"."+node.Port+".keyspace_hits", node.Hits)
		client.Count(node.Host+"."+node.Role+"."+node.Port+".network.input", node.InstantaneousInputKbps)
		client.Count(node.Host+"."+node.Role+"."+node.Port+".network.output", node.InstantaneousOutputKbps)
		client.Count(node.Host+"."+node.Role+"."+node.Port+".ops", node.InstantaneousOpsPerSec)
		client.Count(node.Host+"."+node.Role+"."+node.Port+".nonexpirykeys", node.NonExpiryKeys)
		result <- "SuccessFully pushed to statsd " + node.Host + ":" + node.Port
	}
	client.Flush()
}
