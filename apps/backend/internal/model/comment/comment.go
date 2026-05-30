package comment

import (
	"github.com/google/uuid"
	"github.com/6sLOGAN78/go-protask/internal/model"
)

type Comment struct {
	model.Base
	UserID  string    `json:"userId" db:"user_id"`
	TodoID  uuid.UUID `json:"todoId" db:"todo_id"`
	Content string    `json:"content" db:"content"`
}