package container

import (
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// 创建容器文件系统
func newWorkSpace(rootURL string, mntURL string, volume string) {
	createReadOnlyLayer(rootURL)
	createWriteLayer(rootURL)
	createMountPoint(rootURL, mntURL)
	// 根据 volume 判断是否执行挂载数据卷操作
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			mountVolume(rootURL, mntURL, volumeURLs)
			log.Infof("mount volume: %q", volumeURLs)
		} else {
			log.Infof("Volume parameter input is not correct.")
		}
	}
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
func DeleteWorkSpace(rootURL string, mntURL string, volume string) {
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			deleteMountPointWithVolume(rootURL, mntURL, volumeURLs)
		} else {
			deleteMountPoint(rootURL, mntURL)
		}
	} else {
		deleteMountPoint(rootURL, mntURL)
	}
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

func mountVolume(_ string, mntURL string, volumeURLs []string) {
	// 创建宿主机文件目录
	parentUrl := volumeURLs[0]
	if err := os.MkdirAll(parentUrl, 0777); err != nil {
		log.Infof("Mkdir parent dir %s error. %v", parentUrl, err)
	}
	// 在容器文件系统里创建挂载点
	containerUrl := volumeURLs[1]
	containerVolumeURL := mntURL + containerUrl
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		log.Infof("Mkdir container dir %s error. %v", containerVolumeURL, err)
	}
	// 把宿主机文件目录挂载到容器挂载点
	dirs := "dirs=" + parentUrl
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Mount volume failed. %v", err)
	}
}

func deleteMountPointWithVolume(_ string, mntURL string, volumeURLs []string) {
	// 卸载容器里 volume 挂载点的文件系统
	containerUrl := mntURL + volumeURLs[1]
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Umount volume failed. %v", err)
	}
	// 卸载整个容器文件系统的挂载点
	cmd = exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Umount mountpoint failed. %v", err)
	}
	// 删除容器文件系统挂载点
	if err := os.RemoveAll(mntURL); err != nil {
		log.Infof("Remove mountpoint dir %s error %v", mntURL, err)
	}
}
