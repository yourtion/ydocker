package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

//  memory subsystem 的实现
type MemorySubSystem struct {
}

// 设置 cgroupPath 对应的 cgroup 的内存资源限制
func (s *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	// GetCgroupPath 的作用是获取当前 subsystem 在虚拟文件系统中的路径
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true)
	if err != nil || res.MemoryLimit == "" {
		return err
	}
	// 设置这个 cgroup 的内存限制，即将限制写入到 cgroup 对应目录的 memory.limit_in_bytes 文件中
	if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "memory.limit_in_bytes"), []byte(res.MemoryLimit), 0644); err != nil {
		return fmt.Errorf("set cgroup memory fail %v", err)
	}
	return nil
}

// 删除 cgroupPath 对应的 cgroup
func (s *MemorySubSystem) Remove(cgroupPath string) error {
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(subsystemCgroupPath)
}

// 将一个迸程加入到 cgroupPath 对应的 cgroup 中
func (s *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
	if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("set cgroup proc fail %v", err)
	}
	return nil
}

// 返回 cgroup 的名字
func (s *MemorySubSystem) Name() string {
	return "memory"
}
