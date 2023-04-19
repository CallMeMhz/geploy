package geploy

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	DockerLoginRegistry  = func(usr, pwd string) string { return fmt.Sprintf(`docker login -u %s -p %s`, usr, pwd) }
	DockerRemoveImage    = func(image, tag string) string { return "docker image rm --force " + image + ":" + tag }
	DockerPullImage      = func(image, tag string) string { return "docker pull " + image + ":" + tag }
	DockerStartTraefik   = `docker start traefik`
	DockerRunHealthcheck = func(service, port, image, tag string, env map[string]string) string {
		cmd := `docker run --detach`
		cmd += fmt.Sprintf(` --name healthcheck-%s-%s`, service, tag)
		cmd += fmt.Sprintf(` --publish 3999:%s`, port)
		cmd += fmt.Sprintf(` --label service=healthcheck-%s`, service)
		for k, v := range env {
			cmd += " -e " + k + "=" + v
		}
		cmd += " " + image + ":" + tag
		return cmd
	}
	HealthCheck = func(entrypoint string) string {
		return `curl --silent --output /dev/null --write-out '%{http_code}' --max-time 2 ` + entrypoint
	}
	DockerStopContainer = func(name string) string {
		return fmt.Sprintf(`docker container ls --all --filter name=%s --quiet | xargs docker stop`, name)
	}
	DockerRemoveContainer = func(name string) string {
		return fmt.Sprintf(`docker container ls --all --filter name=%s --quiet | xargs docker container rm`, name)
	}
	DockerStartApplicationContainer = func(service, port, image, tag, healthcheck string, env map[string]string) string {
		return strings.Join([]string{`docker run --detach --restart unless-stopped`,
			`--log-opt max-size=10m`,
			`--name ` + service + "-" + tag,
			func() string {
				entries := make([]string, 0, len(env))
				for k, v := range env {
					entries = append(entries, "-e "+k+"="+v)
				}
				return strings.Join(entries, " ")
			}(),
			`--label service=` + service,
			`--label role="web"`,
			fmt.Sprintf(`--label traefik.http.routers.%s.rule="PathPrefix(\"/\")"`, service),
			fmt.Sprintf(`--label traefik.http.services.%s.loadbalancer.healthcheck.path="%s"`, service, healthcheck),
			fmt.Sprintf(`--label traefik.http.services.%s.loadbalancer.healthcheck.interval="1s"`, service),
			fmt.Sprintf(`--label traefik.http.middlewares.%s.retry.attempts="5"`, service),
			fmt.Sprintf(`--label traefik.http.middlewares.%s.retry.initialinterval="500ms"`, service),
			image + ":" + tag,
		}, " ")
	}
	DockerPruneOldContainers = func(service string) string {
		return fmt.Sprintf(`docker container prune --force --filter label=service=%s --filter until=72h`, service) // 3 days
	}
	DockerPruneOldImages = func(service string) string {
		return fmt.Sprintf(`docker image prune --all --force --filter label=service=%s --filter until=168h`, service) // 7 days
	}
)

func Deploy(g *Group, cfg *DeployConfig, reDeploy bool) {

	service := cfg.Service
	port := cfg.Port
	image := cfg.Image
	tag := randomHex(32)
	if reDeploy {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command("docker", "images",
			"--filter", "label=service="+service,
			"--format", "{{.Tag}}",
		)
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(stderr.String())
			return
		}
		tag = strings.Split(stdout.String(), "\n")[0]
	} else {
		fmt.Println(color.HiMagentaString("Build and Push Application Image ..."))

		// todo rmoete build
		cmd := exec.Command("docker", "build",
			"-t", image+":latest",
			"-t", image+":"+tag,
			"--label", "service="+service,
			".",
		)
		if run(cmd) != nil {
			return
		}

		cmd = exec.Command("docker", "push", image+":"+tag)
		if run(cmd) != nil {
			return
		}
	}

	g.Parallel().Println(color.HiMagentaString("Force pull image ...")).Run(
		DockerRemoveImage(image, tag),
	).Ignore().Run(
		DockerPullImage(image, tag),
	).Println(color.HiMagentaString("Ensure application can pass healthcheck ...")).Run(
		DockerStartTraefik,
		DockerRunHealthcheck(service, port, image, tag, nil),
		HealthCheck("localhost:3999"+cfg.HealthCheck),
	)
	healtcheckFailed := false
	for i := range g.servers {
		lines := strings.Split(g.stdouts[i], "\n")
		code := lines[len(lines)-1]
		if code != "200" {
			healtcheckFailed = true
			fmt.Println(color.RedString("Health check failed"), "(", color.HiYellowString(code), ")", "on", color.HiBlueString(g.servers[i].host))
		} else {
			fmt.Println("Health check succedd", "(", color.HiGreenString(code), ")", "on", color.HiBlueString(g.servers[i].host))
		}
	}
	g.Parallel().Run(
		DockerStopContainer(fmt.Sprintf(`healthcheck-%s-%s`, service, tag)),
		DockerRemoveContainer(fmt.Sprintf(`healthcheck-%s-%s`, service, tag)),
	)
	if healtcheckFailed {
		return
	}

	g.Parallel().Println(color.HiMagentaString("Start application container ...")).Run(
		DockerStartApplicationContainer(service, port, image, tag, cfg.HealthCheck, nil),
	).Println("Prune old containers and images ...").Run(
		DockerPruneOldContainers(service),
		DockerPruneOldImages(service),
	)
}

func run(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	fmt.Printf("Running %s on %s\n", color.HiYellowString(cmd.String()), color.HiBlueString("localhost"))
	start := time.Now()
	if err := cmd.Run(); err != nil {
		fmt.Println(color.RedString(stderr.String()))
		return err
	}
	fmt.Printf("Finished in %s\n", color.HiYellowString(time.Since(start).Truncate(time.Millisecond).String()))
	return nil
}
