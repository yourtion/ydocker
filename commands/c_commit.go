package commands

import (
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// 将容器文件系统打包成 ${imageName}.tar 文件
func commitContainer(imageName string) {
	mntURL := "/root/mnt"
	imageTar := "/root/" + imageName + ".tar"
	fmt.Printf("save to image: %s\n", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", mntURL, err)
	}
}
