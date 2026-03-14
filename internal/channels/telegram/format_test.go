package telegram

import (
	"html"
	"strings"
	"testing"

	"github.com/nextlevelbuilder/goclaw/internal/bus"
)

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{"Khởi động", 9},        // Vietnamese diacritics = single-width
		{"Hardware tối thiểu", 18}, // Vietnamese diacritics = single-width
		{"Ngôn ngữ", 8},
		{"đ", 1},                 // Vietnamese d-stroke = single-width
		{"中文", 4},               // CJK = double-width
		{"日本語", 6},              // CJK = double-width
	}

	for _, tt := range tests {
		got := displayWidth(tt.input)
		if got != tt.want {
			t.Errorf("displayWidth(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestRenderTableAsCode_Vietnamese(t *testing.T) {
	lines := []string{
		"| Metric | OpenClaw | ZeroClaw |",
		"|--------|----------|----------|",
		"| Ngôn ngữ | TypeScript/Node.js | Rust |",
		"| Khởi động | > 500s | < 10ms |",
		"| Hardware tối thiểu | Mac mini $599 | $10 (bao gồm cả Raspberry Pi) |",
	}

	result := renderTableAsCode(lines)

	// Every non-separator line should have the same number of pipes
	resultLines := strings.Split(result, "\n")
	if len(resultLines) < 3 {
		t.Fatalf("expected at least 3 lines, got %d", len(resultLines))
	}

	// Check separator line width matches header line width
	headerWidth := displayWidth(resultLines[0])
	sepWidth := displayWidth(resultLines[1])
	if headerWidth != sepWidth {
		t.Errorf("header width (%d) != separator width (%d)\nheader: %s\nsep:    %s",
			headerWidth, sepWidth, resultLines[0], resultLines[1])
	}

	// Check all data rows match header width
	for i := 2; i < len(resultLines); i++ {
		rowWidth := displayWidth(resultLines[i])
		if rowWidth != headerWidth {
			t.Errorf("row %d width (%d) != header width (%d)\nrow:    %s\nheader: %s",
				i, rowWidth, headerWidth, resultLines[i], resultLines[0])
		}
	}
}

func TestAgentNamePrefix(t *testing.T) {
	tests := []struct {
		name       string
		agentName  string
		content    string
		wantPrefix string
	}{
		{
			name:       "with agent name",
			agentName:  "MyAgent",
			content:    "Hello world",
			wantPrefix: "<b>MyAgent</b>\n\nHello world",
		},
		{
			name:       "with special characters in agent name",
			agentName:  "Agent & Co <test>",
			content:    "Hello",
			wantPrefix: "<b>Agent &amp; Co &lt;test&gt;</b>\n\nHello",
		},
		{
			name:       "empty agent name",
			agentName:  "",
			content:    "Hello world",
			wantPrefix: "Hello world",
		},
		{
			name:       "unicode agent name",
			agentName:  "智能助手",
			content:    "你好",
			wantPrefix: "<b>智能助手</b>\n\n你好",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := bus.OutboundMessage{
				Content:   tt.content,
				AgentName: tt.agentName,
			}

			// Simulate the prefix logic from Send()
			content := msg.Content
			if msg.AgentName != "" {
				content = "<b>" + html.EscapeString(msg.AgentName) + "</b>\n\n" + content
			}

			if content != tt.wantPrefix {
				t.Errorf("got %q, want %q", content, tt.wantPrefix)
			}
		})
	}
}
