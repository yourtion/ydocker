package subsystems

// ResourceConfig 传递资源限制配置
type ResourceConfig struct {
	// 内存限制
	MemoryLimit string
	// CPU 时间片权重
	CpuShare string
	// CPU 核心数
	CpuSet string
}

// Subsystem 接口，每个 Subsystem 可以实现下面的 4 个接口
// 将 cgroup 抽象成了 path， 原因是 cgroup 在 hierarchy 的路径，便是虚拟文件系统中的虚拟路径
type Subsystem interface {
	// 返回 Subsystem 的名字
	Name() string
	// 设置某个 cgroup 在这个 Subsystem 中的资源限制
	Set(path string, res *ResourceConfig) error
	// 将迸程添加到某个 cgroup 中
	Apply(path string, pid int) error
	// 移除某个 cgroup
	Remove(path string) error
}

// 通过不同的 Subsystem 初始化实例创建资源限制处理链数组
var SubsystemsIns = []Subsystem{
	&CpusetSubSystem{},
	&MemorySubSystem{},
	&CpuSubSystem{},
}
