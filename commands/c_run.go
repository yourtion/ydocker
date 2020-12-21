package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/yourtion/ydocker/cgroups"
	"github.com/yourtion/ydocker/cgroups/subsystems"
	"github.com/yourtion/ydocker/container"
)

/*
这里的 Start 方法是真正开始前面创建好的 commands 的调用，它首先会 clone 出来一个 namespace 隔离的进程，
然后在子进程中，调用 /proc/self/exe，也就是调用自己，发送 init 参数，调用我们写的 init 方法，去初始化容器的一些资源。
*/
func run(tty bool, comArray []string, res *subsystems.ResourceConfig, containerName, volume, imageName string) {
	containerId := randStringBytes(10)
	if containerName == "" {
		containerName = containerId
	}

	parent, writePipe := container.NewParentProcess(tty, containerName, volume, imageName)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	// 记录容器信息
	if err := recordContainerInfo(parent.Process.Pid, comArray, containerName, containerId, volume); err != nil {
		log.Errorf("Record container info error %v", err)
		return
	}

	// 创建 cgroup manager，并通过调用 set 和 apply 设置资源限制并使限制在容器上生效
	cgroupManager := cgroups.NewCgroupManager("ydocker-cgroup")
	defer func() { _ = cgroupManager.Destroy() }()
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
	if tty {
		if err := parent.Wait(); err != nil {
			log.Error(err)
		}
		deleteContainerInfo(containerName)
		// TODO: Detach DeleteWorkSpace
		container.DeleteWorkSpace(volume, containerName)
		os.Exit(0)
	}
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof(`commands all is "%s"`, command)
	if _, err := writePipe.WriteString(command); err != nil {
		log.Error(err)
	}
	if err := writePipe.Close(); err != nil {
		log.Error(err)
	}
}

// 记录容器信息
func recordContainerInfo(containerPID int, commandArray []string, containerName, id, volume string) error {
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	// 生成容器信息的结构体实例
	containerInfo := &container.Info{
		Id:          id,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		Name:        containerName,
		Volume:      volume,
	}
	// 拼凑存储容器信息的路径
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	// 如果该路径不存在，就级联地全部创建
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		log.Errorf("Mkdir error %s error %v", dirUrl, err)
		return err
	}
	if err := writeContainerInfoByName(containerName, containerInfo); err != nil {
		return err
	}
	return nil
}
