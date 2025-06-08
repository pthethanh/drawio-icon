package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pthethanh/drawio-icon/iconify"
	"github.com/pthethanh/drawio-icon/kw"
	"github.com/pthethanh/drawio-icon/lib"
)

func main() {
	query := flag.String("query", "json", "query")
	limit := flag.Int("limit", 100, "query limit")
	outDir := flag.String("outputDir", "output", "output directory")

	flag.Parse()

	iconDir, err := os.MkdirTemp(os.TempDir(), *query)
	if err != nil {
		log.Panic(err)
	}
	if err := os.MkdirAll(*outDir, os.ModePerm); err != nil {
		log.Panic(err)
	}
	kws, err := kw.GetRelevantKeywords(*query)
	if err != nil {
		log.Panic(err)
	}
	log.Println("search keywords:", *query)
	log.Println("optimized keywords:", kws)
	for _, q := range kws {
		q := url.QueryEscape(strings.TrimSpace(q))
		if q == "" {
			continue
		}
		if err := iconify.Search(iconDir, q, *limit); err != nil {
			log.Println(err)
		}
	}
	outName, _, _ := strings.Cut(*query, ",")
	outputLibFile := filepath.Join(*outDir, fmt.Sprintf("%s.xml", outName))
	if err := lib.Generate(outputLibFile, iconDir); err != nil {
		panic(err)
	}
}
