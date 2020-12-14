package subsystems

import (
	"os"
	"path"
	"testing"
)

func TestCpuCgroup(t *testing.T) {
	cpuSubSys := CpuSubSystem{}
	resConfig := ResourceConfig{
		CpuShare: "512",
	}
	testCgroup := "test_cpu_limit"

	if err := cpuSubSys.Set(testCgroup, &resConfig); err != nil {
		t.Fatalf("cgroup fail %v\n", err)
	}
	stat, _ := os.Stat(path.Join(FindCgroupMountPoint("cpu"), testCgroup))
	t.Logf("cgroup stats: %+v\n", stat)
	if stat.Name() != testCgroup {
		t.Fatalf("cgroup name fail %s\n", stat.Name())
	}

	if err := cpuSubSys.Apply(testCgroup, os.Getpid()); err != nil {
		t.Fatalf("cgroup Apply %v\n", err)
	}
	// 将进程移回到根 Cgroup 节点
	if err := cpuSubSys.Apply("", os.Getpid()); err != nil {
		t.Fatalf("cgroup Apply %v\n", err)
	}

	if err := cpuSubSys.Remove(testCgroup); err != nil {
		t.Fatalf("cgroup remove %v\n", err)
	}
}
