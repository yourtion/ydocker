package subsystems

import (
	"testing"
)

func runFindCgroupMountPoint(t *testing.T, point string) {
	ret := FindCgroupMountPoint(point)
	t.Logf("%s subsystem mount point: %v\n", point, ret)
	if ret == "" {
		t.Fatalf("%s subsystem mount point failt", point)
	}
}

func TestFindCgroupMountPointCpu(t *testing.T) {
	runFindCgroupMountPoint(t, "cpu")
}

func TestFindCgroupMountPointCpuset(t *testing.T) {
	runFindCgroupMountPoint(t, "cpuset")
}

func TestFindCgroupMountPointMemory(t *testing.T) {
	runFindCgroupMountPoint(t, "memory")
}
