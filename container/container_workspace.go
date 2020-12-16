package container

import (
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// 创建容器文件系统
func newWorkSpace(rootURL string, mntURL string) {
	createReadOnlyLayer(rootURL)
	createWriteLayer(rootURL)
	createMountPoint(rootURL, mntURL)
}

// 将 busybox.tar 解压到 busybox 目录下，作为容器的只读层
func createReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	exist, err := pathExists(busyboxURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", busyboxURL, err)
	}
	if exist == false {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("Untar dir %s error %v", busyboxURL, err)
		}
	}
}

// 创建了一个名为 writeLayer 的文件夹作为容器唯一的可写层
func createWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", writeURL, err)
	}
}

func createMountPoint(rootURL string, mntURL string) {
	// 创建 mnt 文件夹作为挂载点
	if err := os.MkdirAll(mntURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", mntURL, err)
	}
	// 把 writeLayer 目录和 busybox 目录 mount 到 mnt 目录下
	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

// Delete the AUFS filesystem while container exit
func DeleteWorkSpace(rootURL string, mntURL string) {
	deleteMountPoint(rootURL, mntURL)
	deleteWriteLayer(rootURL)
}

// 删除 MountPoint
func deleteMountPoint(_ string, mntURL string) {
	// 在 DeleteMountPoint 函数中 umount mnt 目录
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
	// 删除 mnt 目录
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove dir %s error %v", mntURL, err)
	}
}

// 删除 writeLayer 文件夹
func deleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("Remove dir %s error %v", writeURL, err)
	}
}
