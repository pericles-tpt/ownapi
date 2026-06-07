package pipelines

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pericles-tpt/ownapi/node"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Pipeline struct {
	Name      string            `bson:"name" json:"name"`
	Nodes     [][]node.BaseNode `bson:"nodes" json:"nodes"`
	NodeTypes [][]node.NodeType `bson:"nodeTypes" json:"nodeTypes"`
}

type PipelineFile struct {
	Name      string             `bson:"name" json:"name"`
	Nodes     [][]map[string]any `bson:"nodes" json:"nodes"`
	NodeTypes [][]node.NodeType  `bson:"nodeTypes" json:"nodeTypes"`
}

type PipelineFileBSON struct {
	Pipelines []PipelineFile `bson:"pipelines"`
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

	// TODO: Ultimately should only read/write to BSON, just writing JSON -> BSON for now
	// 		 until I've implemented methods for the application to modify the pipelines
	// 		 (rather than manually modifying the JSON)
	err = writeToBSON(filename, tmpArr)
	if err != nil {
		return fbs, pipelines, errors.Wrap(err, "failed to write JSON to BSON")
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

			numStageNodes       = make([]int, 0, len(nodes))
			nextRunAtMS   int64 = 0
		)

		if len(nodes) != len(nodeTypes) {
			return fbs, pipelines, fmt.Errorf("unequal number of nodes and nodeTypes in pipeline '%s', %d != %d", pipeline.Name, len(nodes), len(nodeTypes))
		}

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

		if len(nodes) > 0 && len(nodes[0]) > 0 {
			maybeNodeTrigger := newPL.Nodes[0][0].GetTrigger()
			if maybeNodeTrigger != nil {
				nodeTriggerInterval := (*maybeNodeTrigger).EveryN
				nextTriggerIntervalMS := (time.Duration(nodeTriggerInterval) * MIN_AUTO_RUN_LOOP_FREQUENCY)
				nextRunAtMS = time.Now().Add(time.Duration(nextTriggerIntervalMS)).UnixMilli()
			}
		}

		pipelines[i] = newPL
		progress := PipelineProgress{
			OverallProgress: NotRunning,
			StagesProgress:  make([]PipelineStatus, len(numStageNodes)),
			NodesProgress:   make([][]PipelineStatus, len(numStageNodes)),

			StagesTimingUs:     make([]int64, len(numStageNodes)),
			NodesTimingUs:      make([][]int64, len(numStageNodes)),
			NextRunAtUnixMilli: nextRunAtMS,
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
	case node.Custom:
		maybeCN := node.CustomNode{}
		err = json.Unmarshal(bs, &maybeCN)
		if err != nil {
			return ret, errors.Wrap(err, "failed to unmarshal JSON for Custom node type")
		}
		cn, err := node.CreateCustomNode(propMap, maybeCN.Config)
		if err != nil {
			return ret, errors.Wrap(err, "failed to create Custom node")
		}
		return &cn, nil
	}
	return ret, fmt.Errorf("failed to find node matching type: %v", nodeType)
}

func writeToBSON(filename string, pipelines []PipelineFile) error {
	data := PipelineFileBSON{Pipelines: pipelines}
	bs, err := bson.Marshal(data)
	if err != nil {
		return err
	}

	parts := strings.Split(filename, ".")
	newFilename := fmt.Sprintf("%s.bson", strings.Join(parts[:len(parts)-1], "."))

	err = os.WriteFile(newFilename, bs, 0666)
	if err != nil {
		return err
	}
	return nil
}
