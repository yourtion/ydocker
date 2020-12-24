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

// 地址分配
type IPAM struct {
	// 分配文件存放位置
	SubnetAllocatorPath string
	// 网段和位图算法的数组 map, key 是网段， value 是分配的位图数组
	Subnets *map[string]string
}

// 初始化一个 IPAM 的对象，默认使用 /var/run/ydocker/network/ipam/subnet.json 作为分配信息存储位置
var ipAllocator = &IPAM{
	SubnetAllocatorPath: defaultAllocatorPath,
}

// 加载网段地址分配信息
func (ipam *IPAM) load() error {
	// 检查存储文件状态，如果不存在，则说明之前没有分配，则不需要加载
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	// 打开并读取存储文件
	data, err := ioutil.ReadFile(ipam.SubnetAllocatorPath)
	if err != nil {
		return err
	}
	// 将文件中的内容反序列化出 IP 的分配信息
	if err = json.Unmarshal(data, ipam.Subnets); err != nil {
		log.Errorf("Error dump allocation info, %v", err)
		return err
	}
	return nil
}

// 存储网段地址分配信息
func (ipam *IPAM) dump() error {
	// 检查存储文件所在文件夹是否存在，如果不存在则创建
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(ipamConfigFileDir, 0644)
		}
		return err
	}
	// 序列化 IPAM 对象到 json 串
	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}
	// 将序列化后的 json 串写入到配置文件中
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
	// net.IPNet.Mask.Size() 函数会返回网段的子网掩码的总长度和网段前面的固定位的长度
	// 比如 127.0.0.0/8 网段的子网掩码是 255.0.0.0
	// 那么 subnet.Mask.Size() 的返回值就是前面 255 所对应的位数和总位数，即 8 和 24
	one, size := subnet.Mask.Size()
	// 如果之前没有分配过这个网段，则初始化网段的分配配置
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		// 用 "0" 填满这个网段的配置， 1 << uint8(size - one) 表示这个网段中有多少个可用地址
		// size - one 是子网掩码后面的网络位数， 2^(size - one) 表示网段中的可用 IP 数
		// 而 2^(size - one) 等价于 1 << uint8(size - one)
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-one))
	}
	// 遍历网段的位图数组
	for c := range (*ipam.Subnets)[subnet.String()] {
		// 找到数组中为 "0" 的项和数组序号，即可以分配的 IP
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			// 设置这个为 "0" 的序号值为 "1" ，即分配这个 IP
			ipAlloc := []byte((*ipam.Subnets)[subnet.String()])
			// Go 的字符串，创建之后就不能修改，所以通过转换成 byte 数组，修改后再转换成字符串赋值
			ipAlloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipAlloc)
			// 这里的 IP 为初始 IP ，比如对于网段 192.168.0.0/16，这里就是 192.168.0.0
			ip = subnet.IP
			/*
				通过网段的 IP 与上面的偏移相加计算出分配的 IP 地址，由于 IP 地址是 uint 的一个数组，
				需要通过数组中的每一项加所窝耍的值，比如网段是 172.16.0.0/12，数组序号是 65555.
				那么在 [172,16,0,0] 上依次加 [uint8(65555 >> 24), uint8(65555 >> 16), uint8(65555 >> 8), uint8 (65555 >> 0)]
				即 [0,1,0,19]，那么获得的 IP 就是 172.17.0.19
			*/
			for t := uint(4); t > 0; t -= 1 {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			// 由于此处 IP 是从 1 开始分配的，所以最后再加 1，最终得到分配的 IP 是 172.17.0.20
			ip[3] += 1
			break
		}
	}
	// 通过调用 dump() 将分配结果保存到文件中
	err = ipam.dump()
	return
}

// 从指定的 subnet 网段中释放掉指定的 IP 地址
func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}
	_, subnet, _ = net.ParseCIDR(subnet.String())
	// 从文件中加载网段的分配信息
	err := ipam.load()
	if err != nil {
		log.Errorf("Error dump allocation info, %v", err)
	}
	// 计算 IP 地址在网段位图数组中的索引位置
	c := 0
	// 将 IP 地址转换成 4 个字节的表示方式
	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		// 与分配 IP 相反，释放 IP 获得索引的方式是 IP 地址的每一位相减之后分别左移将对应的数值加到索引上
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}
	// 将分配的位图数组中索引位置的值置为 0
	ipAlloc := []byte((*ipam.Subnets)[subnet.String()])
	ipAlloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipAlloc)
	// 保存释放掉 IP 之后的网段 IP 分配信息
	_ = ipam.dump()
	return nil
}
