import { Spinner } from "spin.js";

const spinnerOpts = {
    lines: 13, // The number of lines to draw
    length: 38, // The length of each line
    width: 17, // The line thickness
    radius: 45, // The radius of the inner circle
    scale: 0.25, // Scales overall size of the spinner
    corners: 1, // Corner roundness (0..1)
    speed: 1, // Rounds per second
    rotate: 0, // The rotation offset
    animation: 'spinner-line-fade-default', // The CSS animation name for the lines
    direction: 1, // 1: clockwise, -1: counterclockwise
    color: '#000', // CSS color or array of colors
    fadeColor: '#fff', // CSS color or array of colors
    shadow: '0 0 1px transparent', // Box-shadow for the lines
    zIndex: 2000000000, // The z-index (defaults to 2e9)
    className: 'spinner', // The CSS class to assign to the spinner
    position: 'relative', // Element positioning
    left: 'unset',
    top: 'unset',
}

const spinnerMap = new Map<string, Spinner>();
const inactiveSpinnerKeys: string[] = [];

// TODO: Currently manually added a "style" tag to html to load css for Spinner, try to load the css properly in `Spinner()`
//       This might help: https://github.com/fgnass/spin.js/issues/362#issuecomment-411818580
export function ReplaceWithStartSpinnerElem(elemToReplace: HTMLElement): string {
    let key: string;
    if (inactiveSpinnerKeys.length === 0) {
        key = AddSpinnerToMap();
    } else {
        key = inactiveSpinnerKeys.pop();
    }

    const spinner = spinnerMap.get(key).spin();
    // TODO: Not sure if I need to do below...
    spinnerMap.set(key, spinner);

    // TODO: Fix this
    spinner.el.style.margin = "30px auto 0 auto"
    elemToReplace.replaceWith(spinner.el);

    return key;
}

function AddSpinnerToMap(): string {
    const newSpinnerKey = spinnerMap.size.toString(16);
    const spinner = new Spinner(spinnerOpts);
    spinnerMap.set(newSpinnerKey, spinner);
    
    return newSpinnerKey;
}

export function StopSpinnerAtKey(k: string) {
    if (!spinnerMap.has(k)) {
        console.warn("unable to stop spinner, provided key doesn't exist in `spinnerMap`")
        return
    }

    const stoppedSpinner = spinnerMap.get(k).stop();
    inactiveSpinnerKeys.push(k);
    // TODO: Not sure if I need to do below...
    spinnerMap.set(k, stoppedSpinner);
}