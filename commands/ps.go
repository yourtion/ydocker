package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"

	"github.com/yourtion/ydocker/container"
)

func listContainers() {
	// 找到存储容器信息的路径 /var/run/ydocker
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, "")
	dirURL = dirURL[:len(dirURL)-1]
	// 卖取该文件夹下的所有文件
	files, err := ioutil.ReadDir(dirURL)
	if err != nil {
		log.Errorf("Read dir %s error %v", dirURL, err)
		return
	}
	// 遍历该文件夹下的所有文件
	var containers []*container.Info
	for _, file := range files {
		// 获取文件名
		containerName := file.Name()
		// 根据容器配置文件获取对应的信息，然后转换成容器信息的对象
		tmpContainer, err := getContainerInfo(containerName)
		if err != nil {
			log.Errorf("Get container info error %v", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}
	// 使用 tabwriter.NewWriter 在控制台打印出容器信息（用于在控制台打印对齐的表格）
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	// 控制台输出的信息列
	_, _ = fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime)
	}
	// 刷新标准输出流缓存区，将容器列表打印出来
	if err := w.Flush(); err != nil {
		log.Errorf("Flush error %v", err)
		return
	}
}

func getContainerInfo(containerName string) (*container.Info, error) {
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
