package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pthethanh/drawio-icon/drawio"
	"github.com/pthethanh/drawio-icon/iconify"
	"github.com/pthethanh/drawio-icon/keyword"
)

func main() {
	query := flag.String("query", "json", "query")
	limit := flag.Int("limit", 100, "query limit")
	model := flag.String("model", "qwen3.5:0.8b", "ollama model to use")
	combine := flag.Bool("combine", false, "combine all keywords together in a single file output?")
	outDir := flag.String("outputDir", "output", "output directory")

	flag.Parse()

	dirName, _, _ := strings.Cut(*query, ",")
	iconDir := ""
	if *combine {
		dir, err := os.MkdirTemp(os.TempDir(), dirName)
		if err != nil {
			log.Panic(err)
		}
		iconDir = dir
	}

	if err := os.MkdirAll(*outDir, os.ModePerm); err != nil {
		log.Panic(err)
	}
	kws, err := keyword.Generate(*model, *query)
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
		if !*combine {
			dir, err := os.MkdirTemp(os.TempDir(), q)
			if err != nil {
				log.Panic(err)
			}
			iconDir = dir
		}
		if err := iconify.Search(iconDir, q, *limit); err != nil {
			log.Println(err)
		}
		if !*combine {
			outputLibFile := filepath.Join(*outDir, fmt.Sprintf("%s.xml", q))
			if err := drawio.GenerateLib(outputLibFile, iconDir); err != nil {
				log.Panic(err)
			}
			_ = os.RemoveAll(iconDir)
		}
	}
	if *combine {
		outputLibFile := filepath.Join(*outDir, fmt.Sprintf("%s.xml", dirName+"_combine"))
		if err := drawio.GenerateLib(outputLibFile, iconDir); err != nil {
			log.Panic(err)
		}
	}
}
