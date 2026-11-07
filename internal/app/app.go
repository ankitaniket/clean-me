package app

import (
	"context"

	"github.com/ankitaniket/clean-me/internal/cleaner"
	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/scanner"
	"github.com/ankitaniket/clean-me/internal/tui"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type scanDoneMsg struct{ res models.ScanResult }
type progressMsg struct{ o models.ItemOutcome }
type cleanDoneMsg struct{}

// Model is the Bubble Tea model driving the whole dashboard.
type Model struct {
	st     *State
	sp     spinner.Model
	stages []scanner.Stage
	ctx    context.Context
}

// New builds a fresh model. The TUI treats explicit selection as the gate, so
// it scans/cleans with the max tier and force enabled — nothing is ever acted
// on without the user selecting it and passing review.
func New(opts models.Options) Model {
	opts.MaxTier = models.TierDanger
	opts.Force = true
	if len(opts.Paths) == 0 {
		opts.Paths = []string{"."}
	}
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return Model{
		st: &State{
			screen:   ScreenDashboard,
			scanning: true,
			opts:     opts,
			selected: map[string]bool{},
		},
		sp:     sp,
		stages: scanner.All(),
		ctx:    context.Background(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.sp.Tick, m.scanCmd())
}

func (m Model) scanCmd() tea.Cmd {
	stages, opts, ctx := m.stages, m.st.opts, m.ctx
	return func() tea.Msg {
		return scanDoneMsg{res: scanner.Scan(ctx, stages, opts)}
	}
}

func (m Model) beginClean() (tea.Model, tea.Cmd) {
	ch := make(chan models.ItemOutcome)
	m.st.progressCh = ch
	m.st.outcomes = nil
	m.st.cleaning = true
	m.st.screen = ScreenProgress
	items, opts, ctx := m.st.reviewItems, m.st.opts, m.ctx
	go func() {
		cleaner.Execute(ctx, items, opts, func(o models.ItemOutcome) { ch <- o })
		close(ch)
	}()
	return m, tea.Batch(m.sp.Tick, waitProgress(ch))
}

func waitProgress(ch chan models.ItemOutcome) tea.Cmd {
	return func() tea.Msg {
		o, ok := <-ch
		if !ok {
			return cleanDoneMsg{}
		}
		return progressMsg{o: o}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.st.width, m.st.height = msg.Width, msg.Height
		return m, nil
	case spinner.TickMsg:
		if m.st.scanning || m.st.cleaning {
			var cmd tea.Cmd
			m.sp, cmd = m.sp.Update(msg)
			return m, cmd
		}
		return m, nil
	case scanDoneMsg:
		m.st.res = msg.res
		m.st.scanning = false
		return m, nil
	case progressMsg:
		m.st.outcomes = append(m.st.outcomes, msg.o)
		return m, waitProgress(m.st.progressCh)
	case cleanDoneMsg:
		m.st.cleaning = false
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) View() string {
	s := m.st
	switch s.screen {
	case ScreenDashboard:
		return tui.RenderDashboard(s.res, s.menuCursor, s.scanning, m.sp.View(), s.width, s.height)
	case ScreenCategories:
		return tui.RenderCategories(s.res, s.catCursor, s.width, s.height)
	case ScreenExplorer:
		if s.activeCat >= 0 && s.activeCat < len(s.res.Categories) {
			return tui.RenderExplorer(s.res.Categories[s.activeCat], s.itemCursor, s.selected, s.width, s.height)
		}
		return ""
	case ScreenReview:
		return tui.RenderReview(s.reviewItems, s.width, s.height)
	case ScreenProgress:
		return tui.RenderProgress(s.outcomes, len(s.reviewItems), !s.cleaning, m.sp.View(), s.width, s.height)
	}
	return ""
}

// Run launches the interactive dashboard.
func Run(opts models.Options) error {
	p := tea.NewProgram(New(opts), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
