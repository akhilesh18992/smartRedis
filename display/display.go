package display

import (
	"smartRedis/color"
	"smartRedis/diagnostics"
	"smartRedis/model"
	"smartRedis/table"
	"smartRedis/utils"
	"strconv"
	"strings"
)

func isMaxMemorySet(nodesTable model.NodesInfo) bool {
	for _, node := range nodesTable {
		if node.MaxMemory != 0 {
			return true
		}
	}
	return false
}

func DisplayNodeStats(nodesTableInfo model.NodesInfo, masterSlaveIpMap map[string][]string, d *diagnostics.Diagnostics) {
	var host string
	defaultColor := color.GREEN
	count := 1
	tw := table.Init()
	showMaxMemoryColumn := isMaxMemorySet(nodesTableInfo)
	nonExpiryKeys := false
	maxMemoryAlert := false
	for _, node := range nodesTableInfo {
		if node.Type == model.SLAVE {
			continue
		}
		cacheMiss := strconv.FormatFloat((float64(node.Hits)/float64(node.Miss+node.Hits))*100, 'f', 3, 64) + "%"
		colorCode := color.GREEN
		if diagnostics.IsMasterSlaveOnSameMachine(masterSlaveIpMap[node.NodeId], node.Ip) {
			colorCode = color.RED
			//masterSlaveOnSameMachine = true
		}
		if node.Host != "" {
			host = node.Host
		} else {
			host = node.Ip
		}
		if showMaxMemoryColumn {
			maxMemoryColor := defaultColor
			if node.MemoryLeft < 10 {
				maxMemoryColor = color.RED
				maxMemoryAlert = true
			}
			tw.AppendRecord([]table.Record{
				{strconv.Itoa(count), defaultColor},
				{host, defaultColor},
				{node.Port, defaultColor},
				{utils.ReadableMemory(node.UsedMemory), defaultColor},
				{strconv.FormatFloat(node.MemoryLeft, 'f', 3, 64) + "%", maxMemoryColor},
				{cacheMiss, defaultColor},
				{strings.Join(masterSlaveIpMap[node.NodeId], ","), colorCode},
				{node.HashSlot, defaultColor},
				{utils.ReadableMemory(node.UsedMemoryPeak), defaultColor},
				{strconv.Itoa(node.NonExpiryKeys), color.RED},
			})
		} else {
			tw.AppendRecord([]table.Record{
				{strconv.Itoa(count), defaultColor},
				{host, defaultColor},
				{node.Port, defaultColor},
				{utils.ReadableMemory(node.UsedMemory), defaultColor},
				{cacheMiss, defaultColor},
				{strings.Join(masterSlaveIpMap[node.NodeId], ","), colorCode},
				{node.HashSlot, defaultColor},
				{utils.ReadableMemory(node.UsedMemoryPeak), defaultColor},
				{strconv.Itoa(node.NonExpiryKeys), color.RED},
			})
		}
		if node.NonExpiryKeys > 0 {
			nonExpiryKeys = true
		}
		count += 1
	}
	if showMaxMemoryColumn {
		tw.SetHeader([]string{"Id", "Host(Master)", "Port", "Master Data Size", "Memory Left", "Cache Hit Ratio", "Slave Node", "Slot", "Peak Mem Used", "Non Expiry Keys"})
	} else {
		tw.SetHeader([]string{"Id", "Host(Master)", "Port", "Master Data Size", "Cache Hit Ratio", "Slave Node", "Slot", "Peak Mem Used", "Non Expiry Keys"})
	}
	tw.Render()
	if !showMaxMemoryColumn {
		d.Error("CLUSTER ERROR: Max Memory not set on some machine")
	}
	if nonExpiryKeys {
		d.Error("CLUSTER ERROR: Non expiry keys present in cluster")
	}
	if maxMemoryAlert {
		d.Error("CLUSTER ERROR: Less than 10% memory left")
	}
}

func DisplayMachineStats(machineStats map[string]model.MachineStats, totalMaster int, d *diagnostics.Diagnostics) {
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
		d.Error("CLUSTER UNBALANCED: masters non uniformly distributed across cluster")
	}
}
