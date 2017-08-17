package status

import (
	"fmt"
	"os/exec"
	"smartRedis/color"
	"smartRedis/display"
	"smartRedis/model"
	"smartRedis/userInput"
	"sort"
	"strconv"
	"strings"
	"smartRedis/ssh"
)

func Status() {
	host, port := userInput.AskForHostPort()
	username, password, consent := userInput.AskForUsernamePassword()
	if consent == "y" || consent == "Y" {
		ssh.Config(username, password)
	}
	nodesTableInfo, masterSlaveIpMap, totalMasters := getNodesInfo(host, port)
	if len(masterSlaveIpMap) == 0 {
		fmt.Println("Wrong Host Port")
		return
	}
	machineStats := make(map[string]model.MachineStats)
	for _, node := range nodesTableInfo {
		updateMachineStats(machineStats, node)
	}
	nodeStatsError := display.DisplayNodeStats(nodesTableInfo, masterSlaveIpMap)
	if nodeStatsError != nil {
		fmt.Println(color.BRed(nodeStatsError.Error()))
	}
	machineStatsError := display.DisplayMachineStats(machineStats, totalMasters)
	if machineStatsError != nil {
		fmt.Println(color.BRed(machineStatsError.Error()))
	}
}

func redisInfoWorker(nodeList chan model.NodeInfo, redisInfo chan model.NodeInfo) {
	for node := range nodeList {
		// TODO handle if node down
		cmd := "redis-cli -h " + node.Ip + " -p " + node.Port + " info"
		out, _ := exec.Command("sh", "-c", cmd).Output()
		var nodeDetail model.NodeInfo
		var info model.RedisInfo
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "redis_version:") {
				info.Version = strings.Split(line, ":")[1]
			} else if strings.HasPrefix(line, "redis_mode:") {
				info.Mode = strings.Split(line, ":")[1]
			} else if strings.HasPrefix(line, "process_id:") {
				info.Pid, _ = strconv.Atoi(strings.Split(line, ":")[1])
			} else if strings.HasPrefix(line, "uptime_in_seconds:") {
				info.Uptime, _ = strconv.Atoi(strings.Split(line, ":")[1])
			} else if strings.HasPrefix(line, "maxmemory:") {
				info.MaxMemory, _ = strconv.Atoi(strings.Split(line, ":")[1])
			} else if strings.HasPrefix(line, "maxmemory_policy:") {
				info.EvictionPolicy = strings.Split(line, ":")[1]
			} else if strings.HasPrefix(line, "instantaneous_ops_per_sec:") {
				info.InstantaneousOpsPerSec, _ = strconv.Atoi(strings.Split(line, ":")[1])
			} else if strings.HasPrefix(line, "instantaneous_input_kbps:") {
				info.InstantaneousInputKbps, _ = strconv.ParseFloat(strings.Split(line, ":")[1], 2)
			} else if strings.HasPrefix(line, "instantaneous_output_kbps:") {
				info.InstantaneousOutputKbps, _ = strconv.ParseFloat(strings.Split(line, ":")[1], 2)
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

func getNodesInfo(host, port string) (model.NodesInfo, map[string][]string, int) {
	var nodes []model.NodeInfo
	// TODO handle when redis-cli is not there
	cmd := "redis-cli -h " + host + " -p " + port + " cluster nodes"
	out, _ := exec.Command("sh", "-c", cmd).Output()
	clusterStatus := string(out)
	nodeDetail := strings.Split(clusterStatus, "\n")
	masterSlaveIpMap := make(map[string][]string)
	//machineMasterCountMap := make(map[string]int)
	totalMasters := 0
	IpHostMap := make(map[string]string)
	for _, nd := range nodeDetail {
		nodeDetailList := strings.Split(nd, " ")
		if len(nodeDetailList) <= 1 {
			continue
		}
		var nodeInfo model.NodeInfo
		nodeInfo.NodeId = nodeDetailList[0]
		hostPort := strings.Split(nodeDetailList[1], ":")
		nodeInfo.Ip, nodeInfo.Port = hostPort[0], hostPort[1]
		if IpHostMap[nodeInfo.Ip] == "" {
			IpHostMap[nodeInfo.Ip] = ssh.GetHostname(nodeInfo.Ip)
		}
		if strings.Contains(nodeDetailList[2], "slave") {
			masterSlaveIpMap[nodeDetailList[3]] = append(masterSlaveIpMap[nodeInfo.MasterId], nodeDetailList[1])
			nodeInfo.Type = model.SLAVE
		} else {
			nodeInfo.Type = model.MASTER
			nodeInfo.HashSlot = nodeDetailList[8]
			slotRange := strings.Split(nodeInfo.HashSlot, "-")
			nodeInfo.HashSlotStart, _ = strconv.Atoi(slotRange[0])
			nodeInfo.HashSlotEnd, _ = strconv.Atoi(slotRange[1])
			//machineMasterCountMap[nodeInfo.Ip] += 1
			totalMasters += 1

		}
		nodeInfo.Host = IpHostMap[nodeInfo.Ip]
		nodes = append(nodes, nodeInfo)
	}
	nodesTableInfo := getRedisInfo(nodes)
	sort.Sort(nodesTableInfo)
	return nodesTableInfo, masterSlaveIpMap, totalMasters
}

func updateMachineStats(machineStats map[string]model.MachineStats, node model.NodeInfo) {
	stats := machineStats[node.Ip]
	stats.RedisMemory += node.UsedMemory
	if node.Host != "" {
		stats.Hostname = node.Host
	}
	if node.Type == model.MASTER {
		stats.Master += 1
	} else {
		stats.Slave += 1
	}
	stats.RedisNodes += 1
	stats.Memory = node.SystemMemory
	stats.OpsPerSec += node.InstantaneousOpsPerSec
	stats.NetworkBandwidth += node.InstantaneousOutputKbps + node.InstantaneousInputKbps
	machineStats[node.Ip] = stats
}
