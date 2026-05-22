package tui

import tea "github.com/charmbracelet/bubbletea"

func Run(p Player) error {
	prog := tea.NewProgram(NewModel(p))
	_, err := prog.Run()
	return err
}
