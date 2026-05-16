package pipelines

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	log2 "github.com/pericles-tpt/ownapi/log"
	"github.com/pericles-tpt/ownapi/node"
	"github.com/pkg/errors"
)

var (
	pipelineProgressMutex = sync.RWMutex{}
)

func Run(name *string, idx *int, runFromNodeTrigger bool) (bool, error) {
	var (
		propMap = map[string]any{}
		exists  bool
		err     error
	)
	if name == nil && idx == nil {
		return exists, errors.New("at least one of `name` or `idx` must be non-nil")
	}

	var (
		pl Pipeline
		i  int
	)
	if name != nil {
		pl, i, err = GetPipelineByName(*name)
		if err != nil {
			return exists, errors.Wrapf(err, "failed to get pipeline by name: %s", *name)
		}
		idx = &i
	} else if idx != nil {
		pl, err = GetPipelineByIdx(*idx)
		if err != nil {
			return exists, errors.Wrapf(err, "failed to get pipeline by idx: %d", *idx)
		}
	}

	propMap, exists, err = runPipeline(pl, *idx, propMap, runFromNodeTrigger)
	if err != nil {
		fmt.Printf("%s - FAILED TO RUN: %v\n", pl.Name, err)
		return exists, errors.Wrapf(err, "failed to run pipeline: %s", pl.Name)
	}
	return exists, nil
}

func runPipeline(pipeline Pipeline, idx int, propMap map[string]any, runFromNodeTrigger bool) (map[string]any, bool, error) {
	// TODO: These empty stages checks probably don't belong here
	var (
		err       error
		cancelRun bool
		nodes     = pipeline.Nodes
	)
	if len(nodes) == 0 {
		return propMap, cancelRun, errors.New("pipeline contains no stages")
	}
	emptyStages := make([]int, 0, len(nodes))
	for sn, stage := range nodes {
		if len(stage) == 0 {
			emptyStages = append(emptyStages, sn)
		}
	}
	if len(emptyStages) > 0 {
		return propMap, cancelRun, fmt.Errorf("the pipeline stages at the following indices contain no nodes: %v", emptyStages)
	}

	// Don't run if already running
	pipelineProgressMutex.Lock()
	isRunning := pipelinesProgress[idx].OverallProgress == Running
	pipelineProgressMutex.Unlock()
	if isRunning {
		return propMap, cancelRun, fmt.Errorf("pipeline '%s' already running", pipeline.Name)
	}

	// If auto-trigger then first node has already been run
	sn := 0
	if runFromNodeTrigger {
		updatePipelineProgress(idx, sn, 0, Running)
		propMap, err = nodes[0][0].Trigger(propMap)
		if err != nil {
			updatePipelineProgress(idx, sn, 0, Error)
			return propMap, cancelRun, errors.New("trigger node for pipeline failed to run")
		}

		// Has it changed? If not don't run rest of pipeline
		if !nodes[0][0].Changed(propMap) {
			updatePipelineProgress(idx, sn, 0, NotRunning)
			lt := log2.Manual
			if runFromNodeTrigger {
				lt = log2.Auto
			}
			log2.WriteLogs(lt, "SKIPPED", []string{}, [][]any{}, true)
			return propMap, true, nil
		}

		sn = 1
	}

	start := time.Now()
	var wg sync.WaitGroup
	for sn = 0; sn < len(nodes); sn++ {
		var err error
		stage := nodes[sn]
		bef := time.Now()

		wg.Add(len(stage))

		omMx := sync.RWMutex{}

		stageErrCh := make(chan error, len(stage))
		outputMaps := make([]map[string]any, len(stage))
		errs := make([]error, 0, len(stage))

		for nn, step := range stage {
			// Update RUNNING status
			updatePipelineProgress(idx, sn, nn, Running)
			go func(s node.BaseNode) {
				start := time.Now()
				defer wg.Done()

				// Execute the step with context
				var err error
				nodeStatus := Success
				omMx.Lock()
				if outputMaps[nn], err = s.Trigger(propMap); err != nil {
					// Send error to stage-specific error channel
					stageErrCh <- err
					nodeStatus = Error
				}
				omMx.Unlock()
				// Update SUCCESS | ERRROR status
				took := time.Since(start).Microseconds()
				completeNodeProgress(idx, sn, nn, nodeStatus, took, err)
			}(step)
		}

		go func() {
			wg.Wait()
			close(stageErrCh)
		}()

		for err := range stageErrCh {
			errs = append(errs, err)
		}

		omMx.Lock()
		for _, om := range outputMaps {
			for k, v := range om {
				propMap[k] = v
			}
		}
		omMx.Unlock()

		propMap, err = node.UpdateKeys(propMap, sn)
		if err != nil {
			errs = append(errs, err)
		}

		if len(errs) > 0 {
			fmt.Printf("Error(s) occurred at pipeline stage %d: %v", sn, errs)
			break
		}

		fmt.Printf("propMap types at END of stage: %d\n", sn)
		for k, v := range propMap {
			to := reflect.TypeOf(v).String()
			if strings.HasPrefix(to, "[]") || strings.HasPrefix(to, "map[") {
				fmt.Printf("k: %s, tv: %s\n", k, reflect.TypeOf(v))
			} else {
				fmt.Printf("k: %s, tv: %v\n", k, v)
			}
		}

		took := time.Since(bef).Microseconds()
		setPipelineStageTimingUs(idx, sn, took)
	}
	took := time.Since(start).Microseconds()
	setPipelineOverallTimingUs(idx, took)

	time.Sleep(time.Millisecond)

	resetPipelineProgress(idx)
	fmt.Println("FINISHED PIPELINE!")

	return propMap, cancelRun, nil
}

func completeNodeProgress(idx int, stageNo int, nodeNo int, status PipelineStatus, durationMicroseconds int64, nodeError error) {
	pipelineProgressMutex.Lock()
	oldProgress := pipelinesProgress[idx]
	pipelineProgressMutex.Unlock()
	oldProgress.NodesProgress[stageNo][nodeNo] = status

	switch status {
	case Error:
		oldProgress.OverallProgress = status
		oldProgress.StagesProgress[stageNo] = status
		oldProgress.NodeErrors = append(oldProgress.NodeErrors, NodeError{
			NodeIdx:  nodeNo,
			StageIdx: stageNo,
			Error:    nodeError.Error(),
		})
	case Success:
		if nodeNo == len(oldProgress.NodesProgress[stageNo])-1 {
			oldProgress.StagesProgress[stageNo] = status
			if stageNo == len(oldProgress.StagesProgress)-1 {
				oldProgress.OverallProgress = Success
			}
		}
	}

	oldProgress.NodesTimingUs[stageNo][nodeNo] = durationMicroseconds

	pipelineProgressMutex.Lock()
	pipelinesProgress[idx] = oldProgress
	pipelineProgressMutex.Unlock()
}

func updatePipelineProgress(idx int, stageNo int, nodeNo int, status PipelineStatus) {
	pipelineProgressMutex.Lock()
	oldProgress := pipelinesProgress[idx]
	pipelineProgressMutex.Unlock()
	oldProgress.NodesProgress[stageNo][nodeNo] = status

	oldProgress.OverallProgress = status
	oldProgress.StagesProgress[stageNo] = status

	pipelineProgressMutex.Lock()
	pipelinesProgress[idx] = oldProgress
	pipelineProgressMutex.Unlock()
}

func resetPipelineProgress(idx int) {
	pipelineProgressMutex.Lock()
	oldProgress := pipelinesProgress[idx]
	pipelineProgressMutex.Unlock()
	oldProgress.OverallProgress = NotRunning
	for i := range oldProgress.StagesProgress {
		oldProgress.StagesProgress[i] = NotRunning
		for j := range oldProgress.NodesProgress[i] {
			oldProgress.NodesProgress[i][j] = NotRunning
		}
	}

	pipelineProgressMutex.Lock()
	pipelinesProgress[idx] = oldProgress
	pipelineProgressMutex.Unlock()
}

func setPipelineStageTimingUs(idx int, stageNo int, durationMicroseconds int64) {
	pipelineProgressMutex.Lock()
	oldProgress := pipelinesProgress[idx]
	pipelineProgressMutex.Unlock()

	oldProgress.StagesTimingUs[stageNo] = durationMicroseconds

	pipelineProgressMutex.Lock()
	pipelinesProgress[idx] = oldProgress
	pipelineProgressMutex.Unlock()
}

func setPipelineOverallTimingUs(idx int, durationMicroseconds int64) {
	pipelineProgressMutex.Lock()
	oldProgress := pipelinesProgress[idx]
	pipelineProgressMutex.Unlock()

	oldProgress.OverallTimingUs = durationMicroseconds

	pipelineProgressMutex.Lock()
	pipelinesProgress[idx] = oldProgress
	pipelineProgressMutex.Unlock()
}
