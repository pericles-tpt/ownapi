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
	names := make([]string, 0, len(pipelines))
	for _, pl := range pipelines {
		names = append(names, pl.Name)
	}
	return names
}

func GetPipelineByName(name string) (Pipeline, int, error) {
	var pl Pipeline
	for i, pl := range pipelines {
		if name == pl.Name {
			return pl, i, nil
		}
	}
	return pl, -1, fmt.Errorf("pipeline '%s' not found", name)
}

func GetPipelineByIdx(idx int) (Pipeline, error) {
	var pl Pipeline
	if idx < 0 || idx >= len(pipelines) {
		return pl, fmt.Errorf("idx: %d, out of range must be 0 <= idx < %d", idx, len(pipelines))
	}
	return pipelines[idx], nil
}

func GetPipelinesStatuses() []PipelineProgress {
	return pipelinesProgress
}

func PipelineExists(name string) bool {
	for _, pl := range pipelines {
		if name == pl.Name {
			return true
		}
	}
	return false
}
