package actions

import (
	"encoding/json"

	"github.com/benlocal/ai4/internal/db"
	"github.com/benlocal/ai4/internal/utils"
	"github.com/fasthttp/router"
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
