package model

type MachineStats struct {
	Ip               string
	Hostname         string
	Memory           int
	RedisMemory      int
	OpsPerSec        int
	NetworkBandwidth float64
	Master           int
	Slave            int
	RedisNodes       int
}
