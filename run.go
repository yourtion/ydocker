package main

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/yourtion/ydocker/cgroups"
	"github.com/yourtion/ydocker/cgroups/subsystems"
	"github.com/yourtion/ydocker/container"
)

/*
这里的 Start 方法是真正开始前面创建好的 command 的调用，它首先会 clone 出来一个 namespace 隔离的进程，
然后在子进程中，调用 /proc/self/exe，也就是调用自己，发送 init 参数，调用我们写的 init 方法，去初始化容器的一些资源。
*/
func Run(tty bool, comArray []string, res *subsystems.ResourceConfig) {
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	// 创建 cgroup manager，并通过调用 set 和 apply 设置资源限制并使限制在容器上生效
	cgroupManager := cgroups.NewCgroupManager("ydocker-cgroup1")
	defer cgroupManager.Destroy()
	// 设置资源限制
	if err := cgroupManager.Set(res); err != nil {
		log.Error(err)
	}
	// 将容器进程加入到各个 subsystem 挂载对应的 cgroup 中
	if err := cgroupManager.Apply(parent.Process.Pid); err != nil {
		log.Error(err)
	}

	// 发送用户命令
	sendInitCommand(comArray, writePipe)
	if err := parent.Wait(); err != nil {
		log.Error(err)
	}
	os.Exit(-1)
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof(`command all is "%s"`, command)
	if _, err := writePipe.WriteString(command); err != nil {
		log.Error(err)
	}
	if err := writePipe.Close(); err != nil {
		log.Error(err)
	}
}
