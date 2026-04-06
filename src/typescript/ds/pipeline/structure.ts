export type Pipeline = {
    nodes: BaseNode[][],
    nodeTypes: NodeType[][],
};

export type BaseNode = {
    hash: string,
    config: any,
}

export enum NodeType {
    Http = 0,
    Json
}

export type nodeError = {
    stageIdx: number,
    nodeIdx: number,
    error: string,
}

export type pipelineProgress = {
    overallProgress: pipelineStatus,
    stagesProgress: pipelineStatus[],
    nodesProgress: pipelineStatus[][],

    overallTimingUs: number,
    stagesTimingUs: number[],
    nodesTimingUs: number[][],

    nodesErrors: nodeError[],
}

export enum pipelineStatus {
    running = 0,
    error,
    success,
    notRunning,
}

export const pipelineStatusNames = ["running", "error", "success", "notRunning"]