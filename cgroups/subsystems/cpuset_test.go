package subsystems

import (
	"os"
	"path"
	"testing"
)

// 如果报错：cgroup Apply set cgroup proc fail write /sys/fs/cgroup/cpuset/test_cpuset_limit/tasks: no space left on device
// 	- echo 0 > /sys/fs/cgroup/cpuset/test_cpuset_limit/cpuset.mems
// 	- echo 0 > /sys/fs/cgroup/cpuset/test_cpuset_limit/cpuset.cpus

func TestCpusetCgroup(t *testing.T) {
	cpusetSubSys := CpusetSubSystem{}
	resConfig := ResourceConfig{
		CpuSet: "1",
	}
	testCgroup := "test_cpuset_limit"

	if err := cpusetSubSys.Set(testCgroup, &resConfig); err != nil {
		t.Fatalf("cgroup fail %v\n", err)
	}
	stat, _ := os.Stat(path.Join(FindCgroupMountPoint("cpuset"), testCgroup))
	t.Logf("cgroup stats: %+v\n", stat)
	if stat.Name() != testCgroup {
		t.Fatalf("cgroup name fail %s\n", stat.Name())
	}

	if err := cpusetSubSys.Apply(testCgroup, os.Getpid()); err != nil {
		t.Fatalf("cgroup Apply %v\n", err)
	}
	// 将进程移回到根 Cgroup 节点
	if err := cpusetSubSys.Apply("", os.Getpid()); err != nil {
		t.Fatalf("cgroup Apply %v\n", err)
	}

	if err := cpusetSubSys.Remove(testCgroup); err != nil {
		t.Fatalf("cgroup remove %v\n", err)
	}
}
