package service

import (
	"github.com/6sLOGAN78/go-protask/internal/lib/job"
	"github.com/6sLOGAN78/go-protask/internal/repository"
	"github.com/6sLOGAN78/go-protask/internal/server"
)

type Services struct {
	Auth *AuthService
	Job  *job.JobService
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)

	return &Services{
		Job:  s.Job,
		Auth: authService,
	}, nil
}
