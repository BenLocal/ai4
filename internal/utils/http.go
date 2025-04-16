package utils

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func WriteJSON(ctx *fasthttp.RequestCtx, statusCode int, response JSONResponse) {
	ctx.Response.SetStatusCode(statusCode)
	ctx.Response.Header.SetContentType("application/json")

	respJSON, err := json.Marshal(response)
	if err != nil {
		// Fallback if JSON marshaling fails
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Response.SetBodyString("Failed to generate JSON response")
		return
	}

	ctx.Response.SetBody(respJSON)
}

func WriteEmptySuccess(ctx *fasthttp.RequestCtx) {
	statusCode := fasthttp.StatusOK
	WriteJSON(ctx, statusCode, JSONResponse{
		Success: true,
		Message: "",
		Data:    nil,
	})
}

func WriteSuccessWithoutMessage(ctx *fasthttp.RequestCtx, data interface{}) {
	statusCode := fasthttp.StatusOK
	WriteJSON(ctx, statusCode, JSONResponse{
		Success: true,
		Message: "",
		Data:    data,
	})
}

func WriteSuccess(ctx *fasthttp.RequestCtx, message string, data interface{}) {
	statusCode := fasthttp.StatusOK
	WriteJSON(ctx, statusCode, JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func WriteError(ctx *fasthttp.RequestCtx, statusCode int, errMsg string) {
	if statusCode == 0 {
		statusCode = fasthttp.StatusInternalServerError
	}

	WriteJSON(ctx, statusCode, JSONResponse{
		Success: false,
		Error:   errMsg,
	})
}
