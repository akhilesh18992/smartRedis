package diagnostics

import (
	"smartRedis/model"
	"smartRedis/color"
	"fmt"
	"strings"
)

type diagnostics struct {
	error	[]string
	warning	[]string
}

func Init() *diagnostics {
	return &diagnostics{}
}

func (d *diagnostics) RunDiagnostics(nodesTableInfo model.NodesInfo, masterSlaveIpMap map[string][]string)  {
	var masterSlaveOnSameMachine bool
	for _, node := range nodesTableInfo {

		if IsMasterSlaveOnSameMachine(masterSlaveIpMap[node.NodeId], node.Ip) {
			masterSlaveOnSameMachine = true
		}
	}

	if masterSlaveOnSameMachine {
		d.Error("CLUSTER ERROR: Master slave on same machine")
	}
}

func (d *diagnostics) Error(error string)  {
	d.error = append(d.error, error)
}

func (d *diagnostics) Print()  {
	diagnosticResult := color.BBlue("\nSmartRedis Diagnosis") + "\n"
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