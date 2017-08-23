package display

import (
	"github.com/pkg/errors"
	"smartRedis/color"
	"smartRedis/model"
	"smartRedis/table"
	"smartRedis/utils"
	"strconv"
	"strings"
	"smartRedis/diagnostics"
)

func DisplayNodeStats(nodesTableInfo model.NodesInfo, masterSlaveIpMap map[string][]string) (err error) {
	tw := table.Init()
	var masterSlaveOnSameMachine bool
	var host string
	defaultColor := color.GREEN
	count := 1
	for _, node := range nodesTableInfo {
		if node.Type == model.SLAVE {
			continue
		}
		cacheMiss := strconv.FormatFloat((float64(node.Hits)/float64(node.Miss + node.Hits))*100, 'f', 3, 64) + "%"
		colorCode := color.GREEN
		if diagnostics.IsMasterSlaveOnSameMachine(masterSlaveIpMap[node.NodeId], node.Ip) {
			colorCode = color.RED
			masterSlaveOnSameMachine = true
		}
		if node.Host != "" {
			host = node.Host
		} else {
			host = node.Ip
		}
		tw.AppendRecord([]table.Record{
			{strconv.Itoa(count), defaultColor},
			{host, defaultColor},
			{node.Port, defaultColor},
			{utils.ReadableMemory(node.UsedMemory), defaultColor},
			{cacheMiss, defaultColor},
			{strings.Join(masterSlaveIpMap[node.NodeId], ","), colorCode},
			{node.HashSlot, defaultColor},
			{utils.ReadableMemory(node.UsedMemoryPeak), defaultColor},
		})
		count += 1
	}
	tw.SetHeader([]string{"Id", "Host(Master)", "Port", "Master Data Size", "Cache Hit Ratio", "Slave Node", "Slot", "Peak Mem Used"})
	tw.Render()
	if masterSlaveOnSameMachine {
		err = errors.New("CLUSTER ERROR: Master slave on same machine")
	}
	return
}

func DisplayMachineStats(machineStats map[string]model.MachineStats, totalMaster int) (err error) {
	unbalancedCluster := false
	avgMaster := totalMaster / len(machineStats)
	defaultColor := color.GREEN
	t := table.Init()
	t.SetHeader([]string{"Machine", "Total Space Used", "Ops Per second", "Network(kbps)", "Master", "Slave"})
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

		t.AppendRecord([]table.Record{
			{host, defaultColor},
			{spaceUsed, defaultColor},
			{strconv.Itoa(stats.OpsPerSec), defaultColor},
			{strconv.FormatFloat(stats.NetworkBandwidth, 'f', 2, 64), defaultColor},
			{strconv.Itoa(stats.Master), colorCode},
			{strconv.Itoa(stats.Slave), defaultColor}})
	}
	t.Render()
	if unbalancedCluster {
		err = errors.New("CLUSTER UNBALANCED: masters non uniformly distributed across cluster")
	}
	return
}
