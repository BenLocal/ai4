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

type Models struct {
	db db.Datebase
}

func NewModels(db db.Datebase) *Models {
	return &Models{db: db}
}

type AddModelsRequest struct {
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	ModelID   string `json:"model_id"`
	BaseURL   string `json:"base_url"`
	APIKey    string `json:"api_key,omitempty"`
	IsDefault bool   `json:"is_default"`
}

func (m *Models) AddRouters(router *router.Router) {
	router.POST("/models/add", m.addModelsHandler)
	router.GET("/chat", m.chatHandler)
}

func (m *Models) addModelsHandler(ctx *fasthttp.RequestCtx) {
	body := ctx.Request.Body()

	var req AddModelsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		utils.WriteError(ctx, fasthttp.StatusOK, err.Error())
		return
	}

	utils.WriteEmptySuccess(ctx)
}

func (m *Models) chatHandler(ctx *fasthttp.RequestCtx) {
	txt := ctx.QueryArgs().Peek("txt")
	if len(txt) == 0 {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString("Missing 'txt' parameter")
		return
	}
	options := []openai.Option{
		openai.WithModel("qwen-plus"),
		openai.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
	}

	chatHistory := sqlite3.NewSqliteChatMessageHistory(
		sqlite3.WithSession("example"),
		sqlite3.WithDB(m.db.GetNativeDb()),
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
			string(txt),
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
