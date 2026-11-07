package app

import (
	"context"
	"time"

	"github.com/ankitaniket/clean-me/internal/cleaner"
	"github.com/ankitaniket/clean-me/internal/history"
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
			screen:   ScreenWelcome,
			scanning: false,
			opts:     opts,
			selected: map[string]bool{},
		},
		sp:     sp,
		stages: scanner.All(),
		ctx:    context.Background(),
	}
}

// Init does nothing until the user chooses to scan from the welcome screen —
// clean-me only looks when asked.
func (m Model) Init() tea.Cmd {
	return nil
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
	// Capture free space now so the results screen can show before→after.
	if u := m.st.diskUsage(); u.Total > 0 {
		m.st.freeBefore = u.Free
		m.st.haveFree = true
	}
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
		if m.st.scanning || m.st.cleaning || m.st.spaceBusy {
			var cmd tea.Cmd
			m.sp, cmd = m.sp.Update(msg)
			return m, cmd
		}
		return m, nil
	case spaceLoadedMsg:
		if m.st.spaceCache == nil {
			m.st.spaceCache = map[string][]SpaceEntry{}
		}
		m.st.spaceCache[msg.dir] = msg.entries
		m.st.spaceErr = msg.err
		m.st.spaceBusy = false
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
		_ = history.Record(homeDir(), time.Now(), m.st.outcomes)
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) View() string {
	s := m.st
	if s.showHelp {
		return tui.RenderHelp(s.width, s.height)
	}
	switch s.screen {
	case ScreenWelcome:
		return tui.RenderWelcome(s.width, s.height)
	case ScreenDashboard:
		return tui.RenderDashboard(s.res, s.diskUsage(), s.menuCursor, s.scanning, m.sp.View(), s.width, s.height)
	case ScreenCategories:
		return tui.RenderCategories(s.res, s.catCursor, s.width, s.height)
	case ScreenExplorer:
		if s.activeCat >= 0 && s.activeCat < len(s.res.Categories) {
			cat := s.res.Categories[s.activeCat]
			view := tui.ExplorerView{
				Cat:       cat,
				Items:     s.sortedActiveItems(),
				Cursor:    s.itemCursor,
				Top:       s.explorerTop,
				Selected:  s.selected,
				SortBy:    s.sort.label(),
				Filter:    s.filter,
				Filtering: s.filtering,
			}
			out, top := tui.RenderExplorer(view, s.width, s.height)
			s.explorerTop = top // keep scroll window in sync
			return out
		}
		return ""
	case ScreenReview:
		return tui.RenderReview(s.reviewItems, s.width, s.height)
	case ScreenProgress:
		rv := tui.ProgressView{
			Outcomes:   s.outcomes,
			Total:      len(s.reviewItems),
			Done:       !s.cleaning,
			Spinner:    m.sp.View(),
			HaveFree:   s.haveFree,
			FreeBefore: s.freeBefore,
			FreeAfter:  m.currentFree(),
		}
		return tui.RenderProgress(rv, s.width, s.height)
	case ScreenSpace:
		entries := s.spaceCache[s.spaceDir]
		te := make([]tui.SpaceEntry, len(entries))
		for i, e := range entries {
			te[i] = tui.SpaceEntry{Name: e.Name, Size: e.Size, IsDir: e.IsDir}
		}
		sv := tui.SpaceView{
			Dir:     s.spaceDir,
			Entries: te,
			Cursor:  s.spaceCur,
			Top:     s.spaceTop,
			AtRoot:  len(s.spaceStack) == 0,
			Busy:    s.spaceBusy,
			Err:     s.spaceErr,
			Spinner: m.sp.View(),
		}
		out, top := tui.RenderSpace(sv, s.width, s.height)
		s.spaceTop = top
		return out
	case ScreenHistory:
		return tui.RenderHistory(s.histRuns, s.histCur, s.histMsg, s.width, s.height)
	}
	return ""
}

// currentFree reports the free bytes now, for the before→after delta. Returns
// freeBefore when sizing is unavailable so the delta reads as zero.
func (m Model) currentFree() int64 {
	if !m.st.haveFree {
		return 0
	}
	if u := m.st.diskUsage(); u.Total > 0 {
		return u.Free
	}
	return m.st.freeBefore
}

// Run launches the interactive dashboard.
func Run(opts models.Options) error {
	p := tea.NewProgram(New(opts), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
