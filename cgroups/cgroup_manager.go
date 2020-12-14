package cgroups

import (
	"github.com/sirupsen/logrus"

	"github.com/yourtion/ydocker/cgroups/subsystems"
)

// 把不同 subsystem 中的 cgroup 管理起来，并与容器建立关系
type CgroupManager struct {
	// cgroup 在 hierarchy 中的路径 相当于创建的 cgroup 目录相对于 root cgroup 目录的路径
	Path string
	// 资源配置
	Resource *subsystems.ResourceConfig
}

// 创建一个 CgroupManager
func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

// 将进程 pid 加入到这个 cgroup 中
func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		if err := subSysIns.Apply(c.Path, pid); err != nil {
			logrus.Errorf("Apply cgroup fail %v", err)
		}
	}
	return nil
}

// 设置 cgroup 资源限制
func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		if err := subSysIns.Set(c.Path, res); err != nil {
			logrus.Errorf("Set cgroup fail %v", err)
		}
	}
	return nil
}

// 释放 cgroup
func (c *CgroupManager) Destroy() error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		if err := subSysIns.Remove(c.Path); err != nil {
			logrus.Warnf("remove cgroup fail %v", err)
		}
	}
	return nil
}
