package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"9fans.net/go/acme"

	"github.com/jcowgar/acme-utils/internal/config"
	"github.com/jcowgar/acme-utils/internal/conversation"
	"github.com/jcowgar/acme-utils/internal/llm"
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

	win, err := acme.Open(winID, nil)
	if err != nil {
		log.Printf("could not open winID %d: %w", winID, err)
		return
	}

	conv, err := readConversation(win, winID)
	if err != nil {
		log.Printf("failed to read conversation: %v\n", err)
		return
	}
	if conv == nil {
		// There is no new conversation data, ignore this request
		return
	}

	model := conv.Model
	if model == "" {
		model = cfg.LLM.DefaultProvider
	}

	providerConfig := cfg.LLM.Providers[model]
	provider, err := llm.NewProvider(providerConfig.Type, providerConfig)
	if err != nil {
		log.Printf("failed to create provider: %v\n", err)
		return
	}

	if err := sendLLMRequest(provider, win, winID, conv); err != nil {
		log.Printf("error processing LLM request for window %d: %v\n", winID, err)
	}
}

func readConversation(win *acme.Win, winID int) (*conversation.Conversation, error) {
	err := win.Addr("0,$")
	if err != nil {
		return nil, fmt.Errorf("failed to set addr to full content: %w", err)
	}

	bodyBytes := make([]byte, 256*1024)
	n, err := win.Read("body", bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("could not read the body of winID %d: %w", winID, err)
	}

	content := string(bodyBytes[:n])
	conv, err := conversation.ParseContent(content)
	if err != nil {
		return nil, fmt.Errorf("could not parse conversation content: %w", err)
	}

	lastMessage, err := conv.GetLastUserMessage()
	if err != nil {
		return nil, fmt.Errorf("could not get last user message: %w", err)
	}

	// If the last message is empty, return early, we have nothing really to do or
	// send to the LLM
	if strings.TrimSpace(lastMessage) == "" {
		return nil, nil
	}

	winName := getProjectDirectory(winID)

	for _, req := range conv.ResourceRequests {
		resources, err := req.Fetch(winName)
		if err != nil {
			return nil, fmt.Errorf("could not fetch resource: %w", err)
		}

		for _, resource := range resources {
			conv.AddReferenceMaterial(resource.ResourceType, resource.Name, resource.Content)
		}
	}

	includeFiles(conv, winID)

	return conv, nil
}

func getProjectDirectory(winID int) string {
	windows, err := acme.Windows()
	if err != nil {
		// Don't fail if we can't access Acme
		return ""
	}

	for _, winInfo := range windows {
		if winInfo.ID == winID {
			return extractProjectDir(winInfo.Name)
		}
	}

	return ""
}

func extractProjectDir(filename string) string {
	// Split by double underscore to separate components
	parts := strings.Split(strings.TrimSuffix(filename, ".md"), "__")

	// Remove the first part (timestamp)
	if len(parts) < 2 {
		return "" // Return empty string if format is invalid
	}

	// Join the middle parts with "/" but handle the last part specially
	projectParts := parts[1:]
	lastPart := projectParts[len(projectParts)-1]

	// Remove the .md extension from the last part
	lastPart = strings.TrimSuffix(lastPart, ".md")
	projectParts[len(projectParts)-1] = lastPart

	// Join all parts with "/"
	projectPath := "/" + strings.Join(projectParts, "/")

	return projectPath
}

func includeFiles(conv *conversation.Conversation, winID int) {
	if !conv.IncludeFiles {
		return
	}

	windows, err := acme.Windows()
	if err != nil {
		// Don't fail if we can't access Acme
		return
	}

	for _, winInfo := range windows {
		if shouldSkipFile(&winInfo, winID) {
			continue
		}

		win, err := acme.Open(winInfo.ID, nil)
		if err != nil {
			continue
		}
		defer win.CloseFiles()

		content, err := win.ReadAll("body")
		if err != nil {
			continue
		}

		conv.AddReferenceMaterial(conversation.ReferenceTypeFile, winInfo.Name, string(content))
	}
}

func shouldSkipFile(winInfo *acme.WinInfo, winID int) bool {
	return winInfo.IsDir || winInfo.ID == winID || inIgnoreFilenames(winInfo.Name)
}

func sendLLMRequest(provider llm.Provider, win *acme.Win, winID int, conv *conversation.Conversation) error {
	// Write immediately to give the user some feedback
	_, err := win.Write("body", []byte("\n\n### Response... thinking..."))
	if err != nil {
		return fmt.Errorf("failed to write response back to window: %w", err)
	}

	// Convert messages to provider format
	messages := make([]llm.Message, 0, len(conv.Messages))
	filesInserted := false

	for _, msg := range conv.Messages {
		role := "user"
		if msg.Role == "Response" {
			role = "assistant"
		}

		content := msg.Content
		if !filesInserted && role == "user" && len(conv.ReferenceMaterial) > 0 {
			var filesSection strings.Builder

			filesSection.WriteString("\n\n# Relevant Material\n\n")

			for _, file := range conv.ReferenceMaterial {
				filesSection.WriteString(fmt.Sprintf("Reference Material Type: %s\nName: %s\n```\n%s\n```\n\n",
					file.Typ,
					file.Name,
					file.Content))
			}

			content += filesSection.String()

			filesInserted = true
		}

		messages = append(messages, llm.Message{
			Role:    role,
			Content: content,
		})
	}

	response, err := provider.Chat(context.Background(), messages)
	if err != nil {
		return fmt.Errorf("failed to get response from provider: %w", err)
	}

	// Clear the current content
	err = win.Addr("0,$")
	if err != nil {
		return fmt.Errorf("failed to set addr to full content: %w", err)
	}

	// Write the full conversation with the new response
	conv.AddResponse(response)
	newContent := conv.String()

	// Add the new "You" section
	newContent += "## You [[LlmSend]]\n\n"

	_, err = win.Write("data", []byte(newContent))
	if err != nil {
		return fmt.Errorf("failed to write response back to window: %w", err)
	}

	// Save the file
	_, err = win.Write("ctl", []byte("put"))
	if err != nil {
		return fmt.Errorf("failed to save the file: %w", err)
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

	// Search backward to the ### Response
	err = win.Addr("?### Response?")
	if err != nil {
		return fmt.Errorf("failed to move cursor to the last Response: %w", err)
	}

	// Set the dot to the end of the file
	_, err = win.Write("ctl", []byte("dot=addr"))
	if err != nil {
		return fmt.Errorf("failed to set dot to end: %w", err)
	}

	// Be sure to make this selection visible
	_, err = win.Write("ctl", []byte("show"))
	if err != nil {
		return fmt.Errorf("failed to show cursor: %w", err)
	}

	return nil
}

func inIgnoreFilenames(s string) bool {
	ignoreAnywhere := []string{"+dirtree", "+watch", "+win", "+Errors"}
	ignoreJustFilename := []string{"", "guide"}

	for _, item := range ignoreAnywhere {
		if strings.Contains(s, item) {
			return true
		}
	}

	filename := filepath.Base(s)
	for _, item := range ignoreJustFilename {
		if item == filename {
			return true
		}
	}

	return false
}
