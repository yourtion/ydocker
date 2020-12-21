package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/yourtion/ydocker/container"
)

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

func getContainerInfoByName(containerName string) (*container.Info, error) {
	// 根据文件名生成文件绝对路径
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFileDir = configFileDir + container.ConfigName
	// 读取 config.json 文件内的容器信息
	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		log.Errorf("Read file %s error %v", configFileDir, err)
		return nil, err
	}
	// 将 json 文件信息反序列化成容器信息对象
	var containerInfo container.Info
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		log.Errorf("Json unmarshal error %v", err)
		return nil, err
	}

	return &containerInfo, nil
}

// 根据提供的容器名获取对应容器的 PIO
func getContainerPidByName(containerName string) (string, error) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}

func writeContainerInfoByName(containerName string, containerInfo *container.Info) error {
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Json marshal %s error %v", containerName, err)
		return err
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	// 重新写入新的数据覆盖原来的信息
	if err := ioutil.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		log.Errorf("Write file %s error %v", configFilePath, err)
		return err
	}
	return nil
}

// 删除容器信息
func deleteContainerInfo(containerId string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("Remove dir %s error %v", dirURL, err)
	}
}
