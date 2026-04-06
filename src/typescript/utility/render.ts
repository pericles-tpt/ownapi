import { isMobile } from "./utlity";

export function GetSelectInput(optionValues: string[], selectedValueOrIndex?: string | number, optionNames?: string[]): HTMLSelectElement | null {
    if (optionNames && optionNames.length !== optionValues.length) {
        return null;
    }

    const elem = document.createElement('select');
    optionValues.forEach((v, i) => {
        const option = document.createElement('option');
        option.value = v;
        option.innerText = optionNames ? optionNames[i] : optionValues[i];
        if (selectedValueOrIndex) {
            if (typeof selectedValueOrIndex === "string" && (selectedValueOrIndex as string) === v) {
                option.toggleAttribute("selected", true);
            } else if (typeof selectedValueOrIndex === "number" && (selectedValueOrIndex as number) === i) {
                option.toggleAttribute("selected", true);
            }
        }
        elem.appendChild(option);
    })
    return elem;
}

export function UpdateSelectInput(target: HTMLSelectElement, optionValues: string[], selectedValueOrIndex?: string | number, optionNames?: string[]): void {
    if (optionNames && optionNames.length !== optionValues.length) {
        return null;
    }

    // Clear out old options before adding new
    Array.from(target.options).forEach(o => {
        o.remove();
    })

    optionValues.forEach((v, i) => {
        const option = document.createElement('option');
        option.value = v;
        option.innerText = optionNames ? optionNames[i] : optionValues[i];
        if (selectedValueOrIndex) {
            if (typeof selectedValueOrIndex === "string" && (selectedValueOrIndex as string) === v) {
                option.toggleAttribute("selected", true);
            } else if (typeof selectedValueOrIndex === "number" && (selectedValueOrIndex as number) === i) {
                option.toggleAttribute("selected", true);
            }
        }
        target.appendChild(option);
    })
}

// TODO: Why does this set the "hidden" attribute on, "thead[header]" and "tbody[data]" to true always???
export function untemplateTable(hideHeader: boolean = false, hideBody: boolean = false): HTMLDivElement {
    const template: HTMLTemplateElement = document.querySelector("#table-tmpl");
    const newElem = template.content.cloneNode(true).childNodes[1] as HTMLDivElement;

    newElem.querySelector("thead[header]").toggleAttribute("hidden", false);
    newElem.querySelector("tbody[data]").toggleAttribute("hidden", false);

    return newElem;
}

export function untemplateDivElement(id: string): HTMLDivElement {
    const template: HTMLTemplateElement = document.querySelector(`#${id}`);
    const newElem = template.content.cloneNode(true).childNodes[1] as HTMLDivElement;

    return newElem;
}


export function untemplateButtonElement(id: string): HTMLButtonElement {
    const template: HTMLTemplateElement = document.querySelector(`#${id}`);
    const newElem = template.content.cloneNode(true).childNodes[1] as HTMLButtonElement;

    return newElem;
}

// Source: https://stackoverflow.com/a/9251864
export function RemoveEventListeners(target: HTMLElement): void {
    const new_element = target.cloneNode(true);
    target.parentNode.replaceChild(new_element, target);
    return;
}

export function setPageTitle(text: string, imageUrl?: string) {
    (document.querySelector("div[app-header] div[page-title]") as HTMLDivElement).innerText = text;
    (document.querySelector("div[app-header]") as HTMLDivElement).style.backgroundImage = imageUrl !== null ? `url("${imageUrl === undefined ? "" : imageUrl}")` : "";
    (document.querySelector("div[app-header] div[page-title]") as HTMLDivElement).toggleAttribute("hidden", false);
    (document.querySelector("div[app-header] div[user]") as HTMLDivElement).toggleAttribute("hidden", isMobile());
}

export function clearPageTitle() {
    (document.querySelector("div[app-header] div[page-title]") as HTMLDivElement).innerText = "";
    (document.querySelector("div[app-header]") as HTMLDivElement).style.backgroundImage = ``;
    (document.querySelector("div[app-header] div[page-title]") as HTMLDivElement).toggleAttribute("hidden", true);
    (document.querySelector("div[app-header] div[user]") as HTMLDivElement).toggleAttribute("hidden", false);
}

export function generateTitlesContentsHTML(titles: string[], contents: (string | string[])[]): HTMLDivElement {
    const ret = document.createElement('div');
    ret.style.display = "flex";
    ret.style.flexDirection = "column";
    const numSections = titles.length;

    for (let i = 0; i < numSections; i++) {
        if (titles[i] !== "") {
            const title = document.createElement("h2");
            title.innerText = titles[i];
            ret.appendChild(title);
        }

        if (i < contents.length) {
            const is2d = (typeof contents[i] !== "string")
            const content = document.createElement(is2d ? "ul" : "p")
            if (is2d) {
                (contents[i] as string[]).forEach(bp => {
                    const li = document.createElement('li');
                    li.innerText = bp;
                    content.appendChild(li);
                })
            } else {
                content.innerText = (contents[i] as string);
            }
            ret.appendChild(content);
        }
    }

    return ret;
}

export function generateRadioInputs(name: string, options: string[], selectedOptionName: string, attachToElem?: HTMLDivElement): HTMLDivElement {
    const ret = attachToElem === undefined ? document.createElement("div") : attachToElem;
    ret.toggleAttribute("radio-row", true);

    options.forEach(n => {
        if (n === "") {
            return;
        }

        const inputLabelDiv = document.createElement("div");

        const input = document.createElement("input");
        input.type = "radio";
        input.id = n;
        input.name = name;
        input.checked = n === selectedOptionName;
        input.value = n;

        const label = document.createElement("label");
        label.htmlFor = input.id;
        label.innerText = n;

        inputLabelDiv.appendChild(input);
        inputLabelDiv.appendChild(label);
        ret.appendChild(inputLabelDiv);
    });

    return ret;
}

export function generateCheckbox(name: string): HTMLDivElement {
    const ret = document.createElement("div");
    ret.toggleAttribute("cb-container", true);

    const input = document.createElement("input");
    input.type = "checkbox";
    input.id = name;

    const label = document.createElement("label");
    label.htmlFor = input.id;
    label.innerText = name;

    ret.appendChild(input);
    ret.appendChild(label);

    return ret;
}

export function generateInput(name: string, type: string = "text"): HTMLDivElement {
    const ret = document.createElement("div");
    ret.toggleAttribute("input-container", true);

    const input = document.createElement("input");
    input.type = type;
    input.id = name;

    const label = document.createElement("label");
    label.htmlFor = input.id;
    label.innerText = name;

    ret.appendChild(label);
    ret.appendChild(input);

    return ret;
}

export function generateScrollableListForSelect(selectOptions: string[]): HTMLUListElement {
    const ret = document.createElement("ul");
    ret.toggleAttribute("select-list")

    selectOptions.forEach(opt => {
        const li = document.createElement("li");
        li.innerText = opt;
        ret.appendChild(li);
    })

    return ret;

}

export function replaceIcon(elem: HTMLElement, iconClassName: string, replaceElemChildren: boolean = false) {
    const newIcon = document.createElement("i");
    newIcon.className = iconClassName;
    if (replaceElemChildren) {
        elem.replaceChildren(newIcon);
        return;
    }
    elem.replaceWith(newIcon);
}