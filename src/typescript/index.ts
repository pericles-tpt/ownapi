import "../styles/styles.scss"
import '@fortawesome/fontawesome-free/js/fontawesome'
import '@fortawesome/fontawesome-free/js/solid'
import '@fortawesome/fontawesome-free/js/regular'
import '@fortawesome/fontawesome-free/js/brands'
import { ButtonType, Modal } from './components/modal';
import { App } from './pages/app';
import axios from "axios"
import { untemplateDivElement } from "./utility/render"
import { Pipeline, pipelineProgress, pipelineStatus, pipelineStatusNames, pipelineWSUpdate } from "./ds/pipeline/structure"
import { getNodeTypeLabel } from "./ds/pipeline/get"
import { getMSPreciseTimeString, getTimingStringFromUs, milliToDurationString } from "./utility/time"

let normalModal: Modal;
let errorModal: Modal;
let app: App | null = null;
let liveMode: boolean = false;

window.onload = () => {
    normalModal = new Modal();
    errorModal = new Modal(1);

    // Fetch pipeline list and populate sidebar
    const sidebarPipelinesContainer = document.querySelector('body div[pipelines]') as HTMLDivElement;
    axios({
        method: 'get',
        url: 'http://localhost:8080/pipelines/list',
    }).then(resp => {
        (resp.data as Array<string>).forEach(pipelineName => {
            const pipelineLabel = untemplateDivElement('sidebar-pipeline-tmpl')
            const pipelineLabelTitle = pipelineLabel.querySelector('[title]') as HTMLDivElement;
            pipelineLabelTitle.innerText = pipelineName;
            pipelineLabel.addEventListener("click", () => {
                const main = document.querySelector('main') as HTMLElement;
                axios({
                    method: 'get',
                    url: `http://localhost:8080/pipelines/content/${pipelineName}`
                }).then(resp => {
                    const pipeline = resp.data as Pipeline;
                    showPipelineInMain(pipeline);
                });
            })
            sidebarPipelinesContainer.appendChild(pipelineLabel)
        })
    })

    let pipelineNextRunAtMap: Map<string, number> = new Map<string, number>();
    const socket = new WebSocket("ws://localhost:8080/events")
    socket.onmessage = async function(e) {
        // TODO: Handle event
        
        // const arr = new Uint8Array(await (e.data as Blob).arrayBuffer());
        const pipelineStatuses = JSON.parse(e.data) as pipelineWSUpdate;

        pipelineStatuses.pipelines.forEach((pl, i) => {
            pipelineNextRunAtMap.set(pl.name, pipelineStatuses.pipelineStatuses[i].nextRunAtUnixMilli);
        })
        updatePipelineUI(pipelineStatuses, liveMode);

    };

    const liveModeCheckbox = document.querySelector('input[name="live-mode"]');
    liveModeCheckbox?.addEventListener('change', (e) => {
        const elem = e.target as HTMLInputElement;
        liveMode = elem.checked;
    });

    setInterval(function() {
        if (!liveMode) {
            return
        }
        const sidebarItems = document.querySelectorAll("[sidebar-item]") as NodeListOf<HTMLDivElement>;
        sidebarItems.forEach(it => {
            const name = (it.querySelector('[title]') as HTMLDivElement).innerText;
            let nextRunAt = pipelineNextRunAtMap.get(name);
            if (nextRunAt == undefined) {
                nextRunAt = -1
            }

            const countdown = it.querySelector('[countdown]') as HTMLDivElement;
            countdown.innerText = milliToDurationString(nextRunAt - Date.now(), 1);
        })
    }, 33);



    // Open websocket
    // TODO: Auth should occur so user only has access to the right pipelines


}

function updatePipelineUI(pipelineStatuses: pipelineWSUpdate, liveMode: boolean) {
    // TODO: This is a bit lazy, could probably avoid generating a map here but not bothered
    const pipelineMap: Map<string, Pipeline> = new Map<string, Pipeline>();
    const pipelineProgressMap: Map<string, pipelineProgress> = new Map<string, pipelineProgress>();
    for (let i = 0; i < pipelineStatuses.pipelines.length; i++) {
        const pl = pipelineStatuses.pipelines[i];
        const plp = pipelineStatuses.pipelineStatuses[i];
        pipelineMap.set(pl.name, pl);
        pipelineProgressMap.set(pl.name, plp);
    }
    
    // Live Mode
    if (liveMode) {
        // 1. Change main pane to the latest modified pipeline
        let lastUpdatedUnix: number = 0;
        let lastUpdatedIdx: number = -1;
        let i = 0;
        for (i = 0; i < pipelineStatuses.pipelineStatuses.length; i++) {
            const thisUpdated = pipelineStatuses.pipelineStatuses[i].lastUpdate;
            if (thisUpdated > lastUpdatedUnix) {
                lastUpdatedUnix = thisUpdated;
                lastUpdatedIdx = i
            }
        }
        if (lastUpdatedIdx < 0) {
            console.warn('failed to find last updated pipeline, probably an issue with the `lastUpdate` property')
            return
        }

        const lastUpdatedPipeline = pipelineStatuses.pipelines[lastUpdatedIdx];
        showPipelineInMain(lastUpdatedPipeline);

        // 2. Reorder sidebar items, depending on latest update
        const sidebarPipelinesContainer = document.querySelector('body div[pipelines]') as HTMLDivElement;
        const sidebarItems = Array.from(sidebarPipelinesContainer?.children) as HTMLDivElement[];
        sidebarItems.sort((a, b) => {
            const aName = (a.querySelector('[title]') as HTMLDivElement).innerText;
            const bName = (b.querySelector('[title]') as HTMLDivElement).innerText;
            const aProgress = pipelineProgressMap.get(aName);
            const bProgress = pipelineProgressMap.get(bName);
            if (aProgress == undefined || bProgress == undefined) {
                return 0;
            }
            const aLastUpdated = aProgress.lastUpdate
            const bLastUpdated = bProgress.lastUpdate 
            return bLastUpdated - aLastUpdated;
        });

        const frag = document.createDocumentFragment();
        sidebarItems.forEach(el => frag.appendChild(el));
        sidebarPipelinesContainer.appendChild(frag);
    }

    const activePipeline = document.querySelector('main').getAttribute('pipeline');
    let mainPipelineProgress: pipelineProgress;
    for (let i = 0; i < pipelineStatuses.pipelines.length; i++) {
        if (activePipeline == pipelineStatuses.pipelines[i].name) {
            mainPipelineProgress = pipelineStatuses.pipelineStatuses[i];
            break;
        }
    }
    let activePipelineHasSuccessAttribute = false;

    // IF success && activePipeline, don't change state

    // Sidebar for overall
    const sidebarItems = document.querySelectorAll("[sidebar-item]") as NodeListOf<HTMLDivElement>;
    sidebarItems.forEach(it => {
        const name = (it.querySelector('[title]') as HTMLDivElement).innerText;
        const progress = pipelineProgressMap.get(name);
        if (progress == undefined) {
            return
        }
        const status = progress.overallProgress;
        
        const isActivePipeline = name === activePipeline;
        if (isActivePipeline) {
            activePipelineHasSuccessAttribute = it.hasAttribute("success");
        }
        const notActivePipelineOrSuccess = (!isActivePipeline || !it.hasAttribute("success"))
        if (notActivePipelineOrSuccess) {
            pipelineStatusNames.forEach((n, i) => {
                const enabled = Number(status) === i;
                it.toggleAttribute(n, enabled);
            })
        }
    })

    // Main section for the active pipeline
    // if (!activePipelineHasSuccessAttribute) {
        const pipelineColumns = document.querySelectorAll('main [pipeline-col]') as NodeListOf<HTMLDivElement>;
        pipelineColumns.forEach((elem, i) => {
            const status = mainPipelineProgress.stagesProgress[i]
            pipelineStatusNames.forEach((sn, j) => {
                const enabled = Number(status) === j;
                elem.toggleAttribute(sn, enabled);
            })
    
            elem.querySelectorAll('[pipeline-node]').forEach((node, j) => {
                const status = mainPipelineProgress.nodesProgress[i][j];
                pipelineStatusNames.forEach((sn, k) => {
                    const enabled = Number(status) === k;
                    node.toggleAttribute(sn, enabled);

                    const timingElem = node.querySelector('[node-timing]') as HTMLDivElement;
                    timingElem.innerText = getTimingStringFromUs(mainPipelineProgress.nodesTimingUs[i][j]);
                })
            });
        })
    // }
}

function resetPipelineStatuses() {
    // Sidebar for overall
    const sidebarItems = document.querySelectorAll("[sidebar-item]") as NodeListOf<HTMLDivElement>;
    sidebarItems.forEach(it => {
        pipelineStatusNames.forEach(n => {
            it.toggleAttribute(n, false);
        })
    })

    // Main section for the active pipeline
    const pipelineColumns = document.querySelectorAll('main [pipeline-col]') as NodeListOf<HTMLDivElement>;
    pipelineColumns.forEach((elem, i) => {
        pipelineStatusNames.forEach((sn, j) => {
            elem.toggleAttribute(sn, false);
        })

        elem.querySelectorAll('[pipeline-node]').forEach((node, j) => {
            pipelineStatusNames.forEach((sn, k) => {
                node.toggleAttribute(sn, false);
            })
        });
    })
}

function manualRunPipeline(pipelineName: string) {
    const main = document.querySelector('main') as HTMLElement;
    axios({
        method: 'get',
        url: `http://localhost:8080/pipelines/content/${pipelineName}`
    }).then(resp => {
        showPipelineInMain(resp.data as Pipeline);
    })

    axios({
        method: 'put',
        url: `http://localhost:8080/pipelines/run/${pipelineName}`
    }).then(resp => {
        if (resp.status === 200) {
            console.log(`ran pipeline: ${pipelineName}!`);
        }
    })
}

export function setErrorModal(message: string, isError: boolean = true, doBeforeHide: () => void | null = null) {
    const title = isError ? "Error" : "Success";
    const content = document.createElement("div");
    content.innerText = message;
    errorModal.setErrorStyling(isError);

    errorModal.replaceContentAndGetPrimaryButton(title, content, ButtonType.Ok).addEventListener("click", () => {
        if (doBeforeHide !== null) {
            doBeforeHide();
        }
        errorModal.hide();
    })
}

function showPipelineInMain(pipeline: Pipeline) {
    const main = document.querySelector('main') as HTMLElement;
    const currPipeline = main.getAttribute("pipeline");
    if (currPipeline == pipeline.name) {
        return
    }
    main.innerHTML = '';

    const pipelineGrid = untemplateDivElement('pipeline-grid-tmpl');

    for (let i = 0; i < pipeline.nodes.length; i++) {
        const pipelineCol = untemplateDivElement('pipeline-col-tmpl');
        
        for (let j = 0; j < pipeline.nodes[i].length; j++) {
            const node = pipeline.nodes[i][j];
            
            const pipelineNode = untemplateDivElement('pipeline-node-tmpl');
            const nameChild = pipelineNode.querySelector('div[node-name]') as HTMLDivElement;
            nameChild.innerText = node.config.hash.substring(0, 5);
            const labelChild = pipelineNode.querySelector('div[node-label]') as HTMLDivElement;
            labelChild.innerText = getNodeTypeLabel(pipeline.nodeTypes[i][j]);

            pipelineCol.appendChild(pipelineNode);
        }

        pipelineGrid.appendChild(pipelineCol);
    }

    main.appendChild(pipelineGrid);
    main.setAttribute("pipeline", pipeline.name)
}