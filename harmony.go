// Package goharmony provides a Go implementation of the OpenAI Harmony format parser.
// The Harmony format is used by models like gpt-oss to structure their outputs with
// multiple channels (analysis, commentary, final) for better separation of concerns.
package goharmony

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Channel represents different message channels in the Harmony format
type Channel string

const (
	// ChannelAnalysis is for internal chain-of-thought reasoning
	ChannelAnalysis Channel = "analysis"
	// ChannelCommentary is for tool calls and intermediate explanations
	ChannelCommentary Channel = "commentary"
	// ChannelFinal is for user-facing responses
	ChannelFinal Channel = "final"
)

// Message represents a parsed message from Harmony format
type Message struct {
	// Role of the message sender (e.g., "assistant", "system", "user")
	Role string `json:"role"`
	// Channel type of the message
	Channel Channel `json:"channel"`
	// Content of the message
	Content string `json:"content"`
	// To field for function calls (e.g., "functions.get_weather")
	To string `json:"to,omitempty"`
	// IsCall indicates whether this is a function/tool call
	IsCall bool `json:"is_call,omitempty"`
}

// Parser handles parsing of OpenAI Harmony format responses
type Parser struct {
	// Regex patterns for parsing
	messagePattern  *regexp.Regexp
	channelPattern  *regexp.Regexp
	functionPattern *regexp.Regexp
	// Configuration options
	config ParserConfig
}

// ParserConfig contains configuration options for the parser
type ParserConfig struct {
	// StrictMode enforces strict Harmony format compliance
	StrictMode bool
	// DefaultRole is the default role when not specified
	DefaultRole string
}

// DefaultConfig returns the default parser configuration
func DefaultConfig() ParserConfig {
	return ParserConfig{
		StrictMode:  false,
		DefaultRole: "assistant",
	}
}

// NewParser creates a new Harmony format parser with default configuration
func NewParser() *Parser {
	return NewParserWithConfig(DefaultConfig())
}

// NewParserWithConfig creates a new Harmony format parser with custom configuration
func NewParserWithConfig(config ParserConfig) *Parser {
	return &Parser{
		// Match messages with optional start tag and optional end tag
		messagePattern: regexp.MustCompile(
			`(?s)(?:<\|start\|>)?(\w+)?<\|channel\|>(\w+)(?:\s+to=([\w.]+))?` +
				`(?:\s*<\|constrain\|>\w+)?<\|message\|>(.*?)(?:<\|(?:end|call|return)\|>|$)`,
		),
		// Match standalone channel markers
		channelPattern: regexp.MustCompile(
			`<\|channel\|>(\w+)<\|message\|>(.*?)(?:<\|end\|>|$)`,
		),
		// Match function calls in various formats
		functionPattern: regexp.MustCompile(
			`FUNCTION_CALL:\s*(\w+)\((.*?)\)`,
		),
		config: config,
	}
}

// ParseResponse parses a Harmony formatted response into structured messages
func (p *Parser) ParseResponse(content string) ([]Message, error) {
	if content == "" {
		return nil, nil
	}

	var messages []Message

	// First, try to parse full Harmony format messages
	matches := p.messagePattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		msg := Message{}
		
		// match[1] = role (if present)
		// match[2] = channel
		// match[3] = to (if present)
		// match[4] = content
		
		if match[1] != "" {
			msg.Role = match[1]
		} else {
			msg.Role = p.config.DefaultRole
		}
		
		msg.Channel = Channel(match[2])
		msg.To = match[3]
		msg.Content = strings.TrimSpace(match[4])
		
		// Check if this is a function call
		if strings.Contains(match[0], "<|call|>") {
			msg.IsCall = true
		}
		
		// Validate channel in strict mode
		if p.config.StrictMode && !p.isValidChannel(msg.Channel) {
			return nil, fmt.Errorf("invalid channel: %s", msg.Channel)
		}
		
		messages = append(messages, msg)
	}

	// If no full format found, try simplified channel format
	if len(messages) == 0 && strings.Contains(content, "<|channel|>") {
		matches = p.channelPattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			msg := Message{
				Role:    p.config.DefaultRole,
				Channel: Channel(match[1]),
				Content: strings.TrimSpace(match[2]),
			}
			
			if p.config.StrictMode && !p.isValidChannel(msg.Channel) {
				return nil, fmt.Errorf("invalid channel: %s", msg.Channel)
			}
			
			messages = append(messages, msg)
		}
	}

	// If still no messages found, check for FUNCTION_CALL format
	if len(messages) == 0 && strings.Contains(content, "FUNCTION_CALL:") {
		if match := p.functionPattern.FindStringSubmatch(content); match != nil {
			msg := Message{
				Role:    p.config.DefaultRole,
				Channel: ChannelCommentary,
				Content: match[2],
				To:      fmt.Sprintf("functions.%s", match[1]),
				IsCall:  true,
			}
			messages = append(messages, msg)
		}
	}

	// If no structured format found and not in strict mode, treat as plain final message
	if len(messages) == 0 && !p.config.StrictMode && content != "" {
		messages = append(messages, Message{
			Role:    p.config.DefaultRole,
			Channel: ChannelFinal,
			Content: content,
		})
	}

	return messages, nil
}

// ExtractFinalMessage extracts only the user-facing final message from a Harmony response
func (p *Parser) ExtractFinalMessage(content string) string {
	messages, err := p.ParseResponse(content)
	if err != nil || len(messages) == 0 {
		// In non-strict mode, return original content as fallback
		if !p.config.StrictMode {
			return content
		}
		return ""
	}

	// Look for final channel messages
	for _, msg := range messages {
		if msg.Channel == ChannelFinal && !msg.IsCall {
			// Skip function call syntax
			if !strings.HasPrefix(msg.Content, "FUNCTION_CALL:") {
				return msg.Content
			}
		}
	}

	// If no final channel found, return empty (don't expose analysis)
	return ""
}

// ExtractFunctionCall extracts function call information from a Harmony response
func (p *Parser) ExtractFunctionCall(content string) (functionName string, args string, found bool) {
	messages, err := p.ParseResponse(content)
	if err != nil {
		return "", "", false
	}

	// Look for function calls in commentary channel
	for _, msg := range messages {
		if msg.IsCall && msg.To != "" {
			// Extract function name from "functions.name" format
			parts := strings.Split(msg.To, ".")
			if len(parts) >= 2 {
				return parts[1], msg.Content, true
			}
		}
	}

	// Also check for FUNCTION_CALL format
	if match := p.functionPattern.FindStringSubmatch(content); match != nil {
		return match[1], match[2], true
	}

	return "", "", false
}

// GetChannelContent extracts content from a specific channel
func (p *Parser) GetChannelContent(content string, channel Channel) []string {
	messages, err := p.ParseResponse(content)
	if err != nil {
		return nil
	}

	var results []string
	for _, msg := range messages {
		if msg.Channel == channel {
			results = append(results, msg.Content)
		}
	}
	return results
}

// GetAllMessages returns all parsed messages with their channels
func (p *Parser) GetAllMessages(content string) ([]Message, error) {
	return p.ParseResponse(content)
}

// HasChannel checks if a response contains a specific channel
func (p *Parser) HasChannel(content string, channel Channel) bool {
	messages, err := p.ParseResponse(content)
	if err != nil {
		return false
	}

	for _, msg := range messages {
		if msg.Channel == channel {
			return true
		}
	}
	return false
}

// ExtractJSON attempts to extract and parse JSON from message content
func (p *Parser) ExtractJSON(content string) (map[string]interface{}, error) {
	// Try to find JSON in the content
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")
	
	if jsonStart == -1 || jsonEnd == -1 || jsonEnd < jsonStart {
		return nil, fmt.Errorf("no JSON found in content")
	}
	
	jsonStr := content[jsonStart : jsonEnd+1]
	
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	return result, nil
}

// isValidChannel checks if a channel is valid
func (p *Parser) isValidChannel(channel Channel) bool {
	switch channel {
	case ChannelAnalysis, ChannelCommentary, ChannelFinal:
		return true
	default:
		return false
	}
}

// String returns a string representation of a Message
func (m Message) String() string {
	if m.IsCall {
		return fmt.Sprintf("[%s/%s] Function call to %s: %s", m.Role, m.Channel, m.To, m.Content)
	}
	return fmt.Sprintf("[%s/%s] %s", m.Role, m.Channel, m.Content)
}