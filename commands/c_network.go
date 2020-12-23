package commands

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/yourtion/ydocker/network"
)

func createNetwork(context *cli.Context) error {
	if len(context.Args()) < 1 {
		return fmt.Errorf("missing network name")
	}
	if err := network.Init(); err != nil {
		return fmt.Errorf("init network error: %+v", err)
	}
	if err := network.CreateNetwork(context.String("driver"), context.String("subnet"), context.Args()[0]); err != nil {
		return fmt.Errorf("create network error: %+v", err)
	}
	return nil
}

func listNetwork(_ *cli.Context) error {
	if err := network.Init(); err != nil {
		return fmt.Errorf("init network error: %+v", err)
	}
	network.ListNetwork()
	return nil
}

func removeNetwork(context *cli.Context) error {
	if len(context.Args()) < 1 {
		return fmt.Errorf("missing network name")
	}
	if err := network.Init(); err != nil {
		return fmt.Errorf("init network error: %+v", err)
	}
	if err := network.DeleteNetwork(context.Args()[0]); err != nil {
		return fmt.Errorf("remove network error: %+v", err)
	}
	return nil
}
