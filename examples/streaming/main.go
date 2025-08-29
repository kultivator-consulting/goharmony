package main

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/raoul/goharmony"
)

// simulateStream simulates receiving chunks of a response
func simulateStream() <-chan string {
	chunks := []string{
		"<|channel|>analysis",
		"<|message|>Analyzing the user's request",
		" for information<|end|>\n",
		"<|channel|>commentary<|message|>",
		"Preparing detailed response",
		"<|end|>\n<|channel|>",
		"final<|message|>Based on my analysis, ",
		"here's what I found: ",
		"The answer is 42.",
		"<|end|>",
	}

	ch := make(chan string)
	go func() {
		defer close(ch)
		for _, chunk := range chunks {
			ch <- chunk
			time.Sleep(100 * time.Millisecond) // Simulate network delay
		}
	}()
	return ch
}

func main() {
	fmt.Println("=== Streaming Response Handler ===\n")

	parser := goharmony.NewParser()
	stream := simulateStream()
	
	var buffer strings.Builder
	var lastFinalContent string
	
	fmt.Println("Receiving stream...")
	for chunk := range stream {
		buffer.WriteString(chunk)
		currentContent := buffer.String()
		
		// Try to parse what we have so far
		messages, err := parser.ParseResponse(currentContent)
		if err != nil {
			continue // Wait for more data
		}
		
		// Look for final channel content
		for _, msg := range messages {
			if msg.Channel == goharmony.ChannelFinal {
				// Check if we have new content
				if msg.Content != lastFinalContent && msg.Content != "" {
					// Display incremental content
					if lastFinalContent == "" {
						fmt.Print("User sees: ")
					}
					
					// Show only the new part
					if strings.HasPrefix(msg.Content, lastFinalContent) {
						newPart := msg.Content[len(lastFinalContent):]
						fmt.Print(newPart)
					} else {
						fmt.Print(msg.Content)
					}
					
					lastFinalContent = msg.Content
				}
			}
		}
	}
	fmt.Println("\n\nStream complete!")
	
	// Show full parsed structure
	fmt.Println("\n=== Full Parsed Structure ===")
	messages, _ := parser.ParseResponse(buffer.String())
	for _, msg := range messages {
		fmt.Printf("[%s] %s\n", msg.Channel, msg.Content)
	}
	
	// Demonstrate real-time filtering
	fmt.Println("\n=== Real-time Channel Filtering ===")
	demoRealTimeFiltering()
}

func demoRealTimeFiltering() {
	parser := goharmony.NewParser()
	
	// Simulated real-time input
	input := `<|channel|>analysis<|message|>Processing request<|end|>
<|channel|>commentary to=functions.search<|message|>{"query": "latest news"}<|call|>
<|channel|>final<|message|>I'm searching for the latest news for you.<|end|>`
	
	reader := bufio.NewReader(strings.NewReader(input))
	var accumulated strings.Builder
	
	fmt.Println("Processing character by character:")
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			break
		}
		
		accumulated.WriteRune(r)
		
		// Check if we have a complete message
		content := accumulated.String()
		if strings.Contains(content, "<|end|>") || strings.Contains(content, "<|call|>") {
			messages, _ := parser.ParseResponse(content)
			
			for _, msg := range messages {
				switch msg.Channel {
				case goharmony.ChannelAnalysis:
					fmt.Printf("  [Internal] %s\n", msg.Content)
				case goharmony.ChannelCommentary:
					if msg.IsCall {
						fmt.Printf("  [Tool Call] Calling %s with %s\n", msg.To, msg.Content)
					} else {
						fmt.Printf("  [Commentary] %s\n", msg.Content)
					}
				case goharmony.ChannelFinal:
					fmt.Printf("  [User Sees] %s\n", msg.Content)
				}
			}
			
			// Reset for next message
			accumulated.Reset()
		}
	}
}