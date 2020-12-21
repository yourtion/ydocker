package commands

import (
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"

	"github.com/yourtion/ydocker/container"
)

// 用子目录集合制作 ${imageName}.tar 的镜像
func commitContainer(containerName, imageName string) {
	mntURL := fmt.Sprintf(container.MntUrl, containerName)
	mntURL += "/"
	imageTar := container.RootUrl + "/" + imageName + ".tar"
	fmt.Printf("save to image: %s\n", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", mntURL, err)
	}
}
