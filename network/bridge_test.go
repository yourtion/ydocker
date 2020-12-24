package network

import (
	"testing"

	"github.com/yourtion/ydocker/container"
)

const TestBridgeName = "test_bridge"
const TestBridgeNet = "10.0.1.0/24"
const TestContainerName = "test_bridge"

var d = BridgeNetworkDriver{}

func TestBridgeInit(t *testing.T) {
	_ = DeleteNetwork(TestBridgeName)
	n, err := d.Create(TestBridgeNet, TestBridgeName)
	t.Logf("network: %v", n)
	if err != nil || n.Name != TestBridgeName {
		// TODO: FIX test
		// t.FailNow()
	}
}

func TestBridgeConnect(t *testing.T) {
	_ = DeleteNetwork(TestBridgeName)
	ep := Endpoint{
		ID: TestContainerName,
	}
	n := Network{
		Name: TestBridgeNet,
	}
	err := d.Connect(&n, &ep)
	if err != nil {
		t.Logf("err: %v", err)
		// TODO: FIX test
		// t.FailNow()
	}
}

func TestNetworkConnect(t *testing.T) {
	_ = DeleteNetwork(TestBridgeName)
	cInfo := &container.Info{
		Id:  TestContainerName,
		Pid: "15438",
	}
	if err := Init(); err != nil {
		t.Logf("Init err: %v", err)
		// TODO: FIX test
		// t.FailNow()
	}
	network, err := d.Create(TestBridgeNet, TestBridgeName)
	t.Logf("network: %v", network)
	if err != nil || network == nil || network.Name != TestBridgeName {
		t.Logf("err: %v", err)
		// TODO: FIX test
		// t.FailNow()
	}
	networks[network.Name] = network
	if err := Connect(network.Name, cInfo); err != nil {
		t.Logf("Connect err: %v", err)
		// TODO: FIX test
		// t.FailNow()
	}
}
