package pipelines

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pericles-tpt/ownapi2/node"
	"github.com/pkg/errors"
)

type Pipeline struct {
	Nodes     [][]node.BaseNode `json:"nodes"`
	NodeTypes [][]node.NodeType `json:"nodeTypes"`
}

type PipelineFile struct {
	Nodes     [][]map[string]any `json:"nodes"`
	NodeTypes [][]node.NodeType  `json:"nodeTypes"`
}

var (
	pipelinesMap      = map[string]Pipeline{}
	pipelinesProgress = map[string]PipelineProgress{}
)

func Load(filename string) ([]byte, map[string]Pipeline, error) {
	fbs, err := os.ReadFile(filename)
	if err != nil {
		return fbs, pipelinesMap, err
	}

	tmpMap := map[string]PipelineFile{}
	err = json.Unmarshal(fbs, &tmpMap)
	if err != nil {
		return fbs, pipelinesMap, err
	}

	for name, pipeline := range tmpMap {
		var (
			nodes     = pipeline.Nodes
			nodeTypes = pipeline.NodeTypes

			newPL = Pipeline{
				NodeTypes: nodeTypes,
				Nodes:     make([][]node.BaseNode, 0, len(nodes)),
			}

			numStageNodes = make([]int, 0, len(nodes))
		)
		for si, stage := range nodes {
			newPLStage := make([]node.BaseNode, 0, len(stage))
			for ni, node := range stage {
				bn, err := getBaseNode(node, nodeTypes[si][ni])
				if err != nil {
					// TODO: Could improve this instead of returning error here, collect errors across ALL pipelines, stages and nodes
					//		 return them all at once
					return fbs, pipelinesMap, err
				}
				newPLStage = append(newPLStage, bn)
			}
			numStageNodes = append(numStageNodes, len(newPLStage))
			newPL.Nodes = append(newPL.Nodes, newPLStage)
		}
		pipelinesMap[name] = newPL
		progress := PipelineProgress{
			OverallProgress: NotRunning,
			StagesProgress:  make([]PipelineStatus, len(numStageNodes)),
			NodesProgress:   make([][]PipelineStatus, len(numStageNodes)),

			StagesTimingUs: make([]int64, len(numStageNodes)),
			NodesTimingUs:  make([][]int64, len(numStageNodes)),
		}
		totalNumNodes := 0
		for i, num := range numStageNodes {
			progress.StagesProgress[i] = NotRunning
			progress.NodesProgress[i] = make([]PipelineStatus, num)
			progress.NodesTimingUs[i] = make([]int64, num)
			for j := range num {
				progress.NodesProgress[i][j] = NotRunning
			}
			totalNumNodes += num
		}
		progress.NodeErrors = make([]NodeError, 0, totalNumNodes)
		pipelinesProgress[name] = progress
	}

	return fbs, pipelinesMap, nil
}

func getBaseNode(maybeNode any, nodeType node.NodeType) (node.BaseNode, error) {
	var ret node.BaseNode

	bs, err := json.Marshal(maybeNode)
	if err != nil {
		return ret, err
	}

	propMap := map[string]any{}
	switch nodeType {
	case node.Http:
		maybeHN := node.HTTPNode{}
		err = json.Unmarshal(bs, &maybeHN)
		if err != nil {
			return ret, errors.Wrap(err, "failed to unmarshal JSON for Http node type")
		}
		hn, err := node.CreateHTTPNode(propMap, maybeHN.Config)
		if err == nil {
			return &hn, nil
		}
	case node.Json:
		maybeJN := node.JSONNode{}
		err = json.Unmarshal(bs, &maybeJN)
		if err != nil {
			return ret, errors.Wrap(err, "failed to unmarshal JSON for Json node type")
		}
		jn, err := node.CreateJSONNode(propMap, maybeJN.Config)
		if err == nil {
			return &jn, nil
		}
	}
	return ret, fmt.Errorf("failed to find node matching type: %v", nodeType)
}
