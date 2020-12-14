package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpuSubSystem struct {
}

func (s *CpuSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true)
	if err != nil || res.CpuShare == "" {
		return err
	}
	if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpu.shares"), []byte(res.CpuShare), 0644); err != nil {
		return fmt.Errorf("set cgroup cpu share fail %v", err)
	}
	return nil
}

func (s *CpuSubSystem) Remove(cgroupPath string) error {
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(subsystemCgroupPath)
}

func (s *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
	if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("set cgroup proc fail %v", err)
	}
	return nil
}

func (s *CpuSubSystem) Name() string {
	return "cpu"
}
