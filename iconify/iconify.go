package iconify

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	searchURL   = "https://api.iconify.design/search?query=%s&limit=%d"
	downloadURL = "https://api.iconify.design/%s.svg"
)

// Struct to parse the search result
type searchResult struct {
	Icons []string `json:"icons"`
}

func Search(outputDir string, query string, limit int) error {
	// Step 1: Perform the search
	searchURL := fmt.Sprintf(searchURL, query, limit)
	log.Println("searching ", searchURL)
	resp, err := http.Get(searchURL)
	if err != nil {
		log.Println("Failed to fetch search results:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Unexpected HTTP status:", resp.Status)
		return fmt.Errorf("invalid http status")
	}

	var result searchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println("Failed to decode JSON:", err)
		return err
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Println("Failed to create output directory:", err)
		return err
	}

	// Step 2: Download each icon
	for _, icon := range result.Icons {
		err := downloadSVG(outputDir, icon)
		if err != nil {
			log.Printf("Failed to download %s: %v\n", icon, err)
		}
	}
	return nil
}

func downloadSVG(outputDir string, icon string) error {
	url := fmt.Sprintf(downloadURL, icon)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error fetching %s: %w", icon, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 response for %s: %s", icon, resp.Status)
	}

	filename := strings.ReplaceAll(icon, ":", "_") + ".svg"
	filePath := filepath.Join(outputDir, filename)

	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", filePath, err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("error saving %s: %w", icon, err)
	}

	log.Printf("Downloaded: %s\n", filename)
	return nil
}
