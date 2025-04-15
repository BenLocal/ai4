package runtime

import (
	"bufio"
	"context"

	"github.com/benlocal/ai4/pkg/gracefulshutdown"
	"github.com/benlocal/ai4/pkg/service"
	"github.com/fasthttp/router"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/valyala/fasthttp"
)

func Start() error {
	g := gracefulshutdown.New()
	g.CatchSignals()

	router := router.New()
	router.GET("/", func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("Hello, World!")
	})
	router.GET("/chat", func(ctx *fasthttp.RequestCtx) {
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
		llm, err := openai.New(options...)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetBody([]byte("Error creating LLM"))
			return
		}

		ctx.Response.Header.SetContentType("text/plain; charset=utf-8")
		ctx.Response.Header.Set("Transfer-Encoding", "chunked")
		ctx.Response.Header.Set("Cache-Control", "no-cache")
		ctx.Response.Header.Set("Connection", "keep-alive")
		ctx.Response.Header.Set("X-Content-Type-Options", "nosniff")

		responseCh := make(chan []byte, 10)
		errorCh := make(chan error, 1)
		ctx.SetBodyStreamWriter(func(w *bufio.Writer) {
			for chunk := range responseCh {
				if _, err := w.Write(chunk); err != nil {
					errorCh <- err
					return
				}
				if err := w.Flush(); err != nil {
					errorCh <- err
					return
				}
			}
		})
		go func() {
			defer close(responseCh)
			bgCtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			_, err = llms.GenerateFromSinglePrompt(bgCtx, llm, string(txt), llms.WithStreamingFunc(func(bg context.Context, chunk []byte) error {
				select {
				case responseCh <- chunk:
					return nil
				case err := <-errorCh:
					return err
				case <-bg.Done():
					return bg.Err()
				}
			}))
			if err != nil {
				responseCh <- []byte("\n\nError: " + err.Error())
			}
		}()
	})

	g.Add(service.New(7080, router))

	ctx := context.Background()
	return g.Start(ctx)
}
