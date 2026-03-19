package pty

import "io"

type Session interface {
	io.Reader
	io.Writer
	Resize(width, height int) error
	Close() error
}

type Manager interface {
	Start(command []string, cwd string, env map[string]string) (Session, error)
}
