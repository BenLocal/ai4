package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	ctx := context.Background()
	options := []openai.Option{
		openai.WithModel("deepseek/deepseek-v3-base:free"),
		openai.WithBaseURL("https://openrouter.ai/api/v1"),
	}
	llm, err := openai.New(options...)
	if err != nil {
		log.Fatal(err)
	}
	prompt := "帮我写一个100字关于春天的故事，给小朋友看的"
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(completion)
}
