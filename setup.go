package geploy

import (
	"fmt"

	"github.com/fatih/color"
)

const (
	EnsureDocker  = `which docker || (apt-get update -y && apt-get install docker.io -y)`
	EnsureCurl    = `which curl || (apt-get update -y && apt-get install curl)`
	EnsureTraefik = `docker start traefik || docker run --name traefik --detach --restart unless-stopped --log-opt max-size=10m --publish 80:80 --publish 8080:8080 --volume /var/run/docker.sock:/var/run/docker.sock traefik --api.insecure=true --providers.docker --log.level=DEBUG` // fixme 为了方便调试，这里加了 insecure
)

func Setup(g *Group, cfg *DeployConfig) {
	g.Parallel().Println(color.HiMagentaString("Ensrue docker is installed ...")).Run(
		EnsureDocker,
		EnsureCurl,
	).Println(color.HiMagentaString("Ensrue traefik is running ...")).Run(
		EnsureTraefik,
	).Println(color.HiMagentaString("Ensrue registry is logged in ...")).Run(
		DockerLoginRegistry(lookup(cfg.Registry.Username), lookup(cfg.Registry.Password)),
	)

	for i, err := range g.Errors() {
		if err != nil {
			s := g.servers[i]
			fmt.Printf("[%s] %s on %s\n", color.HiBlueString(s.hostname), color.HiRedString(err.Error()), color.HiBlueString(s.host))
		}
	}
}
