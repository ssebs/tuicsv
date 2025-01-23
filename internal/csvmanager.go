package internal

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/76creates/stickers/flexbox"
	"github.com/76creates/stickers/table"
	"github.com/charmbracelet/bubbles/textinput"
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
	textInput     textinput.Model
	mode          TUIMode
}

const infoText = `Use the arrows to navigate
Type to filter column
Enter to edit selected cell
Ctrl+C to quit
`

// Creates mgr and loads from fullPath
func NewCSVManager(fullPath string) (mgr *CSVManager, err error) {
	mgr = &CSVManager{
		fullPath:      fullPath,
		infoBox:       flexbox.New(0, len(strings.Split(infoText, "\n"))),
		selectedValue: "",
		textInput:     textinput.New(),
		mode:          MODE_NORMAL,
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
	var cmd tea.Cmd

	selectedField := mgr.infoBox.GetRow(0).GetCell(1)

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
		case "escape":
			mgr.mode = MODE_NORMAL
		case "enter", " ":
			if mgr.mode == MODE_NORMAL {
				mgr.mode = MODE_EDIT

				cmd = mgr.textInput.Focus()

				mgr.textInput.SetValue(mgr.table.GetCursorValue())
				mgr.textInput.Prompt = "Edit: "
				mgr.textInput.CursorEnd()

				selectedField.SetContent(mgr.textInput.View())
			} else {
				mgr.mode = MODE_NORMAL

				selectedField.SetContent("")
				mgr.textInput.Blur()

				x, y := mgr.table.GetCursorLocation()
				mgr.SetCell(NewCellPosition(y+1, x), mgr.textInput.Value())

				// Create a [][]any slice
				anySlice := make([][]any, len(mgr.contents[1:]))

				// Convert each []string to []any
				for i, inner := range mgr.contents[1:] {
					anyInner := make([]any, len(inner))
					for j, val := range inner {
						anyInner[j] = val
					}
					anySlice[i] = anyInner
				}

				mgr.table.ClearRows().AddRows(anySlice)
			}
		case "backspace":
			if mgr.mode == MODE_NORMAL {
				mgr.filterWithStr(msg.String())
			} else {
				mgr.textInput, cmd = mgr.textInput.Update(msg)
				selectedField.SetContent(mgr.textInput.View())
			}
		default:
			if mgr.mode == MODE_NORMAL {
				if len(msg.String()) == 1 {
					r := msg.Runes[0]
					if unicode.IsLetter(r) || unicode.IsDigit(r) {
						mgr.filterWithStr(msg.String())
					}
				}
			} else {
				mgr.textInput, cmd = mgr.textInput.Update(msg)
				selectedField.SetContent(mgr.textInput.View())
			}

		}
	case tea.WindowSizeMsg:
		mgr.table.SetWidth(msg.Width)
		mgr.table.SetHeight(msg.Height - mgr.infoBox.GetHeight())
		mgr.infoBox.SetWidth(msg.Width)
	}
	return mgr, cmd
}

func (mgr *CSVManager) View() string {
	if mgr.mode == MODE_NORMAL {
		return lipgloss.JoinVertical(lipgloss.Top, mgr.table.Render(), mgr.infoBox.Render())
	}

	return lipgloss.JoinVertical(lipgloss.Top, mgr.table.Render(), mgr.infoBox.Render())
	// return lipgloss.JoinVertical(lipgloss.Top, mgr.textInput.View(), mgr.table.Render(), mgr.infoBox.Render())
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
func (mgr *CSVManager) SetCell(pos CellPosition, value string) error {
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

	mgr.padCSV()

	return nil
}

func (mgr *CSVManager) padCSV() {
	fillLimit := 100

	if len(mgr.contents) < 100 {
		for idx := len(mgr.contents); idx <= fillLimit; idx++ {
			mgr.contents = append(mgr.contents, make([]string, 1))
		}
	}

}

func (mgr *CSVManager) Init() tea.Cmd {
	return textinput.Blink
}

// 0 indexed
type CellPosition struct {
	Row int
	Col int
}

func NewCellPosition(row int, col int) CellPosition {
	return CellPosition{
		Row: row,
		Col: col,
	}
}

type TUIMode int

const (
	MODE_NORMAL TUIMode = iota
	MODE_EDIT
)
