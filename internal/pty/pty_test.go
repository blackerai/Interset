package pty

import (
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestProcessManagerStartEmitsOutputAndExit(t *testing.T) {
	spec := StartSpec{}
	if runtime.GOOS == "windows" {
		spec.Command = []string{"cmd.exe", "/c", "echo hello"}
	} else {
		spec.Command = []string{"sh", "-c", "echo hello"}
	}

	manager := NewManager()
	session, err := manager.Start(spec)
	if err != nil {
		t.Fatalf("expected process to start, got error: %v", err)
	}

	defer func() {
		_ = session.Close()
	}()

	var sawStarted bool
	var sawExit bool
	var output strings.Builder
	timeout := time.After(5 * time.Second)

loop:
	for {
		select {
		case event, ok := <-session.Events():
			if !ok {
				break loop
			}
			switch event.Type {
			case EventStarted:
				sawStarted = true
			case EventOutput:
				output.WriteString(event.Data)
			case EventExit:
				sawExit = true
			case EventError:
				t.Fatalf("unexpected runtime error event: %s", event.Err)
			}
		case <-timeout:
			t.Fatal("timed out waiting for process events")
		}
	}

	if !sawStarted {
		t.Fatal("expected started event")
	}

	if !sawExit {
		t.Fatal("expected exit event")
	}

	if !strings.Contains(strings.ToLower(output.String()), "hello") {
		t.Fatalf("expected runtime output to contain hello, got %q", output.String())
	}
}
