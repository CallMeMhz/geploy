package geploy

import (
	"fmt"
	"sync"
)

const (
	SequenceMode = 0
	ParallelMode = 1
	LooseMode    = 1 << 1
)

type Group struct {
	servers []*Server
	flags   int
	stdouts []string
	stderrs []string
	errs    []error
}

func GroupServers(servers ...*Server) *Group {
	return &Group{
		servers: servers,
		flags:   0,
		stdouts: make([]string, len(servers)),
		stderrs: make([]string, len(servers)),
		errs:    make([]error, len(servers)),
	}
}

func (g *Group) Sequence() *Group {
	g.flags &^= 1
	return g
}

func (g *Group) Parallel() *Group {
	g.flags |= ParallelMode
	return g
}

func (g *Group) Loose(on bool) *Group {
	if on {
		g.flags |= LooseMode
	} else {
		g.flags &^= LooseMode
	}
	return g
}

func (g *Group) Ignore() *Group {
	for i := range g.errs {
		g.errs[i] = nil
	}
	return g
}

func (g *Group) checkError() bool {
	for i := range g.errs {
		if g.errs[i] != nil {
			if g.flags&LooseMode == 0 {
				return true
			} else {
				g.errs[i] = nil
			}
		}
	}
	return false
}

func (g *Group) Run(cmds ...string) *Group {
	if g.checkError() {
		return g
	}

	wg := new(sync.WaitGroup)
	mode := g.flags & (SequenceMode | ParallelMode)
	for i, s := range g.servers {
		switch mode {
		case SequenceMode:
			g.stdouts[i], g.stderrs[i], g.errs[i] = s.Run(cmds...)
		case ParallelMode:
			wg.Add(1)
			go func(i int, s *Server) {
				defer wg.Done()
				g.stdouts[i], g.stderrs[i], g.errs[i] = s.Run(cmds...)
			}(i, s)
		}
	}
	wg.Wait()

	return g
}

func (g *Group) Printf(format string, args ...any) *Group {
	if !g.checkError() {
		fmt.Printf(format, args...)
	}
	return g
}

func (g *Group) Println(args ...any) *Group {
	if !g.checkError() {
		fmt.Println(args...)
	}
	return g
}

func (g *Group) Errors() []error   { return g.errs }
func (g *Group) Error(i int) error { return g.errs[i] }
