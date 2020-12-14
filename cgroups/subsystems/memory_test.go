package subsystems

import (
	"os"
	"path"
	"testing"
)

func TestMemoryCgroup(t *testing.T) {
	memSubSys := MemorySubSystem{}
	resConfig := ResourceConfig{
		MemoryLimit: "1000m",
	}
	testCgroup := "test_memory_limit"

	if err := memSubSys.Set(testCgroup, &resConfig); err != nil {
		t.Fatalf("cgroup fail %v\n", err)
	}
	stat, _ := os.Stat(path.Join(FindCgroupMountPoint("memory"), testCgroup))
	t.Logf("cgroup stats: %+v\n", stat)
	if stat.Name() != testCgroup {
		t.Fatalf("cgroup name fail %s\n", stat.Name())
	}

	if err := memSubSys.Apply(testCgroup, os.Getpid()); err != nil {
		t.Fatalf("cgroup Apply %v\n", err)
	}
	// 将进程移回到根 Cgroup 节点
	if err := memSubSys.Apply("", os.Getpid()); err != nil {
		t.Fatalf("cgroup Apply %v\n", err)
	}

	if err := memSubSys.Remove(testCgroup); err != nil {
		t.Fatalf("cgroup remove %v\n", err)
	}
}
