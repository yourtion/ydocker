package commands

import (
	"fmt"

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
	Name:   "init",
	Usage:  `初始化容器进程并执行用户进程（禁止外部调用）`,
	Action: initAction,
}

/*
init 实际操作
	1. 获取传递过来的 commands 参数
	2. 执行容器初始化操作
*/
func initAction(ctx *cli.Context) error {
	log.Infof("init come on")
	cmd := ctx.Args().Get(0)
	log.Infof("commands %s", cmd)
	err := container.RunContainerInitProcess()
	return err
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
