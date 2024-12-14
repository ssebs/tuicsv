package internal

import (
	"encoding/csv"
	"fmt"
	"os"
)

type CSVManager struct {
	FullPath string
	Contents [][]string
}

// Creates mgr and loads from fullPath
func NewCSVManager(fullPath string) (mgr *CSVManager, err error) {
	mgr = &CSVManager{
		FullPath: fullPath,
	}

	err = mgr.Load()
	return mgr, err
}

// Set the contents of one cell, 0 indexed
func (mgr *CSVManager) UpdateCell(col, row int, value string) error {
	if col < 0 || row < 0 || row > len(mgr.Contents) || col > len(mgr.Contents[0]) {
		return fmt.Errorf("x/y out of bounds of Contents")
	}

	mgr.Contents[row][col] = value
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
