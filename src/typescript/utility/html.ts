export function elemOrParentsHaveAttribute(elem: HTMLElement, attribute: string, maxIterations: number = 5): boolean {
    if (elem.hasAttribute(attribute)) {
        return true
    } else if (elem.parentElement === null) {
        return false
    }
    return elemOrParentsHaveAttribute(elem.parentElement, attribute, maxIterations-1);
}

export function getElemOrParentWithNodeType(elem: HTMLElement, nodeName: string, maxIterations: number = 5): (HTMLElement | null) {
    if (elem === null) {
        return null
    } else if (elem.nodeName.toUpperCase() === nodeName.toUpperCase()) {
        return elem;
    }
    return getElemOrParentWithNodeType(elem.parentElement, nodeName, maxIterations-1);
}

export function getElemOrParentWithAttribute(elem: HTMLElement, attribute: string, maxIterations: number = 5): (HTMLElement | null) {
    if (elem === null) {
        return null
    } else if (elem.hasAttribute(attribute)) {
        return elem;
    }
    return getElemOrParentWithAttribute(elem.parentElement, attribute, maxIterations-1);
}

export function getCursorPosition(inp: HTMLInputElement): [number, number] {
    const textUpToCursor = inp.value.slice(0, inp.selectionStart);

    const tmpDiv = document.createElement("div");
    tmpDiv.toggleAttribute("tmp-div", true);
    tmpDiv.style.cssText = document.defaultView.getComputedStyle(inp, "").cssText;
    tmpDiv.innerText = textUpToCursor;
    document.body.appendChild(tmpDiv);
    const width     = tmpDiv.clientWidth;
    tmpDiv.remove();

    return [inp.getBoundingClientRect().left + width, inp.getBoundingClientRect().top];
}

// NOTE: Don't use this function for multiple children of the same parent, use `getChildrenAndWidthsFromHiddenElement` instead
export function getChildAndWidthFromHiddenElement(root: HTMLElement, selector: string): [HTMLElement | null, number] {
    root.toggleAttribute("hidden", false);
    const child = root.querySelector(selector) as HTMLElement;
    const width = (child === null ? 0 : child.clientWidth);
    root.toggleAttribute("hidden", true);
    return [child, width];
}

export function getChildrenAndWidthsFromHiddenElement(root: HTMLElement, selectors: string[]): [(HTMLElement | null)[], number[]] {
    const maybeChildren: (HTMLElement | null)[] = [];
    const widths: number[] = [];
    root.toggleAttribute("hidden", false);
    selectors.forEach(s => {
        const child = root.querySelector(s) as HTMLElement;
        maybeChildren.push(child);
        widths.push(child === null ? 0 : child.clientWidth);
    })
    root.toggleAttribute("hidden", true);
    return [maybeChildren, widths];
}