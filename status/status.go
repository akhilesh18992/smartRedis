package status

import (
	"fmt"
	"os/exec"
	"smartRedis/diagnostics"
	"smartRedis/display"
	"smartRedis/flags"
	"smartRedis/model"
	"smartRedis/ssh"
	"smartRedis/userInput"
	"smartRedis/utils"
	"sort"
	"strconv"
	"strings"
	"os"
)

var IpHostMap = make(map[string]string)

func Status() {
	// optional repeat and interval for auto repeat
	// colorful flag
	host, port := flags.RedisHost, flags.RedisPort
	if host == "" || port == "" {
		fmt.Println("Enter flags -redisHost -redisPort")
		os.Exit(1)
	}
	var username, password string
	resolveHostname := flags.ResolveHostname
	if resolveHostname == "y" || resolveHostname == "Y" {
		username, password = userInput.AskForUsernamePassword()
		ssh.Config(username, password)
	}
	for {
		nodesTableInfo, masterSlaveIpMap, totalMasters := GetClusterNodesInfo(host, port, true)
		if len(masterSlaveIpMap) == 0 {
			fmt.Println("Wrong Host Port")
			return
		}
		machineStats := make(map[string]model.MachineStats)

		diagnose := diagnostics.Init()
		diagnose.RunDiagnostics(nodesTableInfo, masterSlaveIpMap)
		for _, node := range nodesTableInfo {
			updateMachineStats(machineStats, node)
		}
		display.DisplayNodeStats(nodesTableInfo, masterSlaveIpMap, diagnose)
		display.DisplayMachineStats(machineStats, totalMasters, diagnose)

		diagnose.Print()
		// error on eviction policy, no slave, linux overcommit etc.
		var tmp string
		fmt.Scanln(&tmp)
	}
}

// returns model.NodesInfo by getting data from redis cluster nodes command and redis info command
func GetClusterNodesInfo(host, port string, clusterInfo bool) (model.NodesInfo, map[string][]string, int) {
	var nodes []model.NodeInfo
	//fmt.Println(utils.ExecCmd("/usr/bin/which a"))
	//redisCliExists := utils.ExecCmd("/usr/bin/which redis-cli")
	//if redisCliExists == "" {
	//	panic("redis-cli not found")
	//	os := utils.ExecCmd("uname")
	//	arch := utils.ExecCmd("uname -m")
	//	if os == "linux" && arch == "x86_64"{
	//
	//	}
	//	// TODO handle when redis-cli is not there
	//}
	cmd := "redis-cli -h " + host + " -p " + port + " cluster nodes"
	out, _ := exec.Command("sh", "-c", cmd).Output()
	clusterStatus := string(out)
	nodeDetail := strings.Split(clusterStatus, "\n")
	masterSlaveIpMap := make(map[string][]string)
	totalMasters := 0
	var machineIp string
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
		if strings.Contains(nodeDetailList[2], "myself") {
			machineIp = hostPort[0]
		}

		if strings.Contains(nodeDetailList[2], "slave") {
			masterSlaveIpMap[nodeDetailList[3]] = append(masterSlaveIpMap[nodeInfo.MasterId], IpHostMap[nodeInfo.Ip]+":"+nodeInfo.Port)
			nodeInfo.Type = model.SLAVE
		} else {
			nodeInfo.Type = model.MASTER
			nodeInfo.HashSlot = nodeDetailList[8]
			slotRange := strings.Split(nodeInfo.HashSlot, "-")
			nodeInfo.HashSlotStart, _ = strconv.Atoi(slotRange[0])
			nodeInfo.HashSlotEnd, _ = strconv.Atoi(slotRange[1])
			totalMasters += 1
		}
		nodeInfo.Host = IpHostMap[nodeInfo.Ip]
		nodes = append(nodes, nodeInfo)
	}
	var nodesTableInfo model.NodesInfo
	if clusterInfo == false {
		host := strings.Trim(utils.ExecCmd("/bin/hostname"), "\t\r\n")
		var machineNodes []model.NodeInfo
		for _, nd := range nodes {
			if nd.Ip == machineIp {
				nd.Host = host
				machineNodes = append(machineNodes, nd)
			}
		}
		nodesTableInfo = getRedisInfo(machineNodes)
	} else {
		nodesTableInfo = getRedisInfo(nodes)
		sort.Sort(nodesTableInfo)
	}
	return nodesTableInfo, masterSlaveIpMap, totalMasters
}

// returns model.NodesInfo by getting data from redis cluster nodes command and redis info command
func GetNodeInfo(hostInput, portInput string) model.NodesInfo {
	var nodes []model.NodeInfo
	var nodeInfo model.NodeInfo
	ports := strings.Split(portInput, ",")
	host := strings.Trim(utils.ExecCmd("/bin/hostname"), "\t\n\r")
	for _, port := range ports {
		nodeInfo.Ip, nodeInfo.Port = hostInput, port
		nodeInfo.Host = host
		nodes = append(nodes, nodeInfo)
	}
	nodesTableInfo := getRedisInfo(nodes)
	return nodesTableInfo
}

// spawns workers to fetch the redis info concurrently
func getRedisInfo(nodes []model.NodeInfo) model.NodesInfo {
	redisInfo := make(chan model.NodeInfo, len(nodes))
	nodeList := make(chan model.NodeInfo, len(nodes))

	for w := 1; w <= 20; w++ {
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

// redis worker to extract data from redis info and map it to model.NodeInfo structure
func redisInfoWorker(nodeList chan model.NodeInfo, redisInfo chan model.NodeInfo) {
	for node := range nodeList {
		// TODO handle if node down
		isMaxMemoryKeyPresent := false
		evictionPolicySet := false

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
			} else if strings.HasPrefix(line, "role:") {
				info.Role = strings.Split(line, ":")[1]
			} else if strings.HasPrefix(line, "process_id:") {
				info.Pid, _ = strconv.Atoi(strings.Split(line, ":")[1])
			} else if strings.HasPrefix(line, "uptime_in_seconds:") {
				info.Uptime, _ = strconv.Atoi(strings.Split(line, ":")[1])
			} else if strings.HasPrefix(line, "maxmemory:") {
				info.MaxMemory, _ = strconv.Atoi(strings.Split(line, ":")[1])
				isMaxMemoryKeyPresent = true
			} else if strings.HasPrefix(line, "maxmemory_policy:") {
				info.EvictionPolicy = strings.Split(line, ":")[1]
				evictionPolicySet = true
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
		if !isMaxMemoryKeyPresent {
			if info.MaxMemory == 0 {
				getMaxMemory := "redis-cli -h " + node.Ip + " -p " + node.Port + " config get maxmemory"
				maxMemoryOutput, _ := exec.Command("sh", "-c", getMaxMemory).Output()
				info.MaxMemory, _ = strconv.Atoi(strings.Split(string(maxMemoryOutput), "\n")[1])
			}
		}
		if !evictionPolicySet {
			if info.EvictionPolicy == "" {
				evictionPolicy, _ := exec.Command("sh", "-c", "redis-cli -h "+node.Ip+" -p "+
					node.Port+" config get maxmemory-policy").Output()
				info.EvictionPolicy = strings.Split(string(evictionPolicy), "\n")[1]
			}
		}
		nodeDetail.CopyFrom(node)
		if info.MaxMemory != 0 {
			info.MemoryLeft = 100 * (float64(info.MaxMemory-info.UsedMemory) / float64(info.MaxMemory))
		}
		nodeDetail.RedisInfo = info
		redisInfo <- nodeDetail
	}
}

// updates machineStats using nodeInfo
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
