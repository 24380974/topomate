package link

import (
	"bytes"
	"fmt"
	"log"

	"github.com/digitalocean/go-openvswitch/ovs"
	"github.com/rahveiz/topomate/config"
	"github.com/rahveiz/topomate/internal/ovsdocker"
	"github.com/rahveiz/topomate/utils"
)

func CreateBridge(name string) {

	c := ovs.New(ovs.Sudo())

	if err := c.VSwitch.AddBridge(name); err != nil {
		log.Fatalf("failed to add bridge: %v", err)
	}

}

func DeleteBridge(name string) {
	c := ovs.New(ovs.Sudo())

	if err := c.VSwitch.DeleteBridge(name); err != nil {
		log.Fatalf("failed to delete bridge: %v", err)
	}
}

// AddPortToContainer links a container to an OVS bridge, creating an interface on the container network namespace
// using a veth pair.
func AddPortToContainer(brName, ifName, containerName string,
	settings ovsdocker.PortSettings, hostIf *ovsdocker.OVSInterface,
	bridge bool) {
	c := ovsdocker.New(containerName)
	if err := c.AddPort(brName, ifName, settings, hostIf, bridge); err != nil {
		utils.Fatalln("AddPort:", err)
	}
}

// DelPortFromContainer removes an OVS port from a container
func DelPortFromContainer(brName, ifName, containerName string) {
	out, err := utils.ExecSudo(
		"ovs-docker",
		"del-port",
		brName,
		ifName,
		containerName,
	).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		log.Fatalf("error using ovs-docker: %s\n", err)
	}
}

// ClearPortsFromContainer removes all OVS ports from a container
func ClearPortsFromContainer(brName, containerName string) {
	out, err := utils.ExecSudo(
		"ovs-docker",
		"del-ports",
		brName,
		containerName,
	).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		log.Fatalf("error using ovs-docker: %s\n", err)
	}
}

func AddFlow(brName, containerA, ifA, containerB, ifB string) {
	portA, _ := ovsdocker.GetOFPort(containerA, ifA)
	portB, _ := ovsdocker.GetOFPort(containerB, ifB)
	var stderr bytes.Buffer
	cmd := utils.ExecSudo(
		"ovs-ofctl",
		"add-flow", brName,
		"in_port="+portA+",actions=output:"+portB,
	)
	cmd.Stderr = &stderr
	if config.VFlag {
		fmt.Println(cmd.String())
	}
	err := cmd.Run()
	if err != nil {
		utils.Fatalln("AddFlow:", string(stderr.Bytes()), err)
	}
	cmd = utils.ExecSudo(
		"ovs-ofctl",
		"add-flow", brName,
		"in_port="+portB+",actions=output:"+portA,
	)
	cmd.Stderr = &stderr
	if config.VFlag {
		fmt.Println(cmd.String())
	}
	err = cmd.Run()
	if err != nil {
		utils.Fatalln("AddFlow:", string(stderr.Bytes()), err)
	}

}
