package handler

import (
	"github.com/6sLOGAN78/go-protask/internal/server"
	"github.com/6sLOGAN78/go-protask/internal/service"
)

type Handlers struct {
	Health  *HealthHandler
	OpenAPI *OpenAPIHandler
	Todo   *TodoHandler
	Comment *CommentHandler
	Category *CategoryHandler
}

func NewHandlers(s *server.Server, services *service.Services) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(s),
		OpenAPI: NewOpenAPIHandler(s),
	}
}
