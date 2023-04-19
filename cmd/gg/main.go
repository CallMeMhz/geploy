package main

import (
	"fmt"
	"os"

	"github.com/callmemhz/geploy"
)

func main() {

	if len(os.Args) == 1 {
		fmt.Println("geploy setup / deploy / remove")
		return
	}

	cfg, err := geploy.LoadDeployConfig()
	if err != nil {
		fmt.Printf("parse deploy.yml failed: %v\n", err)
		return
	}

	usr := cfg.SshConfig.Username
	servers := make([]*geploy.Server, len(cfg.Servers))
	for i, host := range cfg.Servers {
		server, err := geploy.NewServer(host, usr)
		if err != nil {
			fmt.Printf("failed to ssh connect to %s@%s, err: %v", usr, host, err)
			return
		}
		servers[i] = server
	}
	group := geploy.GroupServers(servers...)

	switch method := os.Args[1]; method {
	case "setup":
		geploy.Setup(group, cfg)
	case "deploy", "redeploy":
		geploy.Deploy(group, cfg, method == "redeploy")
	case "remove":

	default:
		fmt.Println("unknwon operation")
	}
}
