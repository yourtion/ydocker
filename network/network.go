package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	"github.com/yourtion/ydocker/container"
)

var (
	defaultNetworkPath = "/var/run/ydocker/network/network/"
	drivers            = map[string]Driver{}
	networks           = map[string]*Network{}
)

// 网络端点
type Endpoint struct {
	ID          string           `json:"id"`  // 网络ID
	Device      netlink.Veth     `json:"dev"` // Veth 设备
	IPAddress   net.IP           `json:"ip"`  // IP 地址
	MacAddress  net.HardwareAddr `json:"mac"` // MAC 地址
	Network     *Network         // 连接的容器和网络
	PortMapping []string         // 端口映射
}

// 网络
type Network struct {
	Name    string     // 网络名
	IpRange *net.IPNet // 地址段
	Driver  string     // 网络驱动名
}

// 网络驱动
type Driver interface {
	// 驱动名
	Name() string
	// 创始网络
	Create(subnet string, name string) (*Network, error)
	// 删除网络
	Delete(network Network) error
	// 连接容#苦网络端点到网络
	Connect(network *Network, endpoint *Endpoint) error
	// 从网络上移除容器网络端点
	Disconnect(network Network, endpoint *Endpoint) error
}

// 保存网络信息
func (nw *Network) dump(dumpPath string) error {
	// 检查保存的目录是否存在，不存在则创建
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(dumpPath, 0644)
		} else {
			return err
		}
	}
	// 通过 json 的库序列化网络对象到 json 的字符串
	nwJson, err := json.Marshal(nw)
	if err != nil {
		logrus.Errorf("Marshal json error：%v", err)
		return err
	}
	// 保存的文件名是网络的名字
	nwPath := path.Join(dumpPath, nw.Name)
	// 将网络配置的 json 字符串写入到文件中
	if err := ioutil.WriteFile(nwPath, nwJson, 0644); err != nil {
		logrus.Errorf("WriteFile error：%v", err)
		return err
	}
	return nil
}

// 删除网络信息
func (nw *Network) remove(dumpPath string) error {
	// 网络对应的配置文件，即配置目录下的网络名文件
	// 检查文件状态，如果文件己经不存在就直接返回
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	// 调用 os.Remove 删除这个网络对应的配置文件
	return os.Remove(path.Join(dumpPath, nw.Name))
}

// 加载网络信息
func (nw *Network) load(dumpPath string) error {
	// 从配置文件中读取网络的配置 json 字符串
	nwJson, err := ioutil.ReadFile(dumpPath)
	if err != nil {
		return err
	}
	// 通过 json 字符串反序列化出网络
	if err = json.Unmarshal(nwJson, nw); err != nil {
		logrus.Errorf("Error load nw info: %v", err)
		return err
	}
	return nil
}

func Init() error {
	// 加载网络驱动
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver
	// 判断网络的配置目录是否存在，不存在则创建
	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(defaultNetworkPath, 0644)
		} else {
			return err
		}
	}

	// 检查网络配置目录中的所有文件
	// filepath.Walk 函数会遍历指定的 path 目录并执行第二个参数中的函数指针去处理目录下的每一个文件
	err := filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		// 如果是目录则跳过
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}
		// 加载文件名作为网络名
		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwName,
		}
		// 加载网络的配置信息
		if err := nw.load(nwPath); err != nil {
			logrus.Errorf("error load network: %s", err)
		}
		// 将网络的配且信息加入到 networks 字典中
		networks[nwName] = nw
		return nil
	})
	logrus.Infof("networks: %v", networks)
	return err
}

// 创建网络
func CreateNetwork(driver, subnet, name string) error {
	// ParseCIDR 是 Golang net 包的函数，功能是将网段的字符串转换成 net.IPNet 的对象
	_, cidr, _ := net.ParseCIDR(subnet)
	// 通过 IPAM 分配网关 IP，获取到网段中第一个 IP 作为网关的 IP
	ip, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = ip
	// 调用指定的网络驱动创建网络，这里的 drivers 字典是各个网络驱动的实例字典，通过调用网络驱动的 Create 方法创建网络
	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}
	// 保存网络信息，将网络的信息保存在文件系统中，以便查询和在网络上连接网络端点
	return nw.dump(defaultNetworkPath)
}

// 展示网络列表
func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprint(w, "NAME\tIpRange\tDriver\n")
	// 遍历网络信息
	for _, nw := range networks {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", nw.Name, nw.IpRange.String(), nw.Driver)
	}
	// 输出到标准输出
	if err := w.Flush(); err != nil {
		logrus.Errorf("Flush error %v", err)
	}
}

// 删除网络
func DeleteNetwork(networkName string) error {
	// 查找网络是否存在
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no Such Network: %s", networkName)
	}
	// 调用 IPAM 的实例 ipAllocator 释放网络网关的 IP
	if err := ipAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return fmt.Errorf("error Remove Network gateway ip: %s", err)
	}
	// 调用网络驱动删除网络创建的设备与配置
	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("error Remove Network DriverError: %s", err)
	}
	// 从网络的配直目录中删除该网络对应的配置文件
	return nw.remove(defaultNetworkPath)
}

func enterContainerNetns(enLink *netlink.Link, cinfo *container.Info) func() {
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		logrus.Errorf("error get container net namespace, %v", err)
	}

	nsFD := f.Fd()
	runtime.LockOSThread()

	// 修改 veth peer 另外一端移到容器的namespace中
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		logrus.Errorf("error set link netns , %v", err)
	}

	// 获取当前的网络namespace
	origns, err := netns.Get()
	if err != nil {
		logrus.Errorf("error get current netns, %v", err)
	}

	// 设置当前进程到新的网络namespace，并在函数执行完成之后再恢复到之前的namespace
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		logrus.Errorf("error set netns, %v", err)
	}
	return func() {
		netns.Set(origns)
		origns.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

func configEndpointIpAddressAndRoute(ep *Endpoint, cinfo *container.Info) error {
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}

	defer enterContainerNetns(&peerLink, cinfo)()

	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress

	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("%v,%s", ep.Network, err)
	}

	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}

	if err = setInterfaceUP("lo"); err != nil {
		return err
	}

	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")

	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}

	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}

	return nil
}

// 配置容器到宿主机的端口映射，
func configPortMapping(ep *Endpoint, cinfo *container.Info) error {
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			logrus.Errorf("port mapping format error, %v", pm)
			continue
		}
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		// err := cmd.Run()
		output, err := cmd.Output()
		if err != nil {
			logrus.Errorf("iptables Output, %v", output)
			continue
		}
	}
	return nil
}

func Connect(networkName string, cinfo *container.Info) error {
	// 从 networks 字典中取到容器连接的网络的信息， networks 字典中保存了当前己经创建的网络
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no Such Network: %s", networkName)
	}

	// 通过调用 IPAM 从网络的网段中获取可用的 IP 作为容器 IP 地址
	ip, err := ipAllocator.Allocate(network.IpRange)
	if err != nil {
		return err
	}

	// 创建网络端点
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: cinfo.PortMapping,
	}
	// 调用网络驱动挂载和配置网络端点
	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}
	// 到容器的 namespace 配置容器网络设备 IP 地址
	if err = configEndpointIpAddressAndRoute(ep, cinfo); err != nil {
		return err
	}
	// 配置容器到宿主机的端口映射
	return configPortMapping(ep, cinfo)
}

func Disconnect(networkName string, cinfo *container.Info) error {
	return nil
}
