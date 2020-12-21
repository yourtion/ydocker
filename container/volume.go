package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// 创建容器文件系统
func newWorkSpace(volume, imageName, containerName string) {
	_ = createReadOnlyLayer(imageName)
	_ = createWriteLayer(containerName)
	_ = createMountPoint(containerName, imageName)
	// 根据 volume 判断是否执行挂载数据卷操作
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			_ = mountVolume(volumeURLs, containerName)
			log.Infof("mount volume: %q", volumeURLs)
		} else {
			log.Infof("Volume parameter input is not correct.")
		}
	}
}

// 解压 tar 格式的锐像文件作为只读层
func createReadOnlyLayer(imageName string) error {
	unTarFolderUrl := RootUrl + "/" + imageName + "/"
	imageUrl := RootUrl + "/" + imageName + ".tar"
	exist, err := pathExists(unTarFolderUrl)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", unTarFolderUrl, err)
		return err
	}
	if exist {
		return nil
	}
	if err := os.MkdirAll(unTarFolderUrl, 0622); err != nil {
		log.Errorf("Mkdir %s error %v", unTarFolderUrl, err)
		return err
	}
	if _, err := exec.Command("tar", "-xvf", imageUrl, "-C", unTarFolderUrl).CombinedOutput(); err != nil {
		log.Errorf("Untar dir %s error %v", unTarFolderUrl, err)
		return err
	}
	return nil
}

// 创建了一个名为 writeLayer 的文件夹作为容器唯一的可写层
func createWriteLayer(containerName string) error {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		log.Infof("Mkdir write layer dir %s error. %v", writeURL, err)
		return err
	}
	return nil
}

// 创建容器的根目录，然后把镜像只读层和容器读写层挂载到容器根目录，成为容器的文件系统
func createMountPoint(containerName, imageName string) error {
	mntUrl := fmt.Sprintf(MntUrl, containerName)
	// 创建 mnt 文件夹作为挂载点
	if err := os.MkdirAll(mntUrl, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", mntUrl, err)
		return err
	}
	// 把 writeLayer 目录和 busybox 目录 mount 到 mnt 目录下
	tmpWriteLayer := fmt.Sprintf(WriteLayerUrl, containerName)
	tmpImageLocation := RootUrl + "/" + imageName
	mntURL := fmt.Sprintf(MntUrl, containerName)
	dirs := "dirs=" + tmpWriteLayer + ":" + tmpImageLocation
	_, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL).CombinedOutput()
	if err != nil {
		log.Errorf("Run command for creating mount point failed %v", err)
		return err
	}
	return nil
}

// 当容器退出时，删除容器的相关文件系统
func DeleteWorkSpace(volume, containerName string) {
	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			_ = deleteMountPointWithVolume(volumeURLs, containerName)
		} else {
			_ = deleteMountPoint(containerName)
		}
	} else {
		_ = deleteMountPoint(containerName)
	}
	_ = deleteWriteLayer(containerName)
}

// 删除未挂载数据卷的容器文件系统
func deleteMountPoint(containerName string) error {
	mntURL := fmt.Sprintf(MntUrl, containerName)
	// 在 DeleteMountPoint 函数中 umount mnt 目录
	if _, err := exec.Command("umount", mntURL).CombinedOutput(); err != nil {
		log.Errorf("%v", err)
		return err
	}
	// 删除 mnt 目录
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove dir %s error %v", mntURL, err)
		return err
	}
	return nil
}

// 删除容器的读写层
func deleteWriteLayer(containerName string) error {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("Remove dir %s error %v", writeURL, err)
		return err
	}
	return nil
}

func mountVolume(volumeURLs []string, containerName string) error {
	// 创建宿主机文件目录
	parentUrl := volumeURLs[0]
	if err := os.MkdirAll(parentUrl, 0777); err != nil {
		log.Infof("Mkdir parent dir %s error. %v", parentUrl, err)
	}
	// 在容器文件系统里创建挂载点
	containerUrl := volumeURLs[1]
	mntURL := fmt.Sprintf(MntUrl, containerName)
	containerVolumeURL := mntURL + "/" + containerUrl
	if err := os.MkdirAll(containerVolumeURL, 0777); err != nil {
		log.Infof("Mkdir container dir %s error. %v", containerVolumeURL, err)
		return err
	}
	// 把宿主机文件目录挂载到容器挂载点
	dirs := "dirs=" + parentUrl
	if _, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL).CombinedOutput(); err != nil {
		log.Errorf("Mount volume failed. %v", err)
		return err
	}
	return nil
}

// 删除挂载数据卷容器的文件系统
func deleteMountPointWithVolume(volumeURLs []string, containerName string) error {
	mntURL := fmt.Sprintf(MntUrl, containerName)
	// 卸载容器里 volume 挂载点的文件系统
	containerUrl := mntURL + "/" + volumeURLs[1]
	if _, err := exec.Command("umount", containerUrl).CombinedOutput(); err != nil {
		log.Errorf("Umount volume %s failed. %v", containerUrl, err)
		return err
	}
	if _, err := exec.Command("umount", mntURL).CombinedOutput(); err != nil {
		log.Errorf("Umount mountpoint %s failed. %v", mntURL, err)
		return err
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove mountpoint dir %s error %v", mntURL, err)
	}
	return nil
}
