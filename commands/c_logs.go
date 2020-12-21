package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/yourtion/ydocker/container"
)

func logContainer(containerName string) {
	// 找到对应文件夹的位置
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	logFileLocation := dirURL + container.LogFile
	// 打开日志文件
	file, err := os.Open(logFileLocation)
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()
	if err != nil {
		log.Errorf("Log container open file %s error %v", logFileLocation, err)
		return
	}
	// 将文件内的内容都读取出来
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("Log container read file %s error %v", logFileLocation, err)
		return
	}
	// 使用 fmt.Fprint 函数将读出来的内容输入到标准输出，也就是控制台上
	_, _ = fmt.Fprint(os.Stdout, string(content))
}
