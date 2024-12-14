package main

import (
	"fmt"
	"log"
	"path/filepath"

	flag "github.com/spf13/pflag"
	"github.com/ssebs/tuicsv/internal"
)

func getFullPathFromFlags() string {
	var shortPath string
	flag.StringVarP(&shortPath, "path", "p", "", "Path to csv file")
	flag.Parse()

	if shortPath == "" {
		flag.Usage()
		log.Fatal("please provide a path to a csv file.")
	}

	csvPath, err := filepath.Abs(shortPath)
	if err != nil {
		flag.Usage()
		log.Fatal(fmt.Errorf("is %s a path to a csv file?", shortPath))
	}

	if filepath.Ext(csvPath) != ".csv" {
		flag.Usage()
		log.Fatal(fmt.Errorf("is %s a path to a csv file?", shortPath))
	}

	return csvPath
}

func main() {
	mgr, err := internal.NewCSVManager(getFullPathFromFlags())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(mgr.Contents)
}
