package pipelines

import (
	"fmt"
)

type PipelineStatus int

const (
	Running PipelineStatus = iota
	Error
	Success
	NotRunning
)

// Node Idx to
type NodeError struct {
	StageIdx int    `json:"stageIdx"`
	NodeIdx  int    `json:"nodeIdx"`
	Error    string `json:"error"`
}

type PipelineProgress struct {
	OverallProgress PipelineStatus     `json:"overallProgress"`
	StagesProgress  []PipelineStatus   `json:"stagesProgress"`
	NodesProgress   [][]PipelineStatus `json:"nodesProgress"`

	OverallTimingUs int64     `json:"overallTimingUs"`
	StagesTimingUs  []int64   `json:"stagesTimingUs"`
	NodesTimingUs   [][]int64 `json:"nodesTimingUs"`

	NodeErrors []NodeError `json:"nodeErrors"`
}

func GetPipelineNames() []string {
	names := make([]string, 0, len(pipelinesMap))
	for name := range pipelinesMap {
		names = append(names, name)
	}
	return names
}

func GetPipleine(name string) (Pipeline, bool, error) {
	var (
		pl     Pipeline
		exists bool
	)
	pl, exists = pipelinesMap[name]
	if !exists {
		return pl, exists, fmt.Errorf("pipeline '%s' not found", name)
	}

	return pl, exists, nil
}

func GetPipelinesStatuses() map[string]PipelineProgress {
	return pipelinesProgress
}

func PipelineExists(name string) bool {
	_, ok := pipelinesMap[name]
	return ok
}
