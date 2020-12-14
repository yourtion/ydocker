package subsystems

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

// 通过 /proc/self/mountinfo 找出挂载了某个 subsystem 的 hierarchy cgroup 根节点所在的目录
func FindCgroupMountPoint(subsystem string) string {
	// 找出与当前进程相关的 mount 信息
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		// 30 27 0:24 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime shared:13 - cgroup cgroup rw,memory
		fields := strings.Split(txt, " ")
		// 通过最后的 option 是 rw,memory，可以看出这一条挂载的 subsystem 是 memory
		// 那么在 /sys/fs/cgroup/memory 中创建文件夹对应创建的 cgroup，就可以用来做内存的限制
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return fields[4]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return ""
	}
	return ""
}

// 得到 cgroup 在文件系统中的绝对路径
func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountPoint(subsystem)
	if _, err := os.Stat(path.Join(cgroupRoot, cgroupPath)); err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path.Join(cgroupRoot, cgroupPath), 0755); err != nil {
				return "", fmt.Errorf("error create cgroup %v", err)
			}
		}
		return path.Join(cgroupRoot, cgroupPath), nil
	} else {
		return "", fmt.Errorf("cgroup path error %v", err)
	}
}
