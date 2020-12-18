package nsenter

/*
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>

// 这里的 attribute((constructor)) 指的是， 一旦这个包被引用，那么这个函数就会被自动执行
// 类似于构造函数，会在程序一启动的时候运行
__attribute__((constructor)) void enter_namespace(void) {
	char *ydocker_pid;
	// 从环统变盘中获取馆要进入的 PIO
	ydocker_pid = getenv("ydocker_pid");
	if (ydocker_pid) {
		// fprintf(stdout, "got mydocker_pid=%s\n", mydocker_pid);
	} else {
		// fprintf(stdout, "missing ydocker_pid env skip nsenter");
		// 如果没有指定 PID，就不需要向下执行，直接退出
		return;
	}
	char *ydocker_cmd;
	ydocker_cmd = getenv("ydocker_cmd");
	if (ydocker_cmd) {
		// fprintf(stdout, "got ydocker_cmd=%s\n", mydocker_cmd);
	} else {
		// fprintf(stdout, "missing ydocker_cmd env skip nsenter");
		return;
	}
	int i;
	char nspath[1024];
    // 要进入的五种 Namespace
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };

	for (i=0; i<5; i++) {
		// 拼接对应的路径 /proc/pid/ns/ipc 类似这样
		sprintf(nspath, "/proc/%s/ns/%s", ydocker_pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);
		// 这里才真正调用 setns 系统调用进入对应的 Namespace
		if (setns(fd, 0) == -1) {
			// fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
		} else {
			// fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[i]);
		}
		close(fd);
	}
	// 在进入的 Namespace 中执行指定的命令
	int res = system(ydocker_cmd);
	// 退出
	exit(0);
	return;
}
*/
import "C"
