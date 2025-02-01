package conversation

import (
	"strings"
	"testing"
	"time"
)

func TestParseContent(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantTitle   string
		wantMsgLen  int
		wantErr     bool
		wantLastMsg string
	}{
		{
			name: "basic conversation",
			input: `# Simple Chat
## You
Hello
### Response
Hi there
## You
How are you?`,
			wantTitle:   "Simple Chat",
			wantMsgLen:  3,
			wantErr:     false,
			wantLastMsg: "How are you?",
		},
		{
			name: "conversation with empty lines",
			input: `# Chat with Spaces

## You

First message

### Response

First response

## You

Second message`,
			wantTitle:   "Chat with Spaces",
			wantMsgLen:  3,
			wantErr:     false,
			wantLastMsg: "Second message",
		},
		{
			name: "conversation without title",
			input: `## You
Hello
### Response
Hi there`,
			wantTitle:   "",
			wantMsgLen:  2,
			wantErr:     false,
			wantLastMsg: "Hello",
		},
		{
			name:      "empty content",
			input:     "",
			wantTitle: "",
			wantErr:   true,
		},
		{
			name: "multi-line messages",
			input: `# Multi-line Chat
## You
This is a
multi-line
message
### Response
This is also a
multi-line response`,
			wantTitle:   "Multi-line Chat",
			wantMsgLen:  2,
			wantErr:     false,
			wantLastMsg: "This is a\nmulti-line\nmessage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv, err := ParseContent(tt.input)

			// Check error condition
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// Check title
			if conv.Title != tt.wantTitle {
				t.Errorf("ParseContent() title = %v, want %v", conv.Title, tt.wantTitle)
			}

			// Check number of messages
			if len(conv.Messages) != tt.wantMsgLen {
				t.Errorf("ParseContent() message count = %v, want %v", len(conv.Messages), tt.wantMsgLen)
			}

			// Check last user message if specified
			if tt.wantLastMsg != "" {
				lastMsg, err := conv.GetLastUserMessage()
				if err != nil {
					t.Errorf("GetLastUserMessage() unexpected error: %v", err)
				}
				if lastMsg != tt.wantLastMsg {
					t.Errorf("GetLastUserMessage() = %v, want %v", lastMsg, tt.wantLastMsg)
				}
			}
		})
	}
}

func TestGetLastUserMessage(t *testing.T) {
	tests := []struct {
		name    string
		conv    *Conversation
		want    string
		wantErr bool
	}{
		{
			name: "has user message",
			conv: &Conversation{
				Messages: []Message{
					{Role: "You", Content: "First"},
					{Role: "Response", Content: "Reply"},
					{Role: "You", Content: "Second"},
				},
			},
			want:    "Second",
			wantErr: false,
		},
		{
			name: "no user message",
			conv: &Conversation{
				Messages: []Message{
					{Role: "Response", Content: "Reply"},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "empty messages",
			conv: &Conversation{
				Messages: []Message{},
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.conv.GetLastUserMessage()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLastUserMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetLastUserMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddResponse(t *testing.T) {
	conv := &Conversation{
		Title:    "Test Chat",
		Messages: []Message{},
	}

	// Add a response
	responseText := "Test response"
	conv.AddResponse(responseText)

	// Check if response was added correctly
	if len(conv.Messages) != 1 {
		t.Errorf("AddResponse() message count = %v, want 1", len(conv.Messages))
	}

	msg := conv.Messages[0]
	if msg.Role != "Response" {
		t.Errorf("AddResponse() role = %v, want Response", msg.Role)
	}
	if msg.Content != responseText {
		t.Errorf("AddResponse() content = %v, want %v", msg.Content, responseText)
	}
	if msg.Timestamp.IsZero() {
		t.Error("AddResponse() timestamp is zero")
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name string
		conv *Conversation
		want string
	}{
		{
			name: "full conversation",
			conv: &Conversation{
				Title: "Test Chat",
				Messages: []Message{
					{Role: "You", Content: "Hello", Timestamp: time.Now()},
					{Role: "Response", Content: "Hi there", Timestamp: time.Now()},
				},
			},
			want: `# Test Chat

## You

Hello

### Response

Hi there

`,
		},
		{
			name: "no title",
			conv: &Conversation{
				Messages: []Message{
					{Role: "You", Content: "Hello", Timestamp: time.Now()},
				},
			},
			want: `## You

Hello

`,
		},
		{
			name: "multi-line content",
			conv: &Conversation{
				Title: "Multi-line Test",
				Messages: []Message{
					{Role: "You", Content: "Line 1\nLine 2", Timestamp: time.Now()},
				},
			},
			want: `# Multi-line Test

## You

Line 1
Line 2

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.conv.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContentRoundTrip(t *testing.T) {
	original := `# Test Chat

## You

Hello there

### Response

Hi! How can I help?

## You

What's the weather?

### Response

I cannot check the weather.`

	// Parse the original content
	conv, err := ParseContent(original)
	if err != nil {
		t.Fatalf("ParseContent() error = %v", err)
	}

	// Convert back to string
	result := conv.String()

	// Compare the normalized versions (removing extra whitespace)
	normalizeSpace := func(s string) string {
		lines := strings.Split(s, "\n")
		var normalized []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				normalized = append(normalized, trimmed)
			}
		}
		return strings.Join(normalized, "\n")
	}

	if normalizeSpace(result) != normalizeSpace(original) {
		t.Errorf("Content round-trip failed.\nOriginal:\n%s\n\nResult:\n%s", original, result)
	}
}
