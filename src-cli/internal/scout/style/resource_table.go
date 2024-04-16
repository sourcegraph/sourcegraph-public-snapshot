package style

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var resourceTableStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type resourceTableModel struct {
	table table.Model
}

func (m resourceTableModel) Init() tea.Cmd { return nil }

func (m resourceTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "ctrl+c":
			return m, tea.Quit
		case "c":
			m.copyRowToClipboard(m.table.SelectedRow())
			copiedMessage := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#32CD32")).
				Render(fmt.Sprintf(
					"Copied resource allocations for '%s' to clipboard",
					m.table.SelectedRow()[0],
				))
			return m, tea.Batch(
				tea.Printf(
					copiedMessage,
				),
			)
		case "C":
			tmpDir := os.TempDir()
			filePath := filepath.Join(tmpDir, "resource-dump.txt")
			m.dump(m.table.Rows(), filePath)
			savedMessage := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#32CD32")).
				Render(fmt.Sprintf(
					"saved resource allocations to %s",
					filePath,
				))
			return m, tea.Batch(
				tea.Printf(
					savedMessage,
				),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m resourceTableModel) View() string {
	s := "\n > Press 'j' and 'k' to go up and down\n"
	s += " > Press 'c' to copy highlighted row to clipboard\n"
	s += " > Press 'C' to copy all rows to a file\n"
	s += " > Press 'q' to quit\n\n"
	s += resourceTableStyle.Render(m.table.View()) + "\n"
	return s
}

func (m resourceTableModel) dump(rows []table.Row, filePath string) error {
	dumpFile, err := os.Create(filePath)
	if err != nil {
		return errors.Wrap(err, "error while creating new file")
	}
	defer dumpFile.Close()

	tw := tabwriter.NewWriter(dumpFile, 0, 0, 3, ' ', 0)
	defer tw.Flush()

	// default to docker terms
	headers := []string{
		"NAME",
		"CPU CORES",
		"CPU SHARES",
		"MEM LIMITS",
		"MEM RESERVATIONS",
	}

	// kubernetes rows will always have 6 items
	// change column headers to reflect k8s terms
	if len(rows[0]) == 6 {
		headers = []string{
			"NAME",
			"CPU LIMITS",
			"CPU REQUESTS",
			"MEM LIMITS",
			"MEM REQUESTS",
			"CAPACITY",
		}
	}

	fmt.Fprintf(tw, strings.Join(headers, "\t")+"\n")

	for _, row := range rows {
		values := []string{row[0], row[1], row[2], row[3], row[4]}
		if len(row) == 6 {
			values = append(values, row[5])
		}
		fmt.Fprintf(tw, strings.Join(values, "\t")+"\n")
	}
	return nil
}

func (m resourceTableModel) copyRowToClipboard(row table.Row) {
	var containerInfo string

	// default to docker headers
	headers := []string{
		"NAME",
		"CPU CORES",
		"CPU SHARES",
		"MEM LIMITS",
		"MEM RESERVATIONS",
	}

	// kubernetes rows will always have 6 items
	// change column headers to reflect k8s terms
	if len(row) == 6 {
		headers = []string{
			"NAME",
			"CPU LIMITS",
			"CPU REQUESTS",
			"MEM LIMITS",
			"MEM REQUESTS",
			"CAPACITY",
		}
	}

	for i, header := range headers {
		containerInfo += fmt.Sprintf("%s: %s\n", header, row[i])
	}

	clipboard.WriteAll(containerInfo)
}

func ResourceTable(columns []table.Column, rows []table.Row) {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(14),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := resourceTableModel{t}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
