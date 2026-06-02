package tui

import tea "github.com/charmbracelet/bubbletea"

func Run(p Player, a Audio) error {
	prog := tea.NewProgram(NewModel(p, a), tea.WithAltScreen())
	_, err := prog.Run()
	return err
}
