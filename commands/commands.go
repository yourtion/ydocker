package commands

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/yourtion/ydocker/cgroups/subsystems"
	"github.com/yourtion/ydocker/container"
)

func GetCommandList() []cli.Command {
	return []cli.Command{
		initCommand,
		runCommand,
		commitCommand,
		listCommand,
		logCommand,
		execCommand,
		stopCommand,
	}
}

// 这里定义了 runCommand 的 Flags，其作用类似于运行命令时使用 一 来指定参数
var runCommand = cli.Command{
	Name: "run",
	Usage: `创建一个包含 namespace 和 cgroups 限制的容器 
			ydocker run -ti [commands]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		// 添加 -v 标签
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
		// 提供 run 后面的 -name 指定容器名字参数
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
	},
	Action: runAction,
}

/*
这里是 run 命令执行的真正函数。
	1. 判断参数是否包含 commands
	2. 获取用户指定的 commands
	3. 调用 Run function 去准备启动容器
*/
func runAction(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("缺少 container 参数")
	}
	var cmdArray []string
	for _, arg := range ctx.Args() {
		cmdArray = append(cmdArray, arg)
	}
	tty := ctx.Bool("ti")
	detach := ctx.Bool("d")
	if tty && detach {
		return fmt.Errorf("ti and d paramter can not both provided")
	}
	resConf := &subsystems.ResourceConfig{
		MemoryLimit: ctx.String("m"),
		CpuSet:      ctx.String("cpuset"),
		CpuShare:    ctx.String("cpushare"),
	}
	// 把 volume 参数传给 Run 函数
	volume := ctx.String("v")
	// 将取到的容器名称传递下去，如果没有则取到的值为空
	containerName := ctx.String("name")
	run(tty, cmdArray, resConf, volume, containerName)
	return nil
}

// 这里，定义了 initCommand 的具体操作，此操作为内部方法，禁止外部调用
var initCommand = cli.Command{
	Name:  "init",
	Usage: `初始化容器进程并执行用户进程（禁止外部调用）`,
	Action: func(ctx *cli.Context) error {
		log.Infof("init come on")
		cmd := ctx.Args().Get(0)
		log.Infof("commands %s", cmd)
		err := container.RunContainerInitProcess()
		return err
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}
		imageName := context.Args().Get(0)
		commitContainer(imageName)
		return nil
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the containers",
	Action: func(context *cli.Context) error {
		listContainers()
		return nil
	},
}

var logCommand = cli.Command{
	Name:  "logs",
	Usage: "print logs of a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("please input your container name")
		}
		containerName := context.Args().Get(0)
		logContainer(containerName)
		return nil
	},
}

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(context *cli.Context) error {
		// This is for callback
		if os.Getenv(EnvExecPid) != "" {
			log.Infof("pid callback pid %s", os.Getgid())
			return nil
		}
		// 我们希望命令格式是 ydocker exec 容器名命令
		if len(context.Args()) < 2 {
			return fmt.Errorf("missing container name or command")
		}
		containerName := context.Args().Get(0)
		// 将除了容器名之外的参数当作需要执行的命令处理
		var commandArray []string
		for _, arg := range context.Args().Tail() {
			commandArray = append(commandArray, arg)
		}
		// 执行命令
		execContainer(containerName, commandArray)
		return nil
	},
}

var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}
		containerName := context.Args().Get(0)
		stopContainer(containerName)
		return nil
	},
}
