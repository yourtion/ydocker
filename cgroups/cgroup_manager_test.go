package cgroups

import (
	"os"
	"testing"

	"github.com/yourtion/ydocker/cgroups/subsystems"
)

func TestNewCgroupManager(t *testing.T) {
	testCgroup := "test_limit"

	manager := NewCgroupManager(testCgroup)
	defer func() {
		if err := manager.Destroy(); err != nil {
			t.Fatalf("Destroy fail :%v\n", err)
		}
	}()

	res := &subsystems.ResourceConfig{
		MemoryLimit: "100m",
		CpuSet:      "0",
		CpuShare:    "1",
	}
	if err := manager.Set(res); err != nil {
		t.Fatalf("Set fail :%v\n", err)
	}
	if err := manager.Apply(os.Getpid()); err != nil {
		t.Fatalf("Apply fail :%v\n", err)
	}
}
