package geploy

import (
	"github.com/fatih/color"
)

const (
	EnsureDocker  = `which docker || (apt-get update -y && apt-get install docker.io -y)`
	EnsureCurl    = `which curl || (apt-get update -y && apt-get install curl)`
	EnsureTraefik = `docker start traefik || docker run --name traefik --detach --restart unless-stopped --log-opt max-size=10m --publish 80:80 --publish 8080:8080 --volume /var/run/docker.sock:/var/run/docker.sock traefik --api.insecure=true --providers.docker --log.level=DEBUG` // fixme 为了方便调试，这里加了 insecure
)

func Setup(g *Group, cfg *DeployConfig) {
	g.
		Println(color.HiMagentaString("Ensure docker is installed ...")).
		Run(EnsureDocker, EnsureCurl).
		Println(color.HiMagentaString("Ensure traefik is running ...")).
		Run(EnsureTraefik).
		Println(color.HiMagentaString("Ensure registry is logged in ...")).
		Run(DockerLoginRegistry(lookup(cfg.Registry.Username), lookup(cfg.Registry.Password)))
}
