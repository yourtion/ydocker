package commands

import (
	"encoding/json"
	"fmt"
	"math/rand"
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
func run(tty bool, comArray []string, res *subsystems.ResourceConfig, volume string, name string) {
	parent, writePipe := container.NewParentProcess(tty, volume, name)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	// 记录容器信息
	containerName, err := recordContainerInfo(parent.Process.Pid, comArray, name)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return
	}

	// 创建 cgroup manager，并通过调用 set 和 apply 设置资源限制并使限制在容器上生效
	cgroupManager := cgroups.NewCgroupManager("ydocker-cgroup")
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
	if tty {
		if err := parent.Wait(); err != nil {
			log.Error(err)
		}
		deleteContainerInfo(containerName)
		// TODO: Detach DeleteWorkSpace
		mntURL := "/root/mnt/"
		rootURL := "/root/"
		container.DeleteWorkSpace(rootURL, mntURL, volume)
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

// 生成随机字符串
func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// 记录容器信息
func recordContainerInfo(containerPID int, commandArray []string, containerName string) (string, error) {
	// 首先生成 10 位数字的容器 ID
	id := randStringBytes(10)
	// 以当前时间为容器创建时间
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	// 如果用户不指定容器名，那么就以容器 id 当作容器名
	if containerName == "" {
		containerName = id
	}
	// 生成容器信息的结构体实例
	containerInfo := &container.Info{
		Id:          id,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		Name:        containerName,
	}

	// 将容器信息的对象 json 序列化成字符串
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return "", err
	}
	jsonStr := string(jsonBytes)

	// 拼凑存储容器信息的路径
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	// 如果该路径不存在，就级联地全部创建
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		log.Errorf("Mkdir error %s error %v", dirUrl, err)
		return "", err
	}
	fileName := dirUrl + "/" + container.ConfigName
	// 创建位终的配置文件 config.json 文件
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		log.Errorf("Create file %s error %v", fileName, err)
		return "", err
	}
	// 将 json 化之后的数据写入到文件中
	if _, err := file.WriteString(jsonStr); err != nil {
		log.Errorf("File write string error %v", err)
		return "", err
	}

	return containerName, nil
}

// 删除容器信息
func deleteContainerInfo(containerId string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("Remove dir %s error %v", dirURL, err)
	}
}
