package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"9fans.net/go/acme"
	ollamaapi "github.com/ollama/ollama/api"

	"github.com/jcowgar/acme-utils/internal/config"
	"github.com/jcowgar/acme-utils/internal/ollama"
)

func main() {
	winIDEnv := os.Getenv("winid")
	if winIDEnv == "" {
		log.Printf("must be ran from within Acme, `winid` not set.\n")
		return
	}

	winID, err := strconv.Atoi(winIDEnv)
	if err != nil {
		log.Printf("could not convert %v into an integer winID\n", err)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		log.Printf("failed to load configuration: %v\n", err)
		return
	}

	baseURL, err := url.Parse(cfg.Ollama.BaseURL)
	if err != nil {
		log.Printf("failed to parse URL: %v\n", err)
		return
	}

	client := ollamaapi.NewClient(baseURL, http.DefaultClient)

	if err := MaybeTalkToOllama(client, winID); err != nil {
		log.Printf("error processing ollama request for window %d: %v\n", winID, err)
	}
}

func MaybeTalkToOllama(client *ollamaapi.Client, winID int) error {
	win, err := acme.Open(winID, nil)
	if err != nil {
		return fmt.Errorf("could not open winID %d: %w", winID, err)
	}

	// Get the current window size
	err = win.Addr("0,$")
	if err != nil {
		return fmt.Errorf("failed to set addr to full content: %w", err)
	}

	bodyBytes := make([]byte, 256*1024)
	n, err := win.Read("body", bodyBytes)
	if err != nil {
		return fmt.Errorf("could not read the body of winID %d: %w", winID, err)
	}

	content := string(bodyBytes[:n])
	conversation, err := ollama.ParseContent(content)
	if err != nil {
		return fmt.Errorf("could not parse conversation content: %w", err)
	}

	// Get the last message and check if it's empty
	lastMessage, err := conversation.GetLastUserMessage()
	if err != nil {
		return fmt.Errorf("could not get last user message: %w", err)
	}

	// If the last message is empty or only whitespace, return early
	if strings.TrimSpace(lastMessage) == "" {
		return nil
	}

	// Write immediately to give the user some feedback
	_, err = win.Write("body", []byte("\n\n### Response... thinking..."))
	if err != nil {
		return fmt.Errorf("failed to write response back to window: %w", err)
	}

	// Insert files into the conversation, if the user requested them.
	if conversation.IncludeFiles {
		// Get all windows from Acme
		windows, err := acme.Windows()
		if err == nil { // Don't fail if we can't access Acme
			for _, winInfo := range windows {
				// Skip if it's a directory
				if winInfo.IsDir {
					continue
				}

				// Skip if it is our AI chat file
				if winInfo.ID == winID {
					continue
				}

				// Skip if filename is empty
				if winInfo.Name == "" {
					continue
				}

				// Skip if the name is not a Normal filename. Acme has buffers such as +win that
				// are not real files. We do not wish to include that content.
				if !strings.Contains(winInfo.Name, ".") || strings.HasPrefix(winInfo.Name, "+") {
					continue
				}

				log.Printf("including file: %v\n", winInfo.Name)

				// Open the window
				win, err := acme.Open(winInfo.ID, nil)
				if err != nil {
					continue
				}
				defer win.CloseFiles()

				// Read the content of the window
				content, err := win.ReadAll("body")
				if err != nil {
					continue
				}

				// Add the file to our conversation
				conversation.AddFile(winInfo.Name, string(content))
			}
		}
	}

	// Convert all messages to Ollama API format
	messages := make([]ollamaapi.Message, 0, len(conversation.Messages))
	filesInserted := false // Local flag to track if files have been inserted

	for _, msg := range conversation.Messages {
		role := "user"
		if msg.Role == "Response" {
			role = "assistant"
		}

		content := msg.Content
		// If this is a user message containing "+files" and we haven't inserted files yet
		if conversation.IncludeFiles &&
			!filesInserted &&
			role == "user" &&
			strings.Contains(msg.Content, "+files") &&
			len(conversation.Files) > 0 {

			var filesSection strings.Builder
			filesSection.WriteString("\n\n# Relevant Files\n\n")

			for _, file := range conversation.Files {
				filesSection.WriteString(fmt.Sprintf("Filename: %s\n```\n%s\n```\n\n",
					file.Name,
					file.Content))
			}

			content += filesSection.String()
			filesInserted = true // Mark that we've inserted the files
		}

		messages = append(messages, ollamaapi.Message{
			Role:    role,
			Content: content,
		})
	}

	stream := false
	req := &ollamaapi.ChatRequest{
		Model:    conversation.Model,
		Messages: messages,
		Stream:   &stream,
	}

	var response *ollamaapi.ChatResponse
	responseHandler := func(r ollamaapi.ChatResponse) error {
		response = &r
		return nil
	}

	err = client.Chat(context.Background(), req, responseHandler)
	if err != nil {
		return fmt.Errorf("failed to get response from Ollama: %w", err)
	}

	// Clear the current content
	err = win.Addr("0,$")
	if err != nil {
		return fmt.Errorf("failed to set addr to full content: %w", err)
	}

	// Write the full conversation with the new response
	conversation.AddResponse(response.Message.Content)
	newContent := conversation.String()

	// Add the new "You" section
	newContent += "## You [[OllamaSend]]\n\n"

	_, err = win.Write("data", []byte(newContent))
	if err != nil {
		return fmt.Errorf("failed to write response back to window: %w", err)
	}

	// Move the cursor to the end of the file
	err = win.Addr("$")
	if err != nil {
		return fmt.Errorf("failed to move cursor to end: %w", err)
	}

	// Set the dot to the end of the file
	_, err = win.Write("ctl", []byte("dot=addr"))
	if err != nil {
		return fmt.Errorf("failed to set dot to end: %w", err)
	}

	// Show the cursor at the end
	_, err = win.Write("ctl", []byte("show"))
	if err != nil {
		return fmt.Errorf("failed to show cursor: %w", err)
	}

	// Save the file
	_, err = win.Write("ctl", []byte("put"))
	if err != nil {
		return fmt.Errorf("failed to save the file: %w", err)
	}

	return nil
}
