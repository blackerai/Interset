package pty

import (
	"io"
	"os"
	"os/exec"
	"sync"
)

type EventType string

const (
	EventStarted EventType = "started"
	EventOutput  EventType = "output"
	EventExit    EventType = "exit"
	EventError   EventType = "error"
)

type Event struct {
	Type     EventType
	Data     string
	Err      string
	ExitCode int
}

type StartSpec struct {
	Command []string
	Cwd     string
	Env     map[string]string
}

type Session interface {
	io.Writer
	Resize(width, height int) error
	Close() error
	Events() <-chan Event
}

type Manager interface {
	Start(spec StartSpec) (Session, error)
}

type ProcessManager struct{}

func NewManager() Manager {
	return &ProcessManager{}
}

func (m *ProcessManager) Start(spec StartSpec) (Session, error) {
	if len(spec.Command) == 0 {
		return nil, exec.ErrNotFound
	}

	cmd := exec.Command(spec.Command[0], spec.Command[1:]...)
	if spec.Cwd != "" && spec.Cwd != "~" {
		cmd.Dir = spec.Cwd
	}
	cmd.Env = mergeEnv(spec.Env)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	session := &processSession{
		cmd:    cmd,
		stdin:  stdin,
		events: make(chan Event, 256),
	}

	session.events <- Event{Type: EventStarted}
	go session.readLoop(stdout)
	go session.readLoop(stderr)
	go session.waitLoop()

	return session, nil
}

type processSession struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	events    chan Event
	closeOnce sync.Once
}

func (s *processSession) Write(p []byte) (int, error) {
	return s.stdin.Write(p)
}

func (s *processSession) Resize(width, height int) error {
	return nil
}

func (s *processSession) Close() error {
	var err error
	s.closeOnce.Do(func() {
		if s.cmd.Process != nil {
			err = s.cmd.Process.Kill()
		}
		_ = s.stdin.Close()
	})
	return err
}

func (s *processSession) Events() <-chan Event {
	return s.events
}

func (s *processSession) readLoop(reader io.ReadCloser) {
	defer reader.Close()

	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			s.events <- Event{
				Type: EventOutput,
				Data: string(buf[:n]),
			}
		}
		if err != nil {
			return
		}
	}
}

func (s *processSession) waitLoop() {
	err := s.cmd.Wait()
	exitCode := 0
	if s.cmd.ProcessState != nil {
		exitCode = s.cmd.ProcessState.ExitCode()
	}

	if err != nil && exitCode == 0 {
		s.events <- Event{
			Type: EventError,
			Err:  err.Error(),
		}
	} else {
		s.events <- Event{
			Type:     EventExit,
			ExitCode: exitCode,
		}
	}

	close(s.events)
}

func mergeEnv(overrides map[string]string) []string {
	base := os.Environ()
	if len(overrides) == 0 {
		return base
	}

	index := make(map[string]int, len(base))
	for i, entry := range base {
		for j := 0; j < len(entry); j++ {
			if entry[j] == '=' {
				index[entry[:j]] = i
				break
			}
		}
	}

	for key, value := range overrides {
		item := key + "=" + value
		if idx, ok := index[key]; ok {
			base[idx] = item
			continue
		}
		base = append(base, item)
	}

	return base
}
