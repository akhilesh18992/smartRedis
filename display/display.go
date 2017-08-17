package display

import (
	"fmt"
	"github.com/pkg/errors"
	"smartRedis/color"
	"smartRedis/model"
	"smartRedis/table"
	"smartRedis/utils"
	"strconv"
	"strings"
)

func DisplayNodeStats(nodesTableInfo model.NodesInfo, masterSlaveIpMap map[string][]string) (err error) {
	tw := table.Init()
	var masterSlaveOnSameMachine bool
	var host string
	for _, node := range nodesTableInfo {
		if node.Type == model.SLAVE {
			continue
		}
		cacheMiss := strconv.FormatFloat((float64(node.Miss)/float64(node.Hits))*100, 'f', 3, 64)
		colorCode := color.GREEN
		if isMasterSlaveOnSameMachine(masterSlaveIpMap[node.NodeId], node.Ip) {
			colorCode = color.RED
			masterSlaveOnSameMachine = true
		}
		if node.Host != "" {
			host = node.Host
		} else {
			host = node.Ip
		}
		tw.Append([]string{host, node.Port, utils.ReadableMemory(node.UsedMemory), utils.ReadableMemory(node.UsedMemoryPeak), cacheMiss,
			strings.Join(masterSlaveIpMap[node.NodeId], ","), node.HashSlot}, colorCode)
	}
	tw.SetHeader([]string{"Host", "Port", "Data Size", "Peak Mem Used", "Cache Miss", "Slave Node", "Slot"})
	tw.Render()
	if masterSlaveOnSameMachine {
		err = errors.New("CLUSTER ERROR: Master slave on same machine")
	}
	return
}

func DisplayMachineStats(machineStats map[string]model.MachineStats, totalMaster int) (err error) {
	unbalancedCluster := false
	avgMaster := totalMaster / len(machineStats)
	t := table.Init()
	t.SetHeader([]string{"Machine", "Space Used", "Ops Per second", "Network(kbps)", "Master", "Slave"})
	fmt.Println("\n\nTotal masters: " + strconv.Itoa(totalMaster))
	var host string
	for ip, stats := range machineStats {
		colorCode := color.GREEN
		if stats.Master > avgMaster {
			colorCode = color.RED
			unbalancedCluster = true
		}
		var spaceUsed string
		if stats.Memory == 0 {
			spaceUsed = utils.ReadableMemory(stats.RedisMemory)
		} else {
			spaceUsed = strconv.FormatFloat((float64(stats.RedisMemory)/float64(stats.Memory))*100, 'f', 2, 64) + "%"
		}
		if stats.Hostname != "" {
			host = stats.Hostname
		} else {
			host = ip
		}

		t.Append([]string{host, spaceUsed, strconv.Itoa(stats.OpsPerSec), strconv.FormatFloat(stats.NetworkBandwidth, 'f', 2, 64),
			strconv.Itoa(stats.Master), strconv.Itoa(stats.Slave)}, colorCode)
	}
	t.Render()
	if unbalancedCluster {
		err = errors.New("CLUSTER UNBALANCED: masters non uniformly distributed across cluster")
	}
	return
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
