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

func NewChats(db db.Datebase) *Chats {
	return &Chats{db: db}
}

func (c *Chats) AddRouters(router *router.Router) {
	router.GET("/chat_txt", c.chatTxtHandler)
	router.POST("/chat", c.chatHandler)
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
	c.innerChat(ctx, string(req.Prompt), func(done bool, chunks []byte, err error) string {
		if err != nil {
			return "Error: " + err.Error() + "\n\n"
		}
		if done {
			return "data: [DONE]\n\n"
		}

		return "data: " + string(chunks) + "\n\n"
	})
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
	c.innerChat(ctx, string(txt), func(done bool, chunks []byte, err error) string {
		if err != nil {
			return "\n\nError: " + err.Error()
		}
		if done {
			return "\n\n"
		}

		return string(chunks)
	})
}

func (c *Chats) innerChat(ctx *fasthttp.RequestCtx, txt string, writer func(done bool, chunks []byte, err error) string) {
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

	ctx.SetBodyStreamWriter(func(w *bufio.Writer) {
		bgCtx := context.Background()

		_, err = chains.Run(
			bgCtx,
			llmChain,
			txt,
			chains.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				v := writer(false, chunk, nil)
				if _, err := w.WriteString(v); err != nil {
					return err
				}
				return w.Flush()
			}),
		)

		if err != nil {
			v := writer(false, nil, err)
			w.WriteString(v)
			w.Flush()
		}

		r := writer(true, nil, nil)
		w.Flush()
		if _, err := w.WriteString(r); err != nil {
			return
		}
	})
}
