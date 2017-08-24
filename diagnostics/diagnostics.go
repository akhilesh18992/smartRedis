package diagnostics

import (
	"smartRedis/model"
	"smartRedis/color"
	"fmt"
	"strings"
)

type Diagnostics struct {
	version string
	evictionPolicy string
	error	[]string
	warning	[]string
}

func Init() *Diagnostics {
	return &Diagnostics{}
}

func (d *Diagnostics) RunDiagnostics(nodesTableInfo model.NodesInfo, masterSlaveIpMap map[string][]string)  {
	var masterSlaveOnSameMachine bool
	var clusterVersion, clusterEvictionPolicy string
	inconsistentRedisVersion := false
	inconsistentEvictionPolicyVersion := false
	for _, node := range nodesTableInfo {
		if clusterVersion == "" {
			clusterVersion = node.Version
		} else if clusterVersion != node.Version {
			inconsistentRedisVersion = true
		}
		if clusterEvictionPolicy == "" {
			clusterEvictionPolicy = node.EvictionPolicy
		} else if clusterEvictionPolicy != node.EvictionPolicy {
			inconsistentEvictionPolicyVersion = true
		}
		if IsMasterSlaveOnSameMachine(masterSlaveIpMap[node.NodeId], node.Ip) {
			masterSlaveOnSameMachine = true
		}
	}
	d.version = clusterVersion
	d.evictionPolicy = clusterEvictionPolicy
	if masterSlaveOnSameMachine {
		d.Error("CLUSTER ERROR: Master slave on same machine")
	}
	if inconsistentRedisVersion {
		d.Warning("CLUSTER WARNING: Different redis version found across the cluster")
	}
	if inconsistentEvictionPolicyVersion {
		d.Warning("CLUSTER WARNING: Different eviction policy found across the cluster")
	} else if d.evictionPolicy == "noeviction" {
		d.Error("CLUSTER ERROR: Noeviction policy set on some machines")
	}
}

func (d *Diagnostics) Error(error string)  {
	d.error = append(d.error, error + "\n")
}

func (d *Diagnostics) Warning(warning string)  {
	d.warning = append(d.warning, warning + "\n")
}

func (d *Diagnostics) Print()  {
	diagnosticResult := color.BBlue("\nSmartRedis Diagnosis") + "\n"
	if d.version != "" {
		diagnosticResult += color.Green("Redis Version: " + d.version) + "\n"
	}
	if d.evictionPolicy != "" {
		diagnosticResult += color.Green("Eviction Policy: " + d.evictionPolicy) + "\n"
	}
	for _, err := range d.error {
		diagnosticResult += color.BRed(err)
	}
	for _, warning := range d.warning {
		diagnosticResult += color.BYellow(warning)
	}
	fmt.Print(diagnosticResult)
}

func IsMasterSlaveOnSameMachine(slaveIps []string, masterId string) (masterSlaveOnSameMachine bool) {
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