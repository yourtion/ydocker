package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"

	_ "github.com/yourtion/ydocker/nsenter"
)

// 为了控制是否执行 C 代 码里面的 setns
const EnvExecPid = "ydocker_pid"
const EnvExecCmd = "ydocker_cmd"

func execContainer(containerName string, comArray []string) {
	// 根据传递过来的容器名获取宿主机对应的 PID
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		log.Errorf("Exec container getContainerPidByName %s error %v", containerName, err)
		return
	}
	// 把命令以空格为分隔符拼接成一个字符串，便于传递
	cmdStr := strings.Join(comArray, " ")
	log.Infof("container pid %s", pid)
	log.Infof("command %s", cmdStr)

	// 简单地 fork 出来一个进程，不 需要这个进程拥有什么命名空间的隔离，然后把这个进程的标准输入输出都绑定到宿主机上
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := os.Setenv(EnvExecPid, pid); err != nil {
		log.Errorf("Setenv EnvExecPid error %v", err)
	}
	if err := os.Setenv(EnvExecCmd, cmdStr); err != nil {
		log.Errorf("Setenv EnvExecPid error %v", err)
	}

	// 获取对应的 PID 环境变量，其实也就是容器的环境变量
	containerEnvs := getEnvsByPid(pid)
	// 将宿主机的环境变量和容器的环境变量都放置到 exec 进程内
	cmd.Env = append(os.Environ(), containerEnvs...)

	if err := cmd.Run(); err != nil {
		log.Errorf("Exec container %s error %v", containerName, err)
	}
}

// 根据指定的 PID 来获取对应进程的环境变量
func getEnvsByPid(pid string) []string {
	// 迸程环境变量存放的位置是 /proc/PID/environ
	path := fmt.Sprintf("/proc/%s/environ", pid)
	contentBytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Errorf("Read file %s error %v", path, err)
		return nil
	}
	// 多个环境变量中的分隔符是 \u0000
	envs := strings.Split(string(contentBytes), "\u0000")
	return envs
}
