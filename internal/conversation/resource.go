package conversation

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jaytaylor/html2text"
	"github.com/jcowgar/acme-utils/internal/config"
)

type Resource struct {
	ResourceType string
	Name         string
	Content      string
}

type ResourceRequest interface {
	Fetch(cfg *config.Config, projectDirectory string) ([]Resource, error)
}

type FileResourceRequest struct {
	Filename string
}

type FileGlobResourceRequest struct {
	Pattern string
}

type URLResourceRequest struct {
	URL string
}

func (r FileResourceRequest) Fetch(cfg *config.Config, projectDirectory string) ([]Resource, error) {
	// Should look for an open Acme window with this filename. If found, the content
	// should be taken directly from the buffer as to have the latest content.
	// The content may not have been saved yet. If it is not an open window, then it
	// should be read from disk.

	fullFilename := r.Filename
	if !filepath.IsAbs(fullFilename) {
		f, err := filepath.Abs(filepath.Join(projectDirectory, r.Filename))
		if err != nil {
			return []Resource{}, fmt.Errorf("failed to convert path to absolute: %w", err)
		}

		fullFilename = f
	}

	relativePath, err := filepath.Rel(projectDirectory, fullFilename)
	if err != nil {
		return []Resource{}, fmt.Errorf("failed to convert file path to relative: %w", err)
	}

	data, err := os.ReadFile(fullFilename)
	if err != nil {
		return []Resource{}, fmt.Errorf("failed to read file: %w", err)
	}

	return []Resource{
		Resource{
			ResourceType: "file",
			Name:         relativePath,
			Content:      string(data),
		},
	}, nil
}

func (r FileGlobResourceRequest) Fetch(cfg *config.Config, projectDirectory string) ([]Resource, error) {
	// Find all matching files on the file system and then create/execute many
	// FileResourceRequest statements.

	globPattern := filepath.Join(projectDirectory, r.Pattern)

	glob, err := filepath.Glob(globPattern)
	if err != nil {
		return []Resource{}, fmt.Errorf("could not glob: %w", err)
	}

	resourceRequests := make([]Resource, 0)

fileLoop:
	for _, filename := range glob {
		for _, ignore := range cfg.LLM.GlobIgnore {
			if strings.Contains(filename, ignore) {
				continue fileLoop
			}
		}

		fileRequest := FileResourceRequest{Filename: filename}
		resource, err := fileRequest.Fetch(cfg, projectDirectory)
		if err != nil {
			return []Resource{}, fmt.Errorf("could not fetch file from glob: %w", err)
		}

		resourceRequests = append(resourceRequests, resource[0])
	}

	return resourceRequests, nil
}

func (r URLResourceRequest) Fetch(cfg *config.Config, projectDirectory string) ([]Resource, error) {
	content, err := fetchURLWithTimeout(r.URL)
	if err != nil {
		return []Resource{}, fmt.Errorf("could not fetch URL: %w", err)
	}

	content, err = htmlToText(content)
	if err != nil {
		return []Resource{}, fmt.Errorf("could not convert HTML to text: %w", err)
	}

	return []Resource{
		Resource{ResourceType: "url", Name: r.URL, Content: content},
	}, nil
}

func fetchURLWithTimeout(url string) (string, error) {
	// Create client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set some common headers
	req.Header.Set("User-Agent", "Mozilla/5.0")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

func htmlToText(html string) (string, error) {
	text, err := html2text.FromString(html)
	if err != nil {
		return "", fmt.Errorf("failed to convert HTML to text: %w", err)
	}
	return text, nil
}
