package storage

import (
	"time"

	"github.com/google/uuid"
)

type UserInput struct {
}
type UserOutput struct {
}
type Workflow_schema struct {
	Workflow_id uuid.UUID
	created_at  time.Time
	status      string
	Input_data  UserInput
	Output_data UserOutput
}
type Tasks_schema struct {
	Workflow_id uuid.UUID
	Task_id     uuid.UUID
	Task_type   string
	Status      string
	Updated_at  time.Time
	Created_at  time.Time
	Input_data  UserInput
	Output_data UserOutput
}

type Edges_schema struct {
	Parent_id uuid.UUID
	Child_id  uuid.UUID
}
