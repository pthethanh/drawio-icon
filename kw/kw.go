package kw

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

func ensureOllamaRunning(ctx context.Context) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Heartbeat(context.Background()); err == nil {
		return nil
	}
	// Not running, start it
	cmd := exec.CommandContext(ctx, "ollama", "serve")

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ollama: %w", err)
	}

	// Wait for readiness
	log.Println("waiting for ollama to become ready...")
	for i := 0; i < 20; i++ {
		time.Sleep(500 * time.Millisecond)
		if err := client.Heartbeat(context.Background()); err == nil {
			log.Println("ollama is ready")
			return nil
		}
	}
	return fmt.Errorf("ollama did not become ready")
}

func ensureModelExists(model string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	rs, err := client.List(context.Background())
	if err != nil {
		return err
	}
	for _, m := range rs.Models {
		if m.Name == model {
			return nil
		}
	}
	log.Println("model not found locally, pulling ", model)
	if err := client.Pull(context.Background(), &api.PullRequest{
		Model: model,
	}, func(pr api.ProgressResponse) error {
		log.Printf("pull progress: %.2f%%\n", pr.Completed/pr.Total*100)
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func generate(model, prompt string) (string, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	req := &api.GenerateRequest{
		Model:  model,
		Prompt: prompt,
		// set streaming to false
		Stream: new(bool),
		Think:  &api.ThinkValue{Value: false},
	}
	ctx := context.Background()
	res := ""
	respFunc := func(resp api.GenerateResponse) error {
		res = resp.Response
		return nil
	}
	err = client.Generate(ctx, req, respFunc)
	if err != nil {
		log.Fatal(err)
	}
	return res, nil
}

func GetRelevantKeywords(model, userQuery string) ([]string, error) {
	ctx := context.Background()

	//sEnsure Ollama server is running
	if err := ensureOllamaRunning(ctx); err != nil {
		return fallbackKeywords(userQuery), nil
	}

	// Ensure model exists locally
	if err := ensureModelExists(model); err != nil {
		return fallbackKeywords(userQuery), nil
	}

	prompt := fmt.Sprintf(`
You are an icon keyword expansion engine.

Task:
Generate related keywords for searching icons.

Rules:
- Return ONLY comma-separated keywords
- No explanations
- No markdown
- No numbering
- No categories
- No duplicate keywords
- Use lowercase only
- Maximum 10 keywords total
- Keep keywords short
- Prefer icon-library-friendly terms
- Include:
  - synonyms
  - technical concepts
  - visual metaphors
  - symbolic objects
  - UI concepts

Examples:
ai → robot,brain,chip,sparkles,automation
security → shield,lock,fingerprint
analytics → chart,graph,dashboard
cloud → server,upload,storage

User query:
%s

Output:
`, userQuery)
	response, err := generate(model, prompt)
	if err != nil {
		return fallbackKeywords(userQuery), nil
	}
	response = strings.TrimSpace(response)
	parts := strings.Split(response, ",")

	var keywords []string

	for _, p := range parts {
		kw := strings.TrimSpace(p)

		if kw != "" {
			keywords = append(keywords, kw)
		}
	}

	// Fallback if model returned garbage
	if len(keywords) == 0 {
		return fallbackKeywords(userQuery), nil
	}

	return keywords, nil
}

func fallbackKeywords(userQuery string) []string {
	parts := strings.Split(userQuery, ",")

	var result []string

	for _, p := range parts {
		p = strings.TrimSpace(p)

		if p != "" {
			result = append(result, p)
		}
	}

	return result
}
