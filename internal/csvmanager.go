package internal

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/evertras/bubble-table/table"
)

type CSVManager struct {
	FullPath string
	Contents [][]string
	Cursor   *CellPosition
	table    table.Model
}

// Creates mgr and loads from fullPath
func NewCSVManager(fullPath string) (mgr *CSVManager, err error) {
	mgr = &CSVManager{
		FullPath: fullPath,
		Cursor:   &CellPosition{0, 0},
		table:    table.New(nil),
	}

	err = mgr.Load()
	if err != nil {
		return mgr, err
	}

	cols := make([]table.Column, 0, len(mgr.Contents))
	rows := make([]table.Row, 0, len(mgr.Contents[0]))
	for idx, row := range mgr.Contents {
		// Set columns
		if idx == 0 {
			for _, col := range row {
				cols = append(cols, table.NewFlexColumn(strings.TrimSpace(col), col, 1))
			}
			continue
		}

		// Set rows
		rowData := table.RowData{}
		for j, cellData := range row {
			rowData[cols[j].Key()] = cellData
		}
		rows = append(rows, table.NewRow(rowData))
	}

	mgr.table = mgr.table.WithColumns(cols).WithRows(rows).BorderRounded()

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
	case tea.WindowSizeMsg:
		mgr.table.WithTargetWidth(80)
	}

	return mgr, nil
}

func (mgr *CSVManager) View() string {
	s := strings.Builder{}

	s.WriteString(mgr.table.View())

	// for y := 0; y < len(mgr.Contents); y++ {
	// 	for x := 0; x < len(mgr.Contents[0]); x++ {

	// 		if mgr.Cursor.Col == x && mgr.Cursor.Row == y {
	// 			tmp := mgr.Contents[y][x]
	// 			s += lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(tmp)
	// 		} else {
	// 			s += mgr.Contents[y][x]
	// 		}
	// 	}
	// 	s += "\n"
	// }

	s.WriteString("\nctrl+c to quit")
	return s.String()
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
