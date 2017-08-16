package model

const SLAVE = "SLAVE"
const MASTER = "MASTER"

type NodeInfo struct {
	NodeId        string
	Ip            string
	Port          string
	MasterId      string
	HashSlot      string
	HashSlotStart int
	HashSlotEnd   int
	Type          string
	RedisInfo
}

type RedisInfo struct {
	Version                 string
	Mode                    string
	Pid                     int
	Uptime                  int // secs
	MaxMemory               int
	EvictionPolicy          string
	InstantaneousOpsPerSec  int
	InstantaneousInputKbps  float64
	InstantaneousOutputKbps float64
	Hits                    int
	Miss                    int
	UsedMemory              int
	UsedMemoryPeak          int
	NonExpiryKeys           int
	SystemMemory            int
}

type NodesInfo []NodeInfo

func (slice NodesInfo) Len() int {
	return len(slice)
}

func (slice NodesInfo) Less(i, j int) bool {
	if slice[i].Ip < slice[j].Ip || (slice[i].Ip == slice[j].Ip && slice[i].Type < slice[j].Type) ||
		(slice[i].Ip == slice[j].Ip && slice[i].Type == slice[j].Type && slice[i].HashSlotStart < slice[j].HashSlotStart) {
		return true
	}
	return false
}

func (slice NodesInfo) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (node *NodeInfo) CopyFrom(nodeInfo NodeInfo) {
	node.Ip = nodeInfo.Ip
	node.Port = nodeInfo.Port
	node.NodeId = nodeInfo.NodeId
	node.Type = nodeInfo.Type
	node.MasterId = nodeInfo.MasterId
	node.HashSlot = nodeInfo.HashSlot
	node.HashSlotStart = nodeInfo.HashSlotStart
	node.HashSlotEnd = nodeInfo.HashSlotEnd
	node.RedisInfo = nodeInfo.RedisInfo
}
