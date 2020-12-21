package commands

import (
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/yourtion/ydocker/cgroups"
	"github.com/yourtion/ydocker/container"
)

func stopContainer(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("Get contaienr info by name %s error %v", containerName, err)
		return
	}
	if containerInfo.Status != container.RUNNING || containerInfo.Pid == " " {
		log.Errorf("Contaienr status '%s' is not RUNNING pid: '%s'", containerInfo.Status, containerInfo.Pid)
		return
	}
	// 将 string 类型的 PID 转换为 int 类型
	pidInt, err := strconv.Atoi(containerInfo.Pid)
	if err != nil {
		log.Errorf("Conver pid from string to int error %v", err)
		return
	}
	// 系统调用 kill 可以发送信号给迸程 ，通过传递 syscall.SIGTERM 信号，去杀掉容搭主进程
	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		log.Errorf("Stop container %s error %v", containerName, err)
		return
	}
	// 容器进程已经被 kill， 所以下面需要修改容器状态，PID 可以直为空
	containerInfo.Status = container.STOP
	containerInfo.Pid = ""
	// 将修改后的信息序列化成 json 的字符串
	_ = writeContainerInfoByName(containerName, containerInfo)
	_ = cgroups.NewCgroupManager(container.CGroupName).Destroy()
}
