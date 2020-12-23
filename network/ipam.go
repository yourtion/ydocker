package network

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

const defaultAllocatorPath = "/var/run/ydocker/network/ipam/subnet.json"

type IPAM struct {
	SubnetAllocatorPath string
	Subnets             *map[string]string
}

var ipAllocator = &IPAM{
	SubnetAllocatorPath: defaultAllocatorPath,
}

func (ipam *IPAM) load() error {
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	data, err := ioutil.ReadFile(ipam.SubnetAllocatorPath)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, ipam.Subnets); err != nil {
		log.Errorf("Error dump allocation info, %v", err)
		return err
	}
	return nil
}

func (ipam *IPAM) dump() error {
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(ipamConfigFileDir, 0644)
		}
		return err
	}
	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(ipam.SubnetAllocatorPath, ipamConfigJson, 0644); err != nil {
		return err
	}
	return nil
}

// 从指定的 subnet 网段中分配 IP 地址
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	// 存放网段中地址分配信息的数组
	ipam.Subnets = &map[string]string{}
	// 从文件中加载已经分配的网段信息
	err = ipam.load()
	if err != nil {
		log.Errorf("Error dump allocation info, %v", err)
	}
	_, subnet, _ = net.ParseCIDR(subnet.String())
	one, size := subnet.Mask.Size()
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-one))
	}

	for c := range (*ipam.Subnets)[subnet.String()] {
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			ipAlloc := []byte((*ipam.Subnets)[subnet.String()])
			ipAlloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipAlloc)
			ip = subnet.IP
			for t := uint(4); t > 0; t -= 1 {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			ip[3] += 1
			break
		}
	}
	err = ipam.dump()
	return
}

// 从指定的 subnet 网段中释放掉指定的 IP 地址
func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	err := ipam.load()
	if err != nil {
		log.Errorf("Error dump allocation info, %v", err)
	}

	c := 0
	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}

	ipAlloc := []byte((*ipam.Subnets)[subnet.String()])
	ipAlloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipAlloc)

	ipam.dump()
	return nil
}
