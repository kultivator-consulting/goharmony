# GoHarmony

A Go implementation of the OpenAI Harmony format parser for structured LLM responses.

[![Go Reference](https://pkg.go.dev/badge/github.com/kultivator-consulting/goharmony.svg)](https://pkg.go.dev/github.com/kultivator-consulting/goharmony)
[![Go Report Card](https://goreportcard.com/badge/github.com/kultivator-consulting/goharmony)](https://goreportcard.com/report/github.com/kultivator-consulting/goharmony)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Overview

GoHarmony provides a robust parser for the OpenAI Harmony response format, which is used by models like `gpt-oss` to structure their outputs with multiple channels (analysis, commentary, final). This format enables clear separation between internal reasoning, tool calls, and user-facing responses.

## Features

- 🎯 **Full Harmony Format Support** - Parses all Harmony format tokens and structures
- 📊 **Multi-Channel Processing** - Separate handling for analysis, commentary, and final channels
- 🔧 **Function Call Extraction** - Automatically detects and parses tool/function calls
- 🚀 **High Performance** - Efficient regex-based parsing with minimal allocations
- 📝 **Comprehensive Documentation** - Full API documentation with examples
- ✅ **Well Tested** - Extensive test suite with edge cases

## Installation

```bash
go get github.com/kultivator-consulting/goharmony
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/kultivator-consulting/goharmony"
)

func main() {
    parser := goharmony.NewParser()
    
    response := `<|channel|>analysis<|message|>Analyzing user request<|end|>
<|channel|>final<|message|>Hello! How can I help you today?<|end|>`
    
    // Extract only the user-facing message
    finalMessage := parser.ExtractFinalMessage(response)
    fmt.Println(finalMessage)
    // Output: Hello! How can I help you today?
    
    // Parse all messages
    messages, _ := parser.ParseResponse(response)
    for _, msg := range messages {
        fmt.Printf("Channel: %s, Content: %s\n", msg.Channel, msg.Content)
    }
}
```

## Format Specification

The Harmony format uses special tokens to structure responses:

| Token | Description |
|-------|-------------|
| `<\|start\|>` | Begins a new message |
| `<\|end\|>` | Ends a message |
| `<\|channel\|>` | Specifies the message channel |
| `<\|message\|>` | Marks the beginning of message content |
| `<\|call\|>` | Indicates a function/tool call |
| `<\|return\|>` | Signals response completion |

### Channels

- **`analysis`** - Internal chain-of-thought reasoning (hidden from users)
- **`commentary`** - Tool calls and intermediate explanations
- **`final`** - User-facing responses

### Example Response

```
<|channel|>analysis<|message|>User wants weather information<|end|>
<|start|>assistant<|channel|>commentary to=functions.get_weather<|message|>{"location":"NYC"}<|call|>
<|start|>assistant<|channel|>final<|message|>The weather in NYC is sunny and 72°F.<|end|>
```

## API Documentation

### Parser

```go
type Parser struct {
    // Parser configuration
}

func NewParser() *Parser
func (p *Parser) ParseResponse(content string) ([]Message, error)
func (p *Parser) ExtractFinalMessage(content string) string
func (p *Parser) ExtractFunctionCall(content string) (name string, args string, found bool)
func (p *Parser) GetChannelContent(content string, channel Channel) []string
```

### Message

```go
type Message struct {
    Role    string  // Message role (system, user, assistant, etc.)
    Channel Channel // Message channel (analysis, commentary, final)
    Content string  // Message content
    To      string  // Target for function calls (e.g., "functions.get_weather")
    IsCall  bool    // Whether this is a function call
}
```

## Advanced Usage

### Extracting Function Calls

```go
response := `<|channel|>commentary to=functions.calculate<|message|>{"x": 5, "y": 3}<|call|>`

name, args, found := parser.ExtractFunctionCall(response)
if found {
    fmt.Printf("Function: %s, Args: %s\n", name, args)
    // Output: Function: calculate, Args: {"x": 5, "y": 3}
}
```

### Filtering Channels

```go
// Get only analysis messages
analysisContent := parser.GetChannelContent(response, goharmony.ChannelAnalysis)

// Get only final messages
finalContent := parser.GetChannelContent(response, goharmony.ChannelFinal)
```

### Stream Processing

```go
// Process streaming responses
func processStream(chunk string, parser *goharmony.Parser) {
    messages, _ := parser.ParseResponse(chunk)
    for _, msg := range messages {
        if msg.Channel == goharmony.ChannelFinal {
            // Display to user
            fmt.Print(msg.Content)
        }
    }
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- OpenAI for the Harmony format specification
- The Rust and Python implementations that inspired this Go version

## Links

- [OpenAI Harmony Format Documentation](https://cookbook.openai.com/articles/openai-harmony)
- [Official Harmony Repository](https://github.com/openai/harmony)
- [Report Issues](https://github.com/kultivator-consulting/goharmony/issues)