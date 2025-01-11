package internal

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/76creates/stickers/flexbox"
	"github.com/76creates/stickers/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CSVManager struct {
	fullPath      string
	contents      [][]string
	container     *flexbox.FlexBox
	table         *table.TableSingleType[string]
	infoBox       *flexbox.FlexBox
	headers       []string
	selectedValue string
	filterTxt     string
}

const infoText = `Use the arrows to navigate
Type to filter column
Enter, spacebar: get column value
Ctrl+C: quit
`

// Creates mgr and loads from fullPath
func NewCSVManager(fullPath string) (mgr *CSVManager, err error) {
	mgr = &CSVManager{
		fullPath:      fullPath,
		infoBox:       flexbox.New(0, len(strings.Split(infoText, "\n"))),
		selectedValue: "Select something with Enter",
		filterTxt:     "",
		// container:     flexbox.New(0, 7),
	}

	err = mgr.Load()
	if err != nil {
		return mgr, err
	}

	// TODO: move headers and contents
	mgr.headers = mgr.contents[0]
	mgr.table = table.NewTableSingleType[string](0, 0, mgr.headers)
	mgr.table.SetStylePassing(true)

	mgr.table.SetStyles(map[table.TableStyleKey]lipgloss.Style{
		table.TableHeaderStyleKey: lipgloss.NewStyle().
			Background(lipgloss.Color("#024b8a")).
			Foreground(lipgloss.Color("#fff")),
		table.TableFooterStyleKey: lipgloss.NewStyle().
			// Background(lipgloss.Color("#222")).
			Foreground(lipgloss.Color("#fff")).Align(lipgloss.Right).Height(1),
		table.TableRowsStyleKey: lipgloss.NewStyle().
			// Background(lipgloss.Color("#222")).
			Foreground(lipgloss.Color("#fff")),
		table.TableRowsSubsequentStyleKey: lipgloss.NewStyle().
			// Background(lipgloss.Color("#444")).
			Foreground(lipgloss.Color("#fff")),
		table.TableRowsCursorStyleKey: lipgloss.NewStyle().
			// Background(lipgloss.Color("#333")).
			Foreground(lipgloss.Color("#fff")),
		table.TableCellCursorStyleKey: lipgloss.NewStyle().
			Background(lipgloss.Color("#024b8a")).
			Foreground(lipgloss.Color("#fff")).
			Bold(true),
	})

	mgr.table.AddRows(mgr.contents[1:])

	mgr.infoBox.AddRows([]*flexbox.Row{
		mgr.infoBox.NewRow().AddCells(
			flexbox.NewCell(1, 1).
				SetID("info").
				SetContent(infoText).
				SetStyle(lipgloss.NewStyle().PaddingLeft(1)),
			flexbox.NewCell(2, 1).
				SetID("val").
				SetContent(mgr.selectedValue).
				SetStyle(lipgloss.NewStyle().Bold(true)),
		).SetStyle(lipgloss.NewStyle().Border(lipgloss.NormalBorder())),
	})

	return mgr, err
}

func (mgr *CSVManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			mgr.table.CursorUp()
		case "down", "j":
			mgr.table.CursorDown()
		case "left", "h":
			mgr.table.CursorLeft()
		case "right", "l":
			mgr.table.CursorRight()
		case "ctrl+c":
			return mgr, tea.Quit
		case "enter", " ":
			// switch mode to edit

			mgr.selectedValue = "Selected: " + mgr.table.GetCursorValue()
			mgr.infoBox.GetRow(0).GetCell(1).SetContent(mgr.selectedValue + mgr.filterTxt)
		case "backspace":
			mgr.filterWithStr(msg.String())
		default:
			if len(msg.String()) == 1 {
				r := msg.Runes[0]
				if unicode.IsLetter(r) || unicode.IsDigit(r) {
					mgr.filterWithStr(msg.String())
				}
			}

		}
	case tea.WindowSizeMsg:
		mgr.table.SetWidth(msg.Width)
		mgr.table.SetHeight(msg.Height - mgr.infoBox.GetHeight())
		mgr.infoBox.SetWidth(msg.Width)
	}

	return mgr, nil
}

func (mgr *CSVManager) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, mgr.table.Render(), mgr.infoBox.Render())
}

func (m *CSVManager) filterWithStr(key string) {
	i, s := m.table.GetFilter()
	x, _ := m.table.GetCursorLocation()
	if x != i && key != "backspace" {
		m.table.SetFilter(x, key)
		return
	}
	if key == "backspace" {
		if len(s) == 1 {
			m.table.UnsetFilter()
			return
		} else if len(s) > 1 {
			s = s[0 : len(s)-1]
		} else {
			return
		}
	} else {
		s = s + key
	}
	m.table.SetFilter(i, s)
}

// Set the contents of one cell by a 0 indexed CellPosition
// err if out of bounds
func (mgr *CSVManager) SetCell(pos *CellPosition, value string) error {
	if pos.Col < 0 || pos.Row < 0 || pos.Row > len(mgr.contents) || pos.Col > len(mgr.contents[0]) {
		return fmt.Errorf("x/y out of bounds of Contents")
	}

	mgr.contents[pos.Row][pos.Col] = value
	return nil
}

// Save csv mgr.Contents to mgr.FullPath
func (mgr *CSVManager) Save() error {
	f, err := os.Create(mgr.fullPath)
	if err != nil {
		return fmt.Errorf("failed to open csv, %s", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)

	if err := w.WriteAll(mgr.contents); err != nil {
		return fmt.Errorf("failed to write fiale, %s", err)
	}
	return nil
}

// Load from mgr.FullPath to csv mgr.Contents
func (mgr *CSVManager) Load() error {
	f, err := os.Open(mgr.fullPath)
	if err != nil {
		return fmt.Errorf("failed to open csv, %s", err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	mgr.contents, err = r.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to parse csv, %s", err)
	}
	return nil
}

func (mgr *CSVManager) Init() tea.Cmd {
	return nil
}

// 0 indexed
type CellPosition struct {
	Row int
	Col int
}
