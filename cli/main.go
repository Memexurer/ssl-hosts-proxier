package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	var domain, url string

	if len(os.Args) != 3 {
		model := initialModel()
		p := tea.NewProgram(&model)
		model.program = p

		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}

		domain = model.textInputs[0].Value()
		url = model.textInputs[1].Value()
	} else {
		domain = os.Args[1]
		url = os.Args[2]
	}

	StartMain(
		domain,
		url,
	)
}

type model struct {
	textInputs []textinput.Model
	err        error
	step       int
	program    *tea.Program
}

func initialModel() model {
	m := model{
		textInputs: make([]textinput.Model, 2),
		err:        nil,
		step:       0,
	}

	var t textinput.Model
	for i := range m.textInputs {
		t = textinput.New()
		t.CharLimit = 156
		t.Width = 40

		switch i {
		case 0:
			t.Placeholder = "meowy.local"
			t.Focus()
		case 1:
			t.Placeholder = "http://localhost:3000"
		}

		m.textInputs[i] = t
	}

	return m
}

func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

func TestUrl(str string) error {
	u, err := url.ParseRequestURI(str)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("Invalid URL format. Please enter a valid URL (e.g., http://localhost:3000)")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return fmt.Errorf("Failed to reach the site: %v", err)
	}
	defer resp.Body.Close()
	return nil
}

func TestDomain(domain string) error {
	if len(domain) == 0 {
		return fmt.Errorf("Domain cannot be empty")
	}

	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			text := m.textInputs[m.step].Value()

			if m.step == 0 {
				m.err = TestDomain(text)
			}

			if m.step == 1 {
				m.err = TestUrl(text)
			}

			if m.err != nil {
				return m, nil
			} else {
				m.err = nil
				m.step++

				if m.step > 1 {
					return m, tea.Quit
				}

				m.textInputs[m.step].Focus()
				return m, textinput.Blink
			}
		}
	}

	m.textInputs[m.step], cmd = m.textInputs[m.step].Update(msg)
	return m, cmd
}

func (m *model) View() string {
	var s string

	if m.step < len(m.textInputs) {
		var question string
		switch m.step {
		case 0:
			question = "What domain do you want to use for your website?"
		case 1:
			question = "Enter the URL of your app that you want to reverse proxy:"
		}

		s = fmt.Sprintf("%s\n\n%s\n\n(esc to quit)\n", question, m.textInputs[m.step].View())

		if m.err != nil {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("\nError: %v\n", m.err))
		}
	}

	return s
}
