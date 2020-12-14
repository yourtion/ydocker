package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpusetSubSystem struct {
}

func (s *CpusetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true)
	if err != nil || res.CpuSet == "" {
		return err
	}
	if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpuset.cpus"), []byte(res.CpuSet), 0644); err != nil {
		return fmt.Errorf("set cgroup cpuset fail %v", err)
	}
	return nil
}

func (s *CpusetSubSystem) Remove(cgroupPath string) error {
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(subsystemCgroupPath)
}

func (s *CpusetSubSystem) Apply(cgroupPath string, pid int) error {
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
	if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("set cgroup proc fail %v", err)
	}
	return nil
}

func (s *CpusetSubSystem) Name() string {
	return "cpuset"
}
