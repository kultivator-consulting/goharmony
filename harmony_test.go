package goharmony

import (
	"reflect"
	"testing"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Fatal("NewParser() returned nil")
	}
	if parser.config.DefaultRole != "assistant" {
		t.Errorf("Expected default role 'assistant', got '%s'", parser.config.DefaultRole)
	}
}

func TestParseResponse_BasicChannels(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		name     string
		input    string
		expected []Message
	}{
		{
			name:  "Single final channel",
			input: `<|channel|>final<|message|>Hello world<|end|>`,
			expected: []Message{
				{Role: "assistant", Channel: ChannelFinal, Content: "Hello world"},
			},
		},
		{
			name:  "Single analysis channel",
			input: `<|channel|>analysis<|message|>Internal reasoning<|end|>`,
			expected: []Message{
				{Role: "assistant", Channel: ChannelAnalysis, Content: "Internal reasoning"},
			},
		},
		{
			name: "Multiple channels",
			input: `<|channel|>analysis<|message|>Thinking...<|end|>
<|channel|>final<|message|>Here's the answer<|end|>`,
			expected: []Message{
				{Role: "assistant", Channel: ChannelAnalysis, Content: "Thinking..."},
				{Role: "assistant", Channel: ChannelFinal, Content: "Here's the answer"},
			},
		},
		{
			name:  "With role specified",
			input: `<|start|>system<|channel|>final<|message|>System message<|end|>`,
			expected: []Message{
				{Role: "system", Channel: ChannelFinal, Content: "System message"},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := parser.ParseResponse(tt.input)
			if err != nil {
				t.Fatalf("ParseResponse() error = %v", err)
			}
			if !reflect.DeepEqual(messages, tt.expected) {
				t.Errorf("ParseResponse() = %v, want %v", messages, tt.expected)
			}
		})
	}
}

func TestParseResponse_FunctionCalls(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		name     string
		input    string
		expected Message
	}{
		{
			name:  "Harmony format function call",
			input: `<|channel|>commentary to=functions.get_weather<|message|>{"location": "NYC"}<|call|>`,
			expected: Message{
				Role:    "assistant",
				Channel: ChannelCommentary,
				Content: `{"location": "NYC"}`,
				To:      "functions.get_weather",
				IsCall:  true,
			},
		},
		{
			name:  "FUNCTION_CALL format",
			input: `FUNCTION_CALL: get_weather({"location": "NYC"})`,
			expected: Message{
				Role:    "assistant",
				Channel: ChannelCommentary,
				Content: `{"location": "NYC"}`,
				To:      "functions.get_weather",
				IsCall:  true,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := parser.ParseResponse(tt.input)
			if err != nil {
				t.Fatalf("ParseResponse() error = %v", err)
			}
			if len(messages) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(messages))
			}
			if !reflect.DeepEqual(messages[0], tt.expected) {
				t.Errorf("ParseResponse() = %v, want %v", messages[0], tt.expected)
			}
		})
	}
}

func TestExtractFinalMessage(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Extract from mixed channels",
			input:    `<|channel|>analysis<|message|>Thinking...<|end|><|channel|>final<|message|>Hello!<|end|>`,
			expected: "Hello!",
		},
		{
			name:     "No final channel",
			input:    `<|channel|>analysis<|message|>Just analysis<|end|>`,
			expected: "",
		},
		{
			name:     "Plain text (non-strict mode)",
			input:    "Plain text message",
			expected: "Plain text message",
		},
		{
			name:     "Skip function calls",
			input:    `<|channel|>final<|message|>FUNCTION_CALL: test()<|end|>`,
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ExtractFinalMessage(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractFinalMessage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractFunctionCall(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		name         string
		input        string
		expectedName string
		expectedArgs string
		expectedOK   bool
	}{
		{
			name:         "Harmony format",
			input:        `<|channel|>commentary to=functions.calculate<|message|>{"x": 5}<|call|>`,
			expectedName: "calculate",
			expectedArgs: `{"x": 5}`,
			expectedOK:   true,
		},
		{
			name:         "FUNCTION_CALL format",
			input:        `FUNCTION_CALL: calculate({"x": 5})`,
			expectedName: "calculate",
			expectedArgs: `{"x": 5}`,
			expectedOK:   true,
		},
		{
			name:         "No function call",
			input:        `<|channel|>final<|message|>Regular message<|end|>`,
			expectedName: "",
			expectedArgs: "",
			expectedOK:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, args, ok := parser.ExtractFunctionCall(tt.input)
			if name != tt.expectedName || args != tt.expectedArgs || ok != tt.expectedOK {
				t.Errorf("ExtractFunctionCall() = (%v, %v, %v), want (%v, %v, %v)",
					name, args, ok, tt.expectedName, tt.expectedArgs, tt.expectedOK)
			}
		})
	}
}

func TestGetChannelContent(t *testing.T) {
	parser := NewParser()
	
	input := `<|channel|>analysis<|message|>First analysis<|end|>
<|channel|>final<|message|>User message<|end|>
<|channel|>analysis<|message|>Second analysis<|end|>`
	
	tests := []struct {
		name     string
		channel  Channel
		expected []string
	}{
		{
			name:     "Get analysis content",
			channel:  ChannelAnalysis,
			expected: []string{"First analysis", "Second analysis"},
		},
		{
			name:     "Get final content",
			channel:  ChannelFinal,
			expected: []string{"User message"},
		},
		{
			name:     "Get commentary content",
			channel:  ChannelCommentary,
			expected: nil,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.GetChannelContent(input, tt.channel)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetChannelContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHasChannel(t *testing.T) {
	parser := NewParser()
	
	input := `<|channel|>analysis<|message|>Test<|end|>
<|channel|>final<|message|>Test<|end|>`
	
	if !parser.HasChannel(input, ChannelAnalysis) {
		t.Error("HasChannel() should return true for analysis channel")
	}
	if !parser.HasChannel(input, ChannelFinal) {
		t.Error("HasChannel() should return true for final channel")
	}
	if parser.HasChannel(input, ChannelCommentary) {
		t.Error("HasChannel() should return false for commentary channel")
	}
}

func TestExtractJSON(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "Valid JSON",
			input:       `Some text {"key": "value", "number": 42} more text`,
			expectError: false,
		},
		{
			name:        "No JSON",
			input:       `Just plain text`,
			expectError: true,
		},
		{
			name:        "Invalid JSON",
			input:       `{invalid json}`,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ExtractJSON(tt.input)
			if tt.expectError {
				if err == nil {
					t.Error("ExtractJSON() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ExtractJSON() unexpected error: %v", err)
				}
				if result == nil {
					t.Error("ExtractJSON() returned nil result")
				}
			}
		})
	}
}

func TestStrictMode(t *testing.T) {
	config := ParserConfig{
		StrictMode:  true,
		DefaultRole: "assistant",
	}
	parser := NewParserWithConfig(config)
	
	// Invalid channel should fail in strict mode
	input := `<|channel|>invalid<|message|>Test<|end|>`
	_, err := parser.ParseResponse(input)
	if err == nil {
		t.Error("Expected error for invalid channel in strict mode")
	}
	
	// Plain text should not be parsed in strict mode
	plainText := "Plain text"
	messages, err := parser.ParseResponse(plainText)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(messages) != 0 {
		t.Error("Plain text should not produce messages in strict mode")
	}
}

func TestMessageString(t *testing.T) {
	tests := []struct {
		name     string
		msg      Message
		expected string
	}{
		{
			name: "Regular message",
			msg: Message{
				Role:    "assistant",
				Channel: ChannelFinal,
				Content: "Hello",
			},
			expected: "[assistant/final] Hello",
		},
		{
			name: "Function call",
			msg: Message{
				Role:    "assistant",
				Channel: ChannelCommentary,
				Content: `{"x": 5}`,
				To:      "functions.calculate",
				IsCall:  true,
			},
			expected: "[assistant/commentary] Function call to functions.calculate: {\"x\": 5}",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkParseResponse(b *testing.B) {
	parser := NewParser()
	input := `<|channel|>analysis<|message|>Thinking about the request<|end|>
<|channel|>commentary to=functions.test<|message|>{"data": "test"}<|call|>
<|channel|>final<|message|>Here is the final response with some longer text content<|end|>`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseResponse(input)
	}
}

func BenchmarkExtractFinalMessage(b *testing.B) {
	parser := NewParser()
	input := `<|channel|>analysis<|message|>Internal processing<|end|>
<|channel|>final<|message|>The actual response to the user<|end|>`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.ExtractFinalMessage(input)
	}
}