package main

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00")).
			MarginBottom(1)

	menuStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	systemMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Italic(true)

	playerColors = []lipgloss.Color{
		lipgloss.Color("#00FF00"), // Green
		lipgloss.Color("#FF00FF"), // Magenta
		lipgloss.Color("#00FFFF"), // Cyan
		lipgloss.Color("#FFFF00"), // Yellow
	}

	foodStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	deadStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))
)
