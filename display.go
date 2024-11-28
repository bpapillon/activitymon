package main

import (
	"fmt"
	"strings"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/rivo/tview"
)

func PrintHeader() {
	ascii := figure.NewFigure("ActivityMon", "doom", true)
	color.Set(color.FgHiCyan)
	ascii.Print()
	color.Unset()

	subtitle := figure.NewFigure("Track Your Time", "mini", true)
	color.Set(color.FgHiMagenta)
	subtitle.Print()
	color.Unset()

	fmt.Println(strings.Repeat("=", 80))
}

type Monitor struct {
	app       *tview.Application
	logView   *tview.TextView
	statsView *tview.TextView
}

func NewMonitor() *Monitor {
	app := tview.NewApplication()
	flex := tview.NewFlex()

	logView := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			app.Draw()
		}).
		SetScrollable(true).
		SetWordWrap(true)
	logView.SetBorder(true).SetTitle(" Activity Log ")

	statsView := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			app.Draw()
		}).
		SetWordWrap(true)
	statsView.SetBorder(true).SetTitle(" Live Statistics ")

	flex.AddItem(logView, 0, 1, true).
		AddItem(statsView, 0, 1, false)

	app.SetRoot(flex, true)

	return &Monitor{
		app:       app,
		logView:   logView,
		statsView: statsView,
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
