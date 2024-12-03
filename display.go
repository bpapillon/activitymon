package main

import (
	"fmt"

	"github.com/rivo/tview"
)

type Monitor struct {
	app        *tview.Application
	headerView *tview.TextView
	logView    *tview.TextView
	statsView  *tview.TextView
}

func NewMonitor() *Monitor {
	app := tview.NewApplication()
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	headerView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	headerText := `[cyan]
 █████╗  ██████╗████████╗██╗██╗   ██╗██╗████████╗██╗   ██╗███╗   ███╗ ██████╗ ███╗   ██╗
██╔══██╗██╔════╝╚══██╔══╝██║██║   ██║██║╚══██╔══╝╚██╗ ██╔╝████╗ ████║██╔═══██╗████╗  ██║
███████║██║        ██║   ██║██║   ██║██║   ██║    ╚████╔╝ ██╔████╔██║██║   ██║██╔██╗ ██║
██╔══██║██║        ██║   ██║╚██╗ ██╔╝██║   ██║     ╚██╔╝  ██║╚██╔╝██║██║   ██║██║╚██╗██║
██║  ██║╚██████╗   ██║   ██║ ╚████╔╝ ██║   ██║      ██║   ██║ ╚═╝ ██║╚██████╔╝██║ ╚████║
╚═╝  ╚═╝ ╚═════╝   ╚═╝   ╚═╝  ╚═══╝  ╚═╝   ╚═╝      ╚═╝   ╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝[white]`

	headerView.SetText(headerText)

	splitFlex := tview.NewFlex()
	logView := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() { app.Draw() }).
		SetScrollable(true).
		SetWordWrap(true)
	logView.SetBorder(true).SetTitle(" Activity Log ")

	statsView := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() { app.Draw() }).
		SetWordWrap(true)
	statsView.SetBorder(true).SetTitle(" Live Statistics ")

	splitFlex.AddItem(logView, 0, 1, true).
		AddItem(statsView, 0, 1, false)

	mainFlex.AddItem(headerView, 8, 0, false).
		AddItem(splitFlex, 0, 1, true)

	app.SetRoot(mainFlex, true)

	return &Monitor{
		app:        app,
		logView:    logView,
		statsView:  statsView,
		headerView: headerView,
	}
}

func (m *Monitor) Start() error {
	go func() {
		if err := m.app.Run(); err != nil {
			panic(err)
		}
	}()
	return nil
}

func (m *Monitor) Stop() {
	m.app.Stop()
}

func (m *Monitor) AddLogEntry(entry string) {
	fmt.Fprintf(m.logView, "%s\n", entry)
	m.logView.ScrollToEnd()
}

func (m *Monitor) UpdateStats(stats string) {
	m.statsView.Clear()
	fmt.Fprintf(m.statsView, "%s", stats)
}
