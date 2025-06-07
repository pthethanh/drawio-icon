package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pthethanh/drawio-icon/iconify"
	"github.com/pthethanh/drawio-icon/lib"
)

func main() {
	query := flag.String("query", "json", "query")
	queryLimit := flag.Int("limit", 25, "query limit")
	outDir := flag.String("outputDir", "output", "output directory")

	flag.Parse()

	downloadDir, err := os.MkdirTemp(os.TempDir(), *query)
	if err != nil {
		log.Panic(err)
	}
	if err := os.MkdirAll(*outDir, os.ModePerm); err != nil {
		log.Panic(err)
	}
	for _, q := range strings.Split(*query, ",") {
		encodedQuery := strings.ReplaceAll(strings.TrimSpace(q), " ", "+")
		if encodedQuery == "" {
			continue
		}
		if err := iconify.Search(downloadDir, encodedQuery, *queryLimit); err != nil {
			log.Println(err)
		}
	}
	outName, _, _ := strings.Cut(*query, ",")
	outputLibFile := filepath.Join(*outDir, fmt.Sprintf("%s.xml", outName))
	if err := lib.Generate(downloadDir, outputLibFile); err != nil {
		panic(err)
	}
}
