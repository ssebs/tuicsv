package internal

import (
	"encoding/csv"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CSVManager struct {
	FullPath string
	Contents [][]string
	Cursor   *CellPosition
}

// Creates mgr and loads from fullPath
func NewCSVManager(fullPath string) (mgr *CSVManager, err error) {
	mgr = &CSVManager{
		FullPath: fullPath,
		Cursor:   &CellPosition{0, 0},
	}

	err = mgr.Load()
	return mgr, err
}

func (mgr *CSVManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return mgr, tea.Quit
		case "up", "k":
			if mgr.Cursor.Row > 0 {
				mgr.Cursor.Row--
			}
		case "down", "j":
			if mgr.Cursor.Row < len(mgr.Contents)-1 {
				mgr.Cursor.Row++
			}
		case "left", "h":
			if mgr.Cursor.Col > 0 {
				mgr.Cursor.Col--
			}
		case "right", "l":
			if mgr.Cursor.Col < len(mgr.Contents[0])-1 {
				mgr.Cursor.Col++
			}
		case "enter":
			// edit

		}
	}

	return mgr, nil
}

func (mgr *CSVManager) View() string {
	s := ""

	for y := 0; y < len(mgr.Contents); y++ {
		for x := 0; x < len(mgr.Contents[0]); x++ {

			if mgr.Cursor.Col == x && mgr.Cursor.Row == y {
				tmp := mgr.Contents[y][x]
				s += lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(tmp)
			} else {
				s += mgr.Contents[y][x]
			}
		}
		s += "\n"
	}

	s += "ctrl+c to quit"
	return s
}

// Set the contents of one cell by a 0 indexed CellPosition
func (mgr *CSVManager) UpdateCell(pos *CellPosition, value string) error {
	if pos.Col < 0 || pos.Row < 0 || pos.Row > len(mgr.Contents) || pos.Col > len(mgr.Contents[0]) {
		return fmt.Errorf("x/y out of bounds of Contents")
	}

	mgr.Contents[pos.Row][pos.Col] = value
	return nil
}

func (mgr *CSVManager) Save() error {
	f, err := os.Create(mgr.FullPath)
	if err != nil {
		return fmt.Errorf("failed to open csv, %s", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)

	if err := w.WriteAll(mgr.Contents); err != nil {
		return fmt.Errorf("failed to write fiale, %s", err)
	}
	return nil
}

func (mgr *CSVManager) Load() error {
	f, err := os.Open(mgr.FullPath)
	if err != nil {
		return fmt.Errorf("failed to open csv, %s", err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	mgr.Contents, err = r.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to parse csv, %s", err)
	}
	return nil
}

func (mgr *CSVManager) Init() tea.Cmd {
	return nil
}

type CellPosition struct {
	Row int
	Col int
}
