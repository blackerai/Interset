package app

import (
	"interset/internal/pty"
	"interset/internal/registry"
	"interset/internal/store"
)

type focusArea int

const (
	focusHome focusArea = iota
	focusSidebar
	focusTabs
	focusSession
)

type appMode int

const (
	modeHome appMode = iota
	modeWorkspace
)

type ProviderDetectionFinishedMsg struct {
	Providers []registry.Provider
}

type SessionStartRequestedMsg struct {
	SessionID      string
	RuntimeVersion int
}

type RestoreCompletedMsg struct {
	Snapshot store.Snapshot
	Err      error
}

type SnapshotSavedMsg struct {
	Err error
}

type SessionStartedMsg struct {
	SessionID      string
	RuntimeVersion int
	Runtime        pty.Session
}

type SessionOutputReceivedMsg struct {
	SessionID      string
	RuntimeVersion int
	Data           string
}

type SessionExitedMsg struct {
	SessionID      string
	RuntimeVersion int
	ExitCode       int
}

type SessionFailedMsg struct {
	SessionID      string
	RuntimeVersion int
	Err            string
}

type SessionRestartRequestedMsg struct {
	SessionID      string
	RuntimeVersion int
}
