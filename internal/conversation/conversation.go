package conversation

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	ReferenceTypeFile          = "file"
	ReferenceTypeURL           = "url"
	ReferenceTypeCommandOutput = "command_output"
)

// ReferenceMaterial represents attached content to the conversation
type ReferenceMaterial struct {
	Name    string
	Content string
	Typ     string
}

// Conversation represents the entire chat interaction
type Conversation struct {
	Title             string
	Prompt            string
	Model             string
	ProjectDirectory  string
	Parameters        map[string]interface{}
	Messages          []Message
	ReferenceMaterial []ReferenceMaterial
	IncludeFiles      bool
	ResourceRequests  []ResourceRequest
}

// Message represents a single message in the conversation
type Message struct {
	Role      string    // "user" or "assistant"
	Content   string    // The actual message content
	Timestamp time.Time // Optional, for future use
}

// AddReferenceMaterial adds a new resource to the conversation
func (c *Conversation) AddReferenceMaterial(typ string, name string, content string) {
	if c.ReferenceMaterial == nil {
		c.ReferenceMaterial = make([]ReferenceMaterial, 0)
	}
	c.ReferenceMaterial = append(c.ReferenceMaterial, ReferenceMaterial{
		Name:    name,
		Content: content,
		Typ:     typ,
	})
}

// ParseContent parses the markdown content and returns a Conversation
func ParseContent(content string) (*Conversation, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	conv := &Conversation{
		Messages:          make([]Message, 0),
		Parameters:        make(map[string]interface{}),
		ReferenceMaterial: make([]ReferenceMaterial, 0),
		IncludeFiles:      false,
	}

	var currentRole string
	var currentContent strings.Builder
	inFrontMatter := false

	for scanner.Scan() {
		line := scanner.Text()

		// Handle front matter
		if line == "---" {
			if !inFrontMatter {
				inFrontMatter = true
				continue
			} else {
				inFrontMatter = false
				continue
			}
		}

		if inFrontMatter {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if key == "model" {
					conv.Model = value
				} else if key == "project_directory" {
					conv.ProjectDirectory = value
				} else {
					conv.Parameters[key] = value
				}
			}
			continue
		}

		// Handle resource requests
		if strings.HasPrefix(line, "+file ") {
			conv.ResourceRequests = append(conv.ResourceRequests, FileResourceRequest{
				Filename: strings.TrimPrefix(line, "+file "),
			})
		} else if strings.HasPrefix(line, "+url ") {
			conv.ResourceRequests = append(conv.ResourceRequests, URLResourceRequest{
				URL: strings.TrimPrefix(line, "+url "),
			})
		} else if strings.HasPrefix(line, "+glob") {
			conv.ResourceRequests = append(conv.ResourceRequests, FileGlobResourceRequest{
				Pattern: strings.TrimPrefix(line, "+glob "),
			})
		}

		// Handle title (first level heading)
		if strings.HasPrefix(line, "# ") {
			conv.Title = strings.TrimPrefix(line, "# ")
			continue
		}

		// Handle message start (second level heading)
		if strings.HasPrefix(line, "## You") {
			// Save previous message if exists
			if currentRole != "" && currentContent.Len() > 0 {
				conv.Messages = append(conv.Messages, Message{
					Role:      currentRole,
					Content:   strings.TrimSpace(currentContent.String()),
					Timestamp: time.Now(),
				})
				currentContent.Reset()
			}
			currentRole = "You"
			continue
		}

		// Handle response (third level heading)
		if strings.HasPrefix(line, "### Response") {
			// Save previous message if exists
			if currentRole != "" && currentContent.Len() > 0 {
				messageContent := strings.TrimSpace(currentContent.String())
				conv.Messages = append(conv.Messages, Message{
					Role:      currentRole,
					Content:   messageContent,
					Timestamp: time.Now(),
				})
				// Check for "+files" in user messages
				if currentRole == "You" && strings.Contains(messageContent, "+files") {
					conv.IncludeFiles = true
				}
				currentContent.Reset()
			}
			currentRole = "Response"
			continue
		}

		// Accumulate content if we're inside a message
		if currentRole != "" {
			currentContent.WriteString(line + "\n")
		}
	}

	// Add the last message if exists
	if currentRole != "" && currentContent.Len() > 0 {
		messageContent := strings.TrimSpace(currentContent.String())
		conv.Messages = append(conv.Messages, Message{
			Role:      currentRole,
			Content:   messageContent,
			Timestamp: time.Now(),
		})
		// Check for "+files" in the last user message
		if currentRole == "You" && strings.Contains(messageContent, "+files") {
			conv.IncludeFiles = true
		}
	}

	if len(conv.Messages) == 0 {
		return nil, errors.New("no messages found in content")
	}

	return conv, nil
}

// GetLastUserMessage returns the content of the last user message
func (c *Conversation) GetLastUserMessage() (string, error) {
	for i := len(c.Messages) - 1; i >= 0; i-- {
		if c.Messages[i].Role == "You" {
			return c.Messages[i].Content, nil
		}
	}

	return "", errors.New("no user messages found")
}

// AddResponse adds a new response message to the conversation
func (c *Conversation) AddResponse(content string) {
	c.Messages = append(c.Messages, Message{
		Role:      "Response",
		Content:   content,
		Timestamp: time.Now(),
	})
}

// String converts the conversation back to markdown format
func (c *Conversation) String() string {
	var sb strings.Builder

	// Write front matter only if model or parameters exist
	if c.Model != "" || len(c.Parameters) > 0 {
		sb.WriteString("---\n")
		if c.Model != "" {
			sb.WriteString(fmt.Sprintf("model: %s\n", c.Model))
		}
		for key, value := range c.Parameters {
			sb.WriteString(fmt.Sprintf("%s: %v\n", key, value))
		}
		sb.WriteString("---\n\n")
	}

	// Write title
	if c.Title != "" {
		sb.WriteString("# " + c.Title + "\n\n")
	}

	// Write messages
	for _, msg := range c.Messages {
		if msg.Role == "You" {
			sb.WriteString("## You\n\n")
		} else {
			sb.WriteString("### Response\n\n")
		}
		sb.WriteString(msg.Content + "\n\n")
	}

	return sb.String()
}
