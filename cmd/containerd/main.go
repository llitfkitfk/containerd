package main

import (
	gocontext "context"
	"fmt"
	"os"

	"github.com/llitfkitfk/containerd/log"
	"github.com/llitfkitfk/containerd/server"
	"github.com/llitfkitfk/containerd/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const usage = `
                    __        _                     __
  _________  ____  / /_____ _(_)___  ___  _________/ /
 / ___/ __ \/ __ \/ __/ __ ` + "`" + `/ / __ \/ _ \/ ___/ __  /
/ /__/ /_/ / / / / /_/ /_/ / / / / /  __/ /  / /_/ /
\___/\____/_/ /_/\__/\__,_/_/_/ /_/\___/_/   \__,_/

high performance container runtime
`

func init() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(c.App.Name, version.Package, c.App.Version, version.Revision)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "containerd"
	app.Version = version.Version
	app.Usage = usage
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config,c",
			Usage: "path to the configuration file",
			Value: defaultConfigPath,
		},
	}
	app.Commands = []cli.Command{}
	app.Action = func(context *cli.Context) error {

		var (
			ctx    = log.WithModule(gocontext.Background(), "containerd")
			config = defaultConfig()
		)
		if err := server.LoadConfig(context.GlobalString("config"), config); err != nil && !os.IsNotExist(err) {
			return err
		}

		// apply flags to the config
		if err := applyFlags(context, config); err != nil {
			return err
		}

		log.G(ctx).WithFields(logrus.Fields{
			"version":  version.Version,
			"revision": version.Revision,
		}).Info("starting containerd")

		return nil
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "containerd: %s\n", err)
		os.Exit(1)
	}
}

func applyFlags(context *cli.Context, config *server.Config) error {
	// the order for config vs flag values is that flags will always override
	// the config values if they are set
	if err := setLevel(context, config); err != nil {
		return err
	}
	for _, v := range []struct {
		name string
		d    *string
	}{
		{
			name: "root",
			d:    &config.Root,
		},
		{
			name: "state",
			d:    &config.State,
		},
		{
			name: "address",
			d:    &config.GRPC.Address,
		},
	} {
		if s := context.GlobalString(v.name); s != "" {
			*v.d = s
		}
	}
	return nil
}

func setLevel(context *cli.Context, config *server.Config) error {
	l := context.GlobalString("log-level")
	if l == "" {
		l = config.Debug.Level
	}
	if l != "" {
		lvl, err := logrus.ParseLevel(l)
		if err != nil {
			return err
		}
		logrus.SetLevel(lvl)
	}
	return nil
}
