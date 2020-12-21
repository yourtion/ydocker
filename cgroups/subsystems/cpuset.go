package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type CpusetSubSystem struct {
}

func (s *CpusetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true)
	if err != nil {
		return err
	}
	if res.CpuSet == "" {
		res.CpuSet = "0"
	}
	if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpuset.cpus"), []byte(res.CpuSet), 0644); err != nil {
		return fmt.Errorf("set cgroup cpuset fail %v", err)
	}
	if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpuset.mems"), []byte("0"), 0644); err != nil {
		log.Errorf("set cgroup cpuset.mems fail %v", err)
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
