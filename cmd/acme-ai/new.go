package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"9fans.net/go/acme"
)

func findGitDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for dir != "/" {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}

	return ""
}

func findPrompt() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for dir != "/" {
		promptPath := filepath.Join(dir, ".prompt")
		if content, err := os.ReadFile(promptPath); err == nil {
			prompt := fmt.Sprintf("\n## Prompt\n\n%s\n\n", string(content))
			return strings.TrimSpace(prompt)
		}
		dir = filepath.Dir(dir)
	}

	return ""
}

func findProjectDirectory() (string, error) {
	projectDir := findGitDir()
	if projectDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("could not get current working directory: %w\n", err)
		}

		projectDir = cwd
	}

	return projectDir, nil
}

func generateChatFilename(basePath string) (string, error) {
	const (
		codeLength = 6
		charset    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	)

	for {
		// Generate random bytes
		bytes := make([]byte, codeLength)
		if _, err := rand.Read(bytes); err != nil {
			return "", err
		}

		// Map random bytes to charset
		for i := range bytes {
			bytes[i] = charset[bytes[i]%byte(len(charset))]
		}

		// Construct full filename with path and extension
		fullPath := filepath.Join(basePath, string(bytes)+".md")

		// Check if file exists
		_, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			// File doesn't exist, we can use this path
			return fullPath, nil
		}
		if err != nil {
			return "", err
		}

		// If file exists, loop will continue and generate a new code
	}
}

func actionNew(args []string) {
	prompt := findPrompt()
	projectDir, err := findProjectDirectory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding project directory: %v\n", err)
		os.Exit(1)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}
	chatDirname := filepath.Join(homeDir, ".local/state/acme-utils/llm")

	chatFname, err := generateChatFilename(chatDirname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating chat filename: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(chatDirname, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	frontmatter := fmt.Sprintf("---\nproject_directory: %s\n", projectDir)
	if len(args) >= 1 {
		modelName := args[0]
		frontmatter += fmt.Sprintf("model: %s\n", modelName)
	}
	frontmatter += "---\n"

	content := fmt.Sprintf("%s\n# Title Here\n\n%s\n\n## You [[AiSend]]\n\n", frontmatter, prompt)
	// os.WriteFile(chatFname, []byte(content), 0644)

	win, err := acme.New()
	if err != err {
		fmt.Fprintf(os.Stderr, "Error creating new Acme window: %v\n", err)
		os.Exit(1)
	}

	err = win.Name(chatFname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting Acme window name: %v\n", err)
		os.Exit(1)
	}

	_, err = win.Write("body", []byte(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to Acme window: %v\n", err)
		os.Exit(1)
	}
}
