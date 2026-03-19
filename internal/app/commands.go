package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"interset/internal/pty"
	"interset/internal/registry"
	"interset/internal/store"
)

func waitForAsyncMsg(events <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-events
		if !ok {
			return nil
		}
		return msg
	}
}

func detectProvidersCmd(providers []registry.Provider) tea.Cmd {
	return func() tea.Msg {
		return ProviderDetectionFinishedMsg{
			Providers: registry.DetectProviders(providers),
		}
	}
}

func restoreSnapshotCmd(enabled bool) tea.Cmd {
	return func() tea.Msg {
		if !enabled {
			return RestoreCompletedMsg{}
		}
		snapshot, err := store.Load()
		return RestoreCompletedMsg{
			Snapshot: snapshot,
			Err:      err,
		}
	}
}

func saveSnapshotCmd(snapshot store.Snapshot) tea.Cmd {
	return func() tea.Msg {
		return SnapshotSavedMsg{Err: store.Save(snapshot)}
	}
}

func startRuntimeCmd(sessionID string, runtimeVersion int, runtimeMgr pty.Manager, spec pty.StartSpec, events chan<- tea.Msg) tea.Cmd {
	return func() tea.Msg {
		runtime, err := runtimeMgr.Start(spec)
		if err != nil {
			return SessionFailedMsg{
				SessionID:      sessionID,
				RuntimeVersion: runtimeVersion,
				Err:            err.Error(),
			}
		}

		go func() {
			for event := range runtime.Events() {
				switch event.Type {
				case pty.EventStarted:
					events <- SessionStartedMsg{
						SessionID:      sessionID,
						RuntimeVersion: runtimeVersion,
						Runtime:        runtime,
					}
				case pty.EventOutput:
					events <- SessionOutputReceivedMsg{
						SessionID:      sessionID,
						RuntimeVersion: runtimeVersion,
						Data:           event.Data,
					}
				case pty.EventExit:
					events <- SessionExitedMsg{
						SessionID:      sessionID,
						RuntimeVersion: runtimeVersion,
						ExitCode:       event.ExitCode,
					}
				case pty.EventError:
					events <- SessionFailedMsg{
						SessionID:      sessionID,
						RuntimeVersion: runtimeVersion,
						Err:            event.Err,
					}
				}
			}
		}()

		return nil
	}
}

func resizeRuntimeCmd(runtime pty.Session, width, height int) tea.Cmd {
	return func() tea.Msg {
		if runtime == nil {
			return nil
		}
		_ = runtime.Resize(width, height)
		return nil
	}
}

func closeRuntimeCmd(runtime pty.Session) tea.Cmd {
	return func() tea.Msg {
		if runtime == nil {
			return nil
		}
		_ = runtime.Close()
		return nil
	}
}
