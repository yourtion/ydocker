package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/yourtion/ydocker/cgroups/subsystems"
	"github.com/yourtion/ydocker/container"
)

// 这里定义了 runCommand 的 Flags，其作用类似于运行命令时使用 一 来指定参数
var runCommand = cli.Command{
	Name: "run",
	Usage: `创建一个包含 namespace 和 cgroups 限制的容器 
			ydocker run -ti [command ]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
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
	},
	Action: runAction,
}

/*
这里是 run 命令执行的真正函数。
	1. 判断参数是否包含 command
	2. 获取用户指定的 command
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
	resConf := &subsystems.ResourceConfig{
		MemoryLimit: ctx.String("m"),
		CpuSet:      ctx.String("cpuset"),
		CpuShare:    ctx.String("cpushare"),
	}
	Run(tty, cmdArray, resConf)
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
	1. 获取传递过来的 command 参数
	2. 执行容器初始化操作
*/
func initAction(ctx *cli.Context) error {
	log.Infof("init come on")
	cmd := ctx.Args().Get(0)
	log.Infof("command %s", cmd)
	err := container.RunContainerInitProcess()
	return err
}
