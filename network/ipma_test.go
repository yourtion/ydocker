package network

import (
	"net"
	"testing"
)

var ipAllocatorTest = &IPAM{
	SubnetAllocatorPath: "/tmp/subnet.json",
}

func TestAllocate(t *testing.T) {
	// 在 192.168.0.0/24 网段中分配 IP
	_, ipNet, _ := net.ParseCIDR("192.168.0.1/24")
	ip, err := ipAllocatorTest.Allocate(ipNet)
	t.Logf("alloc ip: %v", ip)
	if err != nil || ip == nil {
		t.FailNow()
	}
	if ip.String() != "192.168.0.1" {
		t.FailNow()
	}
}

func TestRelease(t *testing.T) {
	// 在 192.168.0.0/24 网段中释放刚才分配的 192.168.0.1 的 IP
	ip, ipNet, _ := net.ParseCIDR("192.168.0.1/24")
	err := ipAllocatorTest.Release(ipNet, &ip)
	if err != nil {
		t.FailNow()
	}
}
