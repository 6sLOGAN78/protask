package service

import (
	"fmt"

	"github.com/6sLOGAN78/go-protask/internal/lib/aws"
	"github.com/6sLOGAN78/go-protask/internal/lib/job"
	"github.com/6sLOGAN78/go-protask/internal/repository"
	"github.com/6sLOGAN78/go-protask/internal/server"
)

type Services struct {
	Auth     *AuthService
	Job      *job.JobService
	Todo     *TodoService
	Comment  *CommentService
	Category *CategoryService
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)

	awsClient, err := aws.NewAWS(s)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client: %w", err)
	}

	return &Services{
		Job:      s.Job,
		Auth:     authService,
		Category: NewCategoryService(s, repos.Category),
		Comment:  NewCommentService(s, repos.Comment, repos.Todo),
		Todo:     NewTodoService(s, repos.Todo, repos.Category, awsClient),
	}, nil
}