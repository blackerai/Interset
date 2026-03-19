package app

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
