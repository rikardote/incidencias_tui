package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
const (
	Primary      = "#00BFFF" // Deep Sky Blue
	Secondary    = "#FF6B6B" // Soft Red
	Success      = "#00FF7F" // Spring Green
	Warning      = "#FFD700" // Gold
	Info         = "#87CEEB" // Light Sky Blue
	DarkBg       = "#1A1A2E" // Dark navy
	Surface      = "#16213E" // Slightly lighter navy
	TextPrimary  = "#FFFFFF"
	TextSecondary = "#A0A0B0"
	TextMuted    = "#6B7280"
	BorderColor  = "#2D3748"
	ErrorColor   = "#FF4444"
)

// Common styles
var (
	AppStyle = lipgloss.NewStyle().
		Padding(1, 2).
		Background(lipgloss.Color(DarkBg))

	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(Primary)).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color(BorderColor)).
		PaddingBottom(1).
		MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Info)).
		Italic(true)

	MenuItemStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1)

	MenuItemSelectedStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1).
		Background(lipgloss.Color(Primary)).
		Foreground(lipgloss.Color(DarkBg)).
		Bold(true)

	InputStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(BorderColor)).
		Padding(0, 1).
		Width(50)

	InputFocusedStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(Primary)).
		Padding(0, 1).
		Width(50)

	LabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextSecondary)).
		MarginRight(1).
		Width(20).
		Align(lipgloss.Right)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ErrorColor)).
		Bold(true).
		MarginTop(1)

	SuccessStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Success)).
		Bold(true).
		MarginTop(1)

	InfoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Info)).
		MarginTop(1)

	HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextMuted)).
		MarginTop(1)

	TableHeaderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextSecondary)).
		Background(lipgloss.Color("#1E293B")).
		Bold(true).
		Padding(0, 2)

	TableRowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextPrimary)).
		Padding(0, 2)

	TableRowAltStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextPrimary)).
		Background(lipgloss.Color("#1E293B")).
		Padding(0, 2)

	SpinnerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Primary))

	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(Primary)).
		MarginBottom(1)

	FooterStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextMuted)).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(lipgloss.Color(BorderColor)).
		PaddingTop(1).
		MarginTop(1)

	BoxStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(BorderColor)).
		Padding(1, 2).
		Margin(0, 1)

	DocStyle = lipgloss.NewStyle().Margin(0, 2)
)
