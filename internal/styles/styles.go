package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── Paleta ──────────────────────────────────────────────────
const (
	Primary   = "#7C3AED"
	Secondary = "#06B6D4"
	Success   = "#10B981"
	Warning   = "#F59E0B"
	Info      = "#60A5FA"
	Error     = "#EF4444"

	Bg         = "#0F172A"
	Surface    = "#1E293B"
	SurfaceAlt = "#334155"

	TextPrimary   = "#F1F5F9"
	TextSecondary = "#94A3B8"
	TextMuted     = "#64748B"

	Border    = "#334155"
	BorderHi  = "#475569"
)

// ── Colores como lipgloss.Color ─────────────────────────────
var (
	CPrimary   = lipgloss.Color(Primary)
	CSecondary = lipgloss.Color(Secondary)
	CSuccess   = lipgloss.Color(Success)
	CWarning   = lipgloss.Color(Warning)
	CInfo      = lipgloss.Color(Info)
	CError     = lipgloss.Color(Error)
	CBg        = lipgloss.Color(Bg)
	CSurface   = lipgloss.Color(Surface)
	CSurfAlt   = lipgloss.Color(SurfaceAlt)
	CText      = lipgloss.Color(TextPrimary)
	CTextSec   = lipgloss.Color(TextSecondary)
	CTextMuted = lipgloss.Color(TextMuted)
	CBorder    = lipgloss.Color(Border)
	CBorderHi  = lipgloss.Color(BorderHi)
)

// ── Estilos base ────────────────────────────────────────────
var (
	// Header / Footer globales
	HeaderBar = lipgloss.NewStyle().
			Background(CPrimary).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	FooterBar = lipgloss.NewStyle().
			Background(CSurface).
			Foreground(CTextMuted).
			Padding(0, 1)

	FooterKey = lipgloss.NewStyle().
			Foreground(CPrimary).
			Bold(true)

	// Textos
	Title = lipgloss.NewStyle().
		Foreground(CPrimary).
		Bold(true)

	Subtitle = lipgloss.NewStyle().
			Foreground(CSecondary)

	Label = lipgloss.NewStyle().
		Foreground(CTextSec).
		Width(16).
		Align(lipgloss.Right)

	InfoText = lipgloss.NewStyle().
		Foreground(CInfo)

	Muted = lipgloss.NewStyle().
		Foreground(CTextMuted)

	SuccessTxt = lipgloss.NewStyle().
			Foreground(CSuccess).
			Bold(true)

	ErrorTxt = lipgloss.NewStyle().
			Foreground(CError).
			Bold(true)

	// Tablas
	TblHeader = lipgloss.NewStyle().
			Background(CSurfAlt).
			Foreground(CText).
			Bold(true)

	TblRow = lipgloss.NewStyle().
		Background(CSurface).
		Foreground(CText)

	TblRowAlt = lipgloss.NewStyle().
			Background(CBg).
			Foreground(CText)

	TblSelected = lipgloss.NewStyle().
			Background(lipgloss.Color("#1E1B4B")).
			Foreground(CPrimary).
			Bold(true)

	// Inputs
	InputBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(CBorder).
			Background(CSurface).
			Foreground(CText).
			Padding(0, 1)

	InputFocused = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(CPrimary).
			Background(CSurface).
			Foreground(CText).
			Padding(0, 1)

	// Paneles
	Panel = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(CBorder).
		Background(CSurface).
		Padding(1, 2)

	PanelAccent = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(CPrimary).
			Background(CSurface).
			Padding(1, 2)

	// Cards del menú
	MenuCard = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(CBorder).
			Background(CSurface).
			Padding(1, 2).
			Width(34)

	MenuCardActive = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(CPrimary).
			Background(lipgloss.Color("#1E1B4B")).
			Padding(1, 2).
			Width(34)

	// Stepper
	StepDone  = lipgloss.NewStyle().Foreground(CSuccess).Bold(true)
	StepNow   = lipgloss.NewStyle().Foreground(CPrimary).Bold(true)
	StepTodo  = lipgloss.NewStyle().Foreground(CTextMuted)
	StepLine  = lipgloss.NewStyle().Foreground(CSuccess)
	StepLineT = lipgloss.NewStyle().Foreground(CTextMuted)

	// Breadcrumb
	Crumb     = lipgloss.NewStyle().Foreground(CTextMuted)
	CrumbNow  = lipgloss.NewStyle().Foreground(CPrimary).Bold(true)
	CrumbSep  = lipgloss.NewStyle().Foreground(CBorderHi)

	// Status badge
	Badge = lipgloss.NewStyle().
		Background(CSurfAlt).
		Foreground(CTextSec).
		Padding(0, 1)
)

// ── Helpers ─────────────────────────────────────────────────

func Breadcrumb(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	var out []string
	for i, p := range parts {
		if i == len(parts)-1 {
			out = append(out, CrumbNow.Render(p))
		} else {
			out = append(out, Crumb.Render(p))
		}
		if i < len(parts)-1 {
			out = append(out, CrumbSep.Render(" › "))
		}
	}
	return strings.Join(out, "")
}

func Stepper(steps []string, active int) string {
	var parts []string
	for i, s := range steps {
		var lbl string
		switch {
		case i < active:
			lbl = StepDone.Render("✓ "+s)
		case i == active:
			lbl = StepNow.Render("● "+s)
		default:
			lbl = StepTodo.Render("○ "+s)
		}
		parts = append(parts, lbl)
		if i < len(steps)-1 {
			if i < active {
				parts = append(parts, StepLine.Render("───"))
			} else {
				parts = append(parts, StepLineT.Render("───"))
			}
		}
	}
	return strings.Join(parts, " ")
}

// ScreenTitle renders a screen title with optional subtitle
func ScreenTitle(title, sub string) string {
	s := Title.Render(title)
	if sub != "" {
		s += "  " + Subtitle.Render(sub)
	}
	return s
}
