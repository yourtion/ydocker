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

// 将容器的网络端点加入到容器的网络空间中
// 并锁定当前程序所执行的线程，使当前线程进入到容器的网络空间
// 返回值是一个函数指针，执行这个返回函数才会退出容器的网络空间，回归到宿主机的网络空间
func enterContainerNetns(enLink *netlink.Link, cInfo *container.Info) func() {
	// 找到容器的 Net Namespace
	// /proc/[pid]/ns/net 打开这个文件的文件描述符就可以来操作 Net Namespace
	// 而 ContainerInfo 中的 PID，即容器在宿主机上映射的进程 ID
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cInfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		logrus.Errorf("error get container net namespace, %v", err)
	}
	// 取到文件的文件描述符
	nsFD := f.Fd()
	// 锁定当前程序所执行的线程，如果不锁定操作系统线程的话
	// Go 语言的 goroutine 可能会被调度到别的线程上去，就不能保证一直在所需要的网络空间中了
	// 所以调用 runtime.LockOSThread 时要先锁定当前程序执行的线程
	runtime.LockOSThread()

	// 修改网络端点 Veth 的另外一端，将其移动到容器的 Net Namespace 中
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		logrus.Errorf("error set link netns , %v", err)
	}

	// 通过 netns.Get 方法获得当前网络的 Net Namespace
	// 以便后面从容器的 Net Namespace 中退出，回到原本网络的 Net Namespace 中
	origins, err := netns.Get()
	if err != nil {
		logrus.Errorf("error get current netns, %v", err)
	}
	// 设置当前进程到新的网络namespace，并在函数执行完成之后再恢复到之前的namespace
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		logrus.Errorf("error set netns, %v", err)
	}
	// 返回之前 Net Namespace 的函数
	// 在容苦苦的网络空间中，执行完容器配置之后调用 此函数就可以将程序恢复到原生的 Net Namespace
	return func() {
		// 恢复到上面获取到的之前的 Net Namespace
		if err := netns.Set(origins); err != nil {
			logrus.Errorf("netns error: %v", err)
		}
		// 关闭 Net Namespace 文件
		if err := origins.Close(); err != nil {
			logrus.Errorf("clonse Net Namespace error: %v", err)
		}
		// 取消对当附程序的线程锁定
		runtime.UnlockOSThread()
		// 关闭 Namespace 文件
		if err := f.Close(); err != nil {
			logrus.Errorf("clonse Namespace error: %v", err)
		}
	}
}

// 配置容器网络端点的地址和路由
func configEndpointIpAddressAndRoute(ep *Endpoint, cInfo *container.Info) error {
	// 通过网络端点中 Veth 的另一端
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}
	// 将容器的网络端点加入到容器的网络空间中
	// 并使这个函数下面的操作都在这个网络空间中进行
	// 执行完函数后，恢复为默认的网络空间
	defer enterContainerNetns(&peerLink, cInfo)()
	// 获取到容器的 IP 地址及网段，用于配置容器内部接口地址
	// 比如容器 IP 是 192.168.1.2，而网络的网段是 192.168.1.0/24
	// 那么这里产出的 IP 字符串就是 192.168.1.2/24，用于容器内 Veth 端点配置
	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress
	// 调用 setInterfaceIP 函数设置容器内 Veth 端点的 IP
	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("%v,%s", ep.Network, err)
	}
	// 启动容器内的 Veth 端点
	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}
	// Ne Namespace 中默认本地地址 127.0.0.1 的"lo"网卡是关闭状态的
	// 启动它以保证容器访问自己的请求
	if err = setInterfaceUP("lo"); err != nil {
		return err
	}
	// 设置容器内的外部请求都通过容器内的 Veth 端点访问
	// 0.0.0.0/0 的网段，表示所有的 IP 地址段
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	// 构建要添加的路由数据，包括网络设备、网关 IP 及目的网段
	// 相当于 route add -net 0.0.0.0/0 gw {Bridge网桥地址} dev {容器内的Veth端点设备}
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}
	// 调用 netlink 的 RouteAdd 添加路由到容器的网络空间
	// RouteAdd 函数相当于 route add命令
	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}

	return nil
}

// 配置容器到宿主机的端口映射，
func configPortMapping(ep *Endpoint, _ *container.Info) error {
	// 遍历容器端口映射列表
	for _, pm := range ep.PortMapping {
		// 分割成宿主机的端口和容器的端口
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			logrus.Errorf("port mapping format error, %v", pm)
			continue
		}
		// 由于 iptables 没有 Go 语言版本的实现，所以采用 exec.Command 的方式直接调用命令配置
		// 在 iptables 的 PREROUTING 中添加 DNAT 规则，将宿主机的端口请求转发到容器的地址和端口上
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		// 执行 iptables 命令，添加端口映射转发规则
		output, err := cmd.Output()
		if err != nil {
			logrus.Errorf("iptables Output, %v", output)
			continue
		}
	}
	return nil
}

// 连楼容2带到之前创建的网络 ydocker run net testnet -p 8080:80 xxxx
func Connect(networkName string, cInfo *container.Info) error {
	// 从 networks 字典中取到容器连接的网络的信息，如果找不到网络则返回错误
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no Such Network: %s", networkName)
	}

	// 通过调用 IPAM 从网络的网段中获取可用的 IP 作为容器 IP 地址
	ip, err := ipAllocator.Allocate(network.IpRange)
	if err != nil {
		return err
	}

	// 创建网络端点，设置网络端点的 IP、网络和端口映射信息
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cInfo.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: cInfo.PortMapping,
	}
	// 调用网络驱动挂载和配置网络端点
	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}
	// 到容器的 namespace 配置容器网络设备 IP 地址
	if err = configEndpointIpAddressAndRoute(ep, cInfo); err != nil {
		return err
	}
	// 配置容器到宿主机的端口映射
	return configPortMapping(ep, cInfo)
}

func Disconnect(networkName string, cinfo *container.Info) error {
	return nil
}
