package main

import (
	"fmt"
	"log"

	"github.com/raoul/goharmony"
)

func main() {
	// Create a new parser
	parser := goharmony.NewParser()

	// Example 1: Parse a simple response
	fmt.Println("=== Example 1: Simple Response ===")
	simpleResponse := `<|channel|>final<|message|>Hello! How can I help you today?<|end|>`
	
	finalMessage := parser.ExtractFinalMessage(simpleResponse)
	fmt.Printf("Final message: %s\n\n", finalMessage)

	// Example 2: Multi-channel response
	fmt.Println("=== Example 2: Multi-Channel Response ===")
	multiChannelResponse := `<|channel|>analysis<|message|>User is greeting me, I should respond politely<|end|>
<|channel|>final<|message|>Hello there! It's great to meet you. How can I assist you today?<|end|>`

	messages, err := parser.ParseResponse(multiChannelResponse)
	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range messages {
		fmt.Printf("Channel: %s, Content: %s\n", msg.Channel, msg.Content)
	}
	fmt.Println()

	// Only show the final message to the user
	userMessage := parser.ExtractFinalMessage(multiChannelResponse)
	fmt.Printf("Message for user: %s\n\n", userMessage)

	// Example 3: Function call
	fmt.Println("=== Example 3: Function Call ===")
	functionCallResponse := `<|channel|>analysis<|message|>User wants weather information<|end|>
<|start|>assistant<|channel|>commentary to=functions.get_weather<|message|>{"location": "New York", "units": "fahrenheit"}<|call|>
<|channel|>final<|message|>I'll check the weather in New York for you.<|end|>`

	// Extract function call
	funcName, args, found := parser.ExtractFunctionCall(functionCallResponse)
	if found {
		fmt.Printf("Function to call: %s\n", funcName)
		fmt.Printf("Arguments: %s\n", args)
	}

	// Get the user message
	userMsg := parser.ExtractFinalMessage(functionCallResponse)
	fmt.Printf("Tell user: %s\n\n", userMsg)

	// Example 4: Working with specific channels
	fmt.Println("=== Example 4: Channel Filtering ===")
	complexResponse := `<|channel|>analysis<|message|>First analysis step<|end|>
<|channel|>analysis<|message|>Second analysis step<|end|>
<|channel|>commentary<|message|>Preparing response<|end|>
<|channel|>final<|message|>Here's your answer!<|end|>`

	// Get only analysis messages
	analysisMessages := parser.GetChannelContent(complexResponse, goharmony.ChannelAnalysis)
	fmt.Println("Analysis steps:")
	for i, msg := range analysisMessages {
		fmt.Printf("  %d. %s\n", i+1, msg)
	}

	// Check if response has final channel
	if parser.HasChannel(complexResponse, goharmony.ChannelFinal) {
		fmt.Println("Response has a final message for the user")
	}
}