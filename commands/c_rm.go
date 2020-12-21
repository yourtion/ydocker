package commands

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/yourtion/ydocker/container"
)

func removeContainer(containerName string) {
	// 根据容器名获取容器对应的信息
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("Get container %s info error %v", containerName, err)
		return
	}
	// 只删除处于停止状态的容器
	if containerInfo.Status != container.STOP {
		log.Errorf("Couldn't remove running container")
		return
	}
	// 找到对应存储容器信息的文件路径
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	// 将所有信息包括子目录都移除
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("Remove file %s error %v", dirURL, err)
	}
	container.DeleteWorkSpace(containerInfo.Volume, containerName)
}
