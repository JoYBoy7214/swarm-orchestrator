package storage

import (
	"context"

	"github.com/google/uuid"
)

type DbDriver interface {
	CreateDbs(ctx context.Context) error
	CreateWorkflow(ctx context.Context) (error, uuid.UUID)
	UpdateTask(ctx context.Context, task_id uuid.UUID, workflow_id uuid.UUID, status string) error
	Evaluate(ctx context.Context, workflow_id uuid.UUID) error
	GetworkflowReadyTasks(ctx context.Context, workflow_id uuid.UUID) (error, []uuid.UUID)
}
