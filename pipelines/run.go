package pipelines

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/pericles-tpt/ownapi/node"
	"github.com/pkg/errors"
)

var (
	pipelineProgressMutex = sync.RWMutex{}
)

func Run(name string) (bool, error) {
	propMap := map[string]any{}
	propMap, exists, err := runPipeline(name, propMap)
	if err != nil {
		fmt.Println("FAILED TO RUN: ", err)
		return exists, errors.Wrap(err, "failed to run pipeline")
	}
	return exists, nil
}

func runPipeline(name string, propMap map[string]any) (map[string]any, bool, error) {
	pl, exists := pipelinesMap[name]
	if !exists {
		return propMap, exists, fmt.Errorf("pipeline '%s' not found", name)
	}

	// Don't run if already running
	pipelineProgressMutex.Lock()
	isRunning := pipelinesProgress[name].OverallProgress == Running
	pipelineProgressMutex.Unlock()
	if isRunning {
		return propMap, exists, fmt.Errorf("pipeline '%s' already running", name)
	}

	start := time.Now()
	var wg sync.WaitGroup
	for sn, stage := range pl.Nodes {
		start := time.Now()
		var err error

		wg.Add(len(stage))

		omMx := sync.RWMutex{}

		stageErrCh := make(chan error, len(stage))
		outputMaps := make([]map[string]any, len(stage))
		errs := make([]error, 0, len(stage))

		for nn, step := range stage {
			// Update RUNNING status
			updatePipelineProgress(name, sn, nn, Running)
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
				completeNodeProgress(name, sn, nn, nodeStatus, took, err)
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

		took := time.Since(start).Microseconds()
		setPipelineStageTimingUs(name, sn, took)
	}
	took := time.Since(start).Microseconds()
	setPipelineOverallTimingUs(name, took)

	time.Sleep(time.Millisecond)

	resetPipelineProgress(name)
	fmt.Println("FINISHED PIPELINE!")

	return propMap, exists, nil
}

func completeNodeProgress(name string, stageNo int, nodeNo int, status PipelineStatus, durationMicroseconds int64, nodeError error) {
	pipelineProgressMutex.Lock()
	oldProgress := pipelinesProgress[name]
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
	pipelinesProgress[name] = oldProgress
	pipelineProgressMutex.Unlock()
}

func updatePipelineProgress(name string, stageNo int, nodeNo int, status PipelineStatus) {
	pipelineProgressMutex.Lock()
	oldProgress := pipelinesProgress[name]
	pipelineProgressMutex.Unlock()
	oldProgress.NodesProgress[stageNo][nodeNo] = status

	oldProgress.OverallProgress = status
	oldProgress.StagesProgress[stageNo] = status

	pipelineProgressMutex.Lock()
	pipelinesProgress[name] = oldProgress
	pipelineProgressMutex.Unlock()
}

func resetPipelineProgress(name string) {
	pipelineProgressMutex.Lock()
	oldProgress := pipelinesProgress[name]
	pipelineProgressMutex.Unlock()
	oldProgress.OverallProgress = NotRunning
	for i := range oldProgress.StagesProgress {
		oldProgress.StagesProgress[i] = NotRunning
		for j := range oldProgress.NodesProgress[i] {
			oldProgress.NodesProgress[i][j] = NotRunning
		}
	}

	pipelineProgressMutex.Lock()
	pipelinesProgress[name] = oldProgress
	pipelineProgressMutex.Unlock()
}

func setPipelineStageTimingUs(name string, stageNo int, durationMicroseconds int64) {
	pipelineProgressMutex.Lock()
	oldProgress := pipelinesProgress[name]
	pipelineProgressMutex.Unlock()

	oldProgress.StagesTimingUs[stageNo] = durationMicroseconds

	pipelineProgressMutex.Lock()
	pipelinesProgress[name] = oldProgress
	pipelineProgressMutex.Unlock()
}

func setPipelineOverallTimingUs(name string, durationMicroseconds int64) {
	pipelineProgressMutex.Lock()
	oldProgress := pipelinesProgress[name]
	pipelineProgressMutex.Unlock()

	oldProgress.OverallTimingUs = durationMicroseconds

	pipelineProgressMutex.Lock()
	pipelinesProgress[name] = oldProgress
	pipelineProgressMutex.Unlock()
}
