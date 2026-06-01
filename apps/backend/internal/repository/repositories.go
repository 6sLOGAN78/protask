package repository

import "github.com/6sLOGAN78/go-protask/internal/server"

type Repositories struct {
	Todo     *TodoRepository
	Comment  *CommentRepository
}
func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{
		Todo:     NewTodoRepository(s),
		Comment:  NewCommentRepository(s),
	}
}