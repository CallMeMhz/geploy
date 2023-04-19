package geploy

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var (
	HostKeyCallback ssh.HostKeyCallback
	Signer          ssh.Signer
)

func init() {
	var err error
	HostKeyCallback, err = knownhosts.New(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		panic(err)
	}

	key, err := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"))
	if err != nil {
		panic(err)
	}
	Signer, err = ssh.ParsePrivateKey(key)
	if err != nil {
		panic(err)
	}
}

type Server struct {
	host     string
	port     string
	usr      string
	hostname string

	cli *ssh.Client
}

func NewServer(host, usr string) (*Server, error) {
	cfg := &ssh.ClientConfig{
		User: usr,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(Signer),
		},
		HostKeyCallback: HostKeyCallback,
	}
	port := "22"
	addr := net.JoinHostPort(host, port)
	cli, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, err
	}

	session, err := cli.NewSession()
	if err != nil {
		cli.Close()
		return nil, err
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout
	if err := session.Run("hostname"); err != nil {
		return nil, err
	}
	hostname := strings.TrimRight(stdout.String(), "\n")

	s := &Server{
		host:     host,
		port:     port,
		usr:      usr,
		hostname: hostname,
		cli:      cli,
	}
	return s, nil
}

func (s *Server) Close() error {
	return s.cli.Close()
}

func (s *Server) Session() (*ssh.Session, error) {
	return s.cli.NewSession()
}

func (s *Server) Run(cmds ...string) (string, string, error) {
	if len(cmds) == 0 {
		return "", "", nil
	}

	var stdout, stderr bytes.Buffer
	for _, cmd := range cmds {
		session, err := s.Session()
		if err != nil {
			return "", "", err
		}
		session.Stdout, session.Stderr = &stdout, &stderr
		defer session.Close()

		id := nextCommandId()
		start := time.Now()
		fmt.Printf("%3d [%s] Running %s on %s\n", id, color.HiBlueString(s.hostname), color.HiYellowString(cmd), color.HiBlueString(s.host))
		if err := session.Run(cmd); err != nil {
			fmt.Printf(color.HiRedString(stderr.String()))
			return stdout.String(), stderr.String(), err
		}
		color.HiBlack("%3d [%s] Finished in %s\n", id, s.hostname, time.Since(start).Truncate(time.Millisecond).String())
	}
	return stdout.String(), stderr.String(), nil
}
