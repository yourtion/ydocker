package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

// 挂载了 memory subsystem 的 hierarchy 的根目录位置
const cgroupMemoryHierarchyMount = "/sys/fs/cgroup/memory"
const groupName = "test-memory-limit2"

func main() {
	if os.Args[0] == "/proc/self/exe" {
		// 容器进程
		fmt.Printf("current pid %d", syscall.Getpid())
		fmt.Println()
		cmd := exec.Command("sh", "-c", "stress --vm-bytes 200m --vm-keep -m 1")
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
		}
		os.Exit(1)
	}
	cmd := exec.Command("/proc/self/exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		// 得到 fork 出来进程映射在外部命名空间的 pid
		fmt.Printf("--> %v \n", cmd.Process.Pid)
		// 在系统默认创建挂载了 memory subsystem 的 Hierarchy 上创建 cgroup
		_ = os.Mkdir(path.Join(cgroupMemoryHierarchyMount, groupName), 0755)
		// 将容器进程加入到这个 cgroup 中，限制 cgroup 进程使用
		_ = ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, groupName, "tasks"), []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
		_ = ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, groupName, "memory.limit_in_bytes"), []byte("100m"), 0644)
	}
	_, _ = cmd.Process.Wait()
}
