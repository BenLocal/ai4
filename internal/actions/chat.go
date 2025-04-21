package actions

import (
	"bufio"
	"context"
	"encoding/json"

	"github.com/benlocal/ai4/internal/db"
	"github.com/benlocal/ai4/internal/utils"
	"github.com/fasthttp/router"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/memory/sqlite3"
	"github.com/valyala/fasthttp"
)

type Chats struct {
	db db.Datebase
}

func NewChats(db db.Datebase) *Models {
	return &Models{db: db}
}

func (c *Chats) AddRouters(router *router.Router) {
	router.GET("/chat_txt", c.chatTxtHandler)
	router.GET("/chat", c.chatHandler)
}

type ChatRequest struct {
	Prompt string `json:"prompt"`
}

func (c *Chats) chatHandler(ctx *fasthttp.RequestCtx) {
	body := ctx.Request.Body()

	var req ChatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		utils.WriteError(ctx, fasthttp.StatusOK, err.Error())
		return
	}

	if len(req.Prompt) == 0 {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString("Missing 'prompt' parameter")
		return
	}

	ctx.Response.Header.SetContentType("text/event-stream; charset=utf-8")
	ctx.Response.Header.Set("Transfer-Encoding", "chunked")
	ctx.Response.Header.Set("Cache-Control", "no-cache")
	ctx.Response.Header.Set("Connection", "keep-alive")
	ctx.Response.Header.Set("X-Content-Type-Options", "nosniff")
	c.innerChat(ctx, string(req.Prompt))
}

func (c *Chats) chatTxtHandler(ctx *fasthttp.RequestCtx) {
	txt := ctx.QueryArgs().Peek("txt")
	if len(txt) == 0 {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString("Missing 'txt' parameter")
		return
	}
	ctx.Response.Header.SetContentType("text/plain; charset=utf-8")
	ctx.Response.Header.Set("Transfer-Encoding", "chunked")
	ctx.Response.Header.Set("Cache-Control", "no-cache")
	ctx.Response.Header.Set("Connection", "keep-alive")
	ctx.Response.Header.Set("X-Content-Type-Options", "nosniff")
	c.innerChat(ctx, string(txt))
}

func (c *Chats) innerChat(ctx *fasthttp.RequestCtx, txt string) {
	options := []openai.Option{
		openai.WithModel("qwen-plus"),
		openai.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
	}

	chatHistory := sqlite3.NewSqliteChatMessageHistory(
		sqlite3.WithSession("example"),
		sqlite3.WithDB(c.db.GetNativeDb()),
	)

	llm, err := openai.New(options...)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBody([]byte("Error creating LLM"))
		return
	}
	conversationBuffer := memory.NewConversationBuffer(memory.WithChatHistory(chatHistory))
	llmChain := chains.NewConversation(llm, conversationBuffer)

	ctx.Response.Header.SetContentType("text/plain; charset=utf-8")
	ctx.Response.Header.Set("Transfer-Encoding", "chunked")
	ctx.Response.Header.Set("Cache-Control", "no-cache")
	ctx.Response.Header.Set("Connection", "keep-alive")
	ctx.Response.Header.Set("X-Content-Type-Options", "nosniff")
	ctx.SetBodyStreamWriter(func(w *bufio.Writer) {
		bgCtx := context.Background()

		_, err = chains.Run(
			bgCtx,
			llmChain,
			txt,
			chains.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				if _, err := w.Write(chunk); err != nil {
					return err
				}
				return w.Flush()
			}),
		)

		if err != nil {
			w.WriteString("\n\nError: " + err.Error())
			w.Flush()
		}
	})
}
