package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"ydocker/container"
)

/*
这里的 Start 方法是真正开始前面创建好的 command 的调用，它首先会 clone 出来一个 namespace 隔离的进程，
然后在子进程中，调用 /proc/self/exe，也就是调用自己，发送 init 参数，调用我们写的 init 方法，去初始化容器的一些资源。
*/
func Run(tty bool, command string) {
	parent := container.NewParentProcess(tty, command)
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	if err := parent.Wait(); err != nil {
		log.Error(err)
	}
	os.Exit(-1)
}
