package pipelines

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pericles-tpt/ownapi/node"
	"github.com/pkg/errors"
)

type Pipeline struct {
	Name      string            `json:"name"`
	Nodes     [][]node.BaseNode `json:"nodes"`
	NodeTypes [][]node.NodeType `json:"nodeTypes"`
}

type PipelineFile struct {
	Name      string             `json:"name"`
	Nodes     [][]map[string]any `json:"nodes"`
	NodeTypes [][]node.NodeType  `json:"nodeTypes"`
}

var (
	pipelines         = []Pipeline{}
	pipelinesProgress = []PipelineProgress{}
)

func Load(filename string) ([]byte, []Pipeline, error) {
	fbs, err := os.ReadFile(filename)
	if err != nil {
		return fbs, pipelines, err
	}

	tmpArr := []PipelineFile{}
	err = json.Unmarshal(fbs, &tmpArr)
	if err != nil {
		return fbs, pipelines, err
	}

	pipelines = make([]Pipeline, len(tmpArr))
	pipelinesProgress = make([]PipelineProgress, len(tmpArr))

	for i, pipeline := range tmpArr {
		var (
			nodes     = pipeline.Nodes
			nodeTypes = pipeline.NodeTypes

			newPL = Pipeline{
				Name:      pipeline.Name,
				NodeTypes: nodeTypes,
				Nodes:     make([][]node.BaseNode, 0, len(nodes)),
			}

			numStageNodes = make([]int, 0, len(nodes))
		)
		for si, stage := range nodes {
			newPLStage := make([]node.BaseNode, 0, len(stage))
			for ni, node := range stage {
				bn, err := getBaseNode(node, nodeTypes[si][ni], true)
				if err != nil {
					// TODO: Could improve this instead of returning error here, collect errors across ALL pipelines, stages and nodes
					//		 return them all at once
					return fbs, pipelines, err
				}
				newPLStage = append(newPLStage, bn)
			}
			numStageNodes = append(numStageNodes, len(newPLStage))
			newPL.Nodes = append(newPL.Nodes, newPLStage)
		}

		pipelines[i] = newPL
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
		pipelinesProgress[i] = progress
	}

	return fbs, pipelines, nil
}

func LoadPipeline(nodes [][]map[string]any, types [][]node.NodeType, reload bool) (Pipeline, []int, error) {
	var (
		pl = Pipeline{
			NodeTypes: types,
			Nodes:     make([][]node.BaseNode, 0, len(nodes)),
		}
		numStageNodes = make([]int, 0, len(nodes))
	)
	if len(nodes) != len(types) {
		return pl, numStageNodes, fmt.Errorf("unequal length between `nodes` and `types`, %d != %d", len(nodes), len(types))
	}

	for si, stage := range nodes {
		newPLStage := make([]node.BaseNode, 0, len(stage))
		for ni, node := range stage {
			bn, err := getBaseNode(node, types[si][ni], reload)
			if err != nil {
				// TODO: Could improve this instead of returning error here, collect errors across ALL pipelines, stages and nodes
				//		 return them all at once
				return pl, numStageNodes, err
			}
			newPLStage = append(newPLStage, bn)
		}
		numStageNodes = append(numStageNodes, len(newPLStage))
		pl.Nodes = append(pl.Nodes, newPLStage)
	}
	return pl, numStageNodes, nil
}

func getBaseNode(maybeNode any, nodeType node.NodeType, reload bool) (node.BaseNode, error) {
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
		if err != nil {
			return ret, errors.Wrap(err, "failed to create Http node")
		}
		return &hn, nil
	case node.Json:
		maybeJN := node.JSONNode{}
		err = json.Unmarshal(bs, &maybeJN)
		if err != nil {
			return ret, errors.Wrap(err, "failed to unmarshal JSON for Json node type")
		}
		jn, err := node.CreateJSONNode(propMap, maybeJN.Config)
		if err != nil {
			return ret, errors.Wrap(err, "failed to create JSON node")
		}
		return &jn, nil
	case node.UsbCopy:
		maybeUC := node.USBCopyFromNode{}
		err = json.Unmarshal(bs, &maybeUC)
		if err != nil {
			return ret, errors.Wrap(err, "failed to unmarshal JSON for USBCopy node type")
		}
		jn, err := node.CreateUSBCopyFromNode(propMap, maybeUC.Config, reload)
		if err != nil {
			return ret, errors.Wrap(err, "failed to create USBCopy node")
		}
		return &jn, nil
	}
	return ret, fmt.Errorf("failed to find node matching type: %v", nodeType)
}
