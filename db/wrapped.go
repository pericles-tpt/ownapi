package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AddPipelineParams struct {
	PipelineID    uuid.UUID
	NodesPerStage []int32
	Nodes         []uuid.UUID
}

func (q *Queries) AddPipeline(ctx context.Context, arg AddPipelineParams) error {
	params := addPipelineParams{
		PipelineID:    arg.PipelineID,
		NodesPerStage: arg.NodesPerStage,
		Nodes:         arg.Nodes,
		UpdatedAt:     time.Now(),
	}
	return q.addPipeline(ctx, params)
}
