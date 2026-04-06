import "../styles/styles.scss"
import '@fortawesome/fontawesome-free/js/fontawesome'
import '@fortawesome/fontawesome-free/js/solid'
import '@fortawesome/fontawesome-free/js/regular'
import '@fortawesome/fontawesome-free/js/brands'
import { ButtonType, Modal } from './components/modal';
import { App } from './pages/app';
import axios from "axios"
import { untemplateDivElement } from "./utility/render"
import { Pipeline, pipelineProgress, pipelineStatus, pipelineStatusNames } from "./ds/pipeline/structure"
import { getNodeTypeLabel } from "./ds/pipeline/get"
import { getMSPreciseTimeString, getTimingStringFromUs } from "./utility/time"

let normalModal: Modal;
let errorModal: Modal;
let app: App | null = null;

window.onload = () => {
    normalModal = new Modal();
    errorModal = new Modal(1);

    // Fetch pipeline list and populate sidebar
    const sidebarPipelinesContainer = document.querySelector('body div[pipelines]')
    console.log('sidebar pipelines container: ', sidebarPipelinesContainer)
    axios({
        method: 'get',
        url: 'http://localhost:8080/pipelines/list',
    }).then(resp => {
        (resp.data as Array<string>).forEach(pipelineName => {
            const pipelineLabel = untemplateDivElement('sidebar-pipeline-tmpl')
            pipelineLabel.innerText = pipelineName;
            pipelineLabel.addEventListener("click", () => {
                showPipelineInMain(pipelineName);
            })
            sidebarPipelinesContainer.appendChild(pipelineLabel)
        })
    })

    const socket = new WebSocket("ws://localhost:8080/events")
    socket.onmessage = async function(e) {
        // TODO: Handle event
        
        // const arr = new Uint8Array(await (e.data as Blob).arrayBuffer());
        const pipelineStatuses = new Map<string,pipelineProgress>(Object.entries(JSON.parse(e.data)));
        console.log("update: ", pipelineStatuses)

        updatePipelineUI(pipelineStatuses);

    };

    // Open websocket
    // TODO: Auth should occur so user only has access to the right pipelines


    console.log('hello world!')
}

function updatePipelineUI(pipelineStatuses: Map<string, pipelineProgress>) {
    const activePipeline = document.querySelector('main').getAttribute('pipeline');
    const mainPipeline = pipelineStatuses.get(activePipeline);
    let activePipelineHasSuccessAttribute = false;

    // IF success && activePipeline, don't change state

    // Sidebar for overall
    const sidebarItems = document.querySelectorAll("[sidebar-item]") as NodeListOf<HTMLDivElement>;
    sidebarItems.forEach(it => {
        const name = it.innerText;
        const status = pipelineStatuses.get(name).overallProgress;
        
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
    if (!activePipelineHasSuccessAttribute) {
        const pipelineColumns = document.querySelectorAll('main [pipeline-col]') as NodeListOf<HTMLDivElement>;
        pipelineColumns.forEach((elem, i) => {
            const status = mainPipeline.stagesProgress[i]
            pipelineStatusNames.forEach((sn, j) => {
                const enabled = Number(status) === j;
                elem.toggleAttribute(sn, enabled);
            })
    
            elem.querySelectorAll('[pipeline-node]').forEach((node, j) => {
                const status = mainPipeline.nodesProgress[i][j];
                pipelineStatusNames.forEach((sn, k) => {
                    const enabled = Number(status) === k;
                    node.toggleAttribute(sn, enabled);

                    const timingElem = node.querySelector('[node-timing]') as HTMLDivElement;
                    timingElem.innerText = getTimingStringFromUs(mainPipeline.nodesTimingUs[i][j]);
                })
            });
        })
    }
}

function resetPipelineStatuses() {
    // Sidebar for overall
    const sidebarItems = document.querySelectorAll("[sidebar-item]") as NodeListOf<HTMLDivElement>;
    sidebarItems.forEach(it => {
        console.log(it.innerText)
        console.log('old it: ', it);
        pipelineStatusNames.forEach(n => {
            it.toggleAttribute(n, false);
        })
        console.log('new it: ', it);
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

function showPipelineInMain(pipelineName: string) {
    const main = document.querySelector('main')
    axios({
        method: 'get',
        url: `http://localhost:8080/pipelines/content/${pipelineName}`
    }).then(resp => {
        main.innerHTML = '';

        const pipeline = resp.data as Pipeline;
        const pipelineGrid = untemplateDivElement('pipeline-grid-tmpl');

        for (let i = 0; i < pipeline.nodes.length; i++) {
            const pipelineCol = untemplateDivElement('pipeline-col-tmpl');

            for (let j = 0; j < pipeline.nodes[i].length; j++) {
                const node = pipeline.nodes[i][j];
                
                const pipelineNode = untemplateDivElement('pipeline-node-tmpl');
                const nameChild = pipelineNode.querySelector('div[node-name]') as HTMLDivElement;
                nameChild.innerText = node.hash.substring(0, 5);
                const labelChild = pipelineNode.querySelector('div[node-label]') as HTMLDivElement;
                labelChild.innerText = getNodeTypeLabel(pipeline.nodeTypes[i][j]);

                pipelineCol.appendChild(pipelineNode);
            }

            pipelineGrid.appendChild(pipelineCol);
        }

        main.appendChild(pipelineGrid);
        main.setAttribute("pipeline", pipelineName)

        // TODO: Create a grid of nodes with the following information:
        // - status (border: grey, red, green)
        // - type (label: http, json)
        // - name: innertext
        // Where each column is a STAGE of NODES
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