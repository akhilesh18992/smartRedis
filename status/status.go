package status

import (
	"strings"
	"os/exec"
	"strconv"
	"redisUtility/table"
	"redisUtility/model"
	"sort"
	"redisUtility/userInput"
	"fmt"
	"redisUtility/color"
)

const SLAVE   = "SLAVE"
const MASTER  = "MASTER"

func Status() {
	host, port := userInput.AskForHostPort()
	nodesTableInfo, masterSlaveIpMap, totalMasters := getNodesInfo(host, port)

	var masterSlaveOnSameMachine bool
	machineStats := make(map[string]model.MachineStats)

	tw := table.Init()
	for _, node := range nodesTableInfo {
		updateMachineStats(machineStats, node)
		if node.Type == SLAVE {
			continue
		}
		cacheMiss := strconv.FormatFloat((float64(node.Miss)/float64(node.Hits))*100, 'f', 3, 64)
		colorCode := color.GREEN
		if isMasterSlaveOnSameMachine(masterSlaveIpMap[node.NodeId], node.Ip) {
			colorCode = color.RED
		}
		tw.Append([]string{node.Ip, node.Port, humanReadableMemory(node.UsedMemory), humanReadableMemory(node.UsedMemoryPeak), cacheMiss,
			strings.Join(masterSlaveIpMap[node.NodeId], ","), node.HashSlot}, colorCode)
	}
	tw.SetHeader([]string{"Host", "Port", "Data Size", "Peak Mem Used", "Cache Miss", "Slave Node", "Slot"})
	tw.Render()
	if masterSlaveOnSameMachine {
		fmt.Println(color.BRed("CLUSTER ERROR: Master slave on same machine"))
	}
	displayMachineStats(machineStats, totalMasters)
}

func redisInfoWorker(nodeList chan model.NodeInfo, redisInfo chan model.NodeInfo) {
	for node := range nodeList {
		// TODO handle if node down
		cmd := "redis-cli -h " + node.Ip + " -p " + node.Port + " info"
		out, _ := exec.Command("sh","-c", cmd).Output()
		var nodeDetail model.NodeInfo
		var info model.RedisInfo
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "redis_version") {
				info.RedisVersion = strings.Split(line, ":")[1]
			} else if strings.HasPrefix(line, "keyspace_hits") {
				info.Hits, _ = strconv.Atoi(strings.Split(line, ":")[1])
			} else if strings.HasPrefix(line, "keyspace_misses") {
				info.Miss, _ = strconv.Atoi(strings.Split(line, ":")[1])
			} else if strings.HasPrefix(line, "used_memory:") {
				info.UsedMemory, _ = strconv.Atoi(strings.Split(line, ":")[1])
			} else if strings.HasPrefix(line, "total_system_memory:") {
				info.SystemMemory, _ = strconv.Atoi(strings.Split(line, ":")[1])
			} else if strings.HasPrefix(line, "used_memory_peak:") {
				info.UsedMemoryPeak, _ = strconv.Atoi(strings.Split(line, ":")[1])
			} else if strings.Contains(line, "avg_ttl") {
				if strings.Contains(line, "db0") {
					db0 := strings.Split(line, ":")[1]
					parts := strings.Split(db0, ",")
					keys, _ := strconv.Atoi(strings.Split(parts[0], "=")[1])
					expires, _ := strconv.Atoi(strings.Split(parts[1], "=")[1])
					info.NonExpiryKeys = keys - expires
				}
			}
		}
		nodeDetail.CopyFrom(node)
		nodeDetail.RedisInfo = info
		redisInfo <- nodeDetail
	}
}

func getRedisInfo(nodes []model.NodeInfo) model.NodesInfo {
	redisInfo := make(chan model.NodeInfo, len(nodes))
	nodeList := make(chan model.NodeInfo, len(nodes))

	for w := 1; w <= 3; w++ {
		go redisInfoWorker(nodeList, redisInfo)
	}

	for _, node := range nodes {
		nodeList <- node
	}
	close(nodeList)
	var nodeDetail []model.NodeInfo
	for i := 1; i <= len(nodes); i++ {
		nodeDetail = append(nodeDetail, <-redisInfo)
	}
	return nodeDetail
}

func humanReadableMemory(mem int) (string) {
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

func getNodesInfo(host, port string) (model.NodesInfo, map[string][]string, int) {
	var nodes []model.NodeInfo
	// TODO handle when redis-cli is not there
	cmd := "redis-cli -h " + host + " -p " + port + " cluster nodes"
	out, _ := exec.Command("sh","-c", cmd).Output()
	clusterStatus := string(out)
	nodeDetail := strings.Split(clusterStatus, "\n")
	masterSlaveIpMap := make(map[string][]string)
	//machineMasterCountMap := make(map[string]int)
	totalMasters := 0
	for _, nd := range nodeDetail {
		nodeDetailList := strings.Split(nd, " ")
		if len(nodeDetailList) <= 1 {
			continue
		}
		var nodeInfo model.NodeInfo
		nodeInfo.NodeId = nodeDetailList[0]
		hostPort := strings.Split(nodeDetailList[1], ":")
		nodeInfo.Ip, nodeInfo.Port = hostPort[0], hostPort[1]

		if strings.Contains(nodeDetailList[2], "slave") {
			masterSlaveIpMap[nodeDetailList[3]] = append(masterSlaveIpMap[nodeInfo.MasterId], nodeDetailList[1])
			nodeInfo.Type = SLAVE
		} else {
			nodeInfo.Type = MASTER
			nodeInfo.HashSlot = nodeDetailList[8]
			slotRange := strings.Split(nodeInfo.HashSlot, "-")
			nodeInfo.HashSlotStart, _ = strconv.Atoi(slotRange[0])
			nodeInfo.HashSlotEnd, _ = strconv.Atoi(slotRange[1])
			//machineMasterCountMap[nodeInfo.Ip] += 1
			totalMasters += 1

		}
		nodes = append(nodes, nodeInfo)
	}
	nodesTableInfo := getRedisInfo(nodes)
	sort.Sort(nodesTableInfo)
	return nodesTableInfo, masterSlaveIpMap, totalMasters
}

func isMasterSlaveOnSameMachine(slaveIps []string, masterId string) (masterSlaveOnSameMachine bool) {
	for _, hostPort := range slaveIps {
		hostPort := strings.Split(hostPort, ":")
		slaveId := hostPort[0]
		if slaveId == masterId {
			masterSlaveOnSameMachine = true
			break
		}
	}
	return
}

func updateMachineStats(machineStats map[string]model.MachineStats, node model.NodeInfo)  {
	stats := machineStats[node.Ip]
	stats.RedisMemory += node.UsedMemory
	if node.Type == MASTER {
		stats.Master += 1
	} else {
		stats.Slave += 1
	}
	stats.RedisNodes += 1
	stats.Memory = node.SystemMemory
	machineStats[node.Ip] = stats
}

func displayMachineStats(machineStats map[string]model.MachineStats, totalMaster int) {
	unbalancedCluster := false
	avgMaster := totalMaster/len(machineStats)
	t := table.Init()
	t.SetHeader([]string{"Machine", "Space Used", "Available(Percentage)", "Master", "Slave"})
	fmt.Println("\n\nTotal masters: " + strconv.Itoa(totalMaster))
	for ip, stats := range machineStats {
		colorCode := color.GREEN
		if stats.Master > avgMaster {
			colorCode = color.RED
			unbalancedCluster = true
		}
		spaceUsed := humanReadableMemory(stats.RedisMemory)
		available := humanReadableMemory(stats.Memory - stats.RedisMemory)
		availablePercentage := "0"
		if stats.Memory != 0 {
			availablePercentage = strconv.FormatFloat((float64(stats.Memory - stats.RedisMemory) / float64(stats.Memory))*100, 'f', 2, 64)
		}
		t.Append([]string{ip, spaceUsed, available + "(" + availablePercentage + "%)", strconv.Itoa(stats.Master),
			strconv.Itoa(stats.Slave)}, colorCode)
	}
	t.Render()
	if unbalancedCluster {
		fmt.Println(color.BRed("CLUSTER UNBALANCED: masters non uniformly distributed across cluster"))
	}
}