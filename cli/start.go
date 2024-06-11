package main

import (
	"fmt"
	"log"
	"strings"

	sslhostsproxier "github.com/Memexurer/ssl-hosts-proxier"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func StartMain(wompWomp map[string]string) {
	model := initialModelMain()
	model.app = sslhostsproxier.CreateApp(wompWomp)

	p := tea.NewProgram(&model)
	model.program = p

	model.app.Start(func(msg string) {
		model.messages = append(model.messages, msg)
	})
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type modelMain struct {
	err      error
	spinner  spinner.Model
	program  *tea.Program
	app      sslhostsproxier.App
	messages []string
}

func initialModelMain() modelMain {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	m := modelMain{
		err:     nil,
		spinner: s,
	}

	return m
}

func (m *modelMain) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *modelMain) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.messages = make([]string, 0)

			m.app.Stop(func(msg string) {
				m.messages = append(m.messages, msg)
			})

			return m, tea.Quit
		case tea.KeyEnter:
			m.err = fmt.Errorf("meow")
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
	}

	return m, cmd
}

func (m *modelMain) View() string {
	var s string
	var status string

	if m.app.ShuttingDown {
		status = m.spinner.View() + " Shutting down..."
	} else if m.app.IsReady() {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓ App is running!")
	} else {
		status = m.spinner.View() + " Starting app..."
	}

	s = fmt.Sprintf("%s\n\n%s\n\n(esc to quit)\n", status, strings.Join(m.messages, "\n"))

	return s
}
