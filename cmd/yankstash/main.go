package main

import (
	"log"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/sys/windows"

	"github.com/FarazAG/yankstash/internal/tui"
)

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := windows.CoInitializeEx(0, windows.COINIT_APARTMENTTHREADED); err != nil {
		log.Fatal(err)
	}
	defer windows.CoUninitialize()

	p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
