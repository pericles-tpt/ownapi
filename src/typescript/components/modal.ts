import { RemoveEventListeners, untemplateDivElement } from "../utility/render";

export enum ButtonType {
    Ok,
    Submit,
    Close,
    None
}

export class Modal {
    private baseElem: HTMLDivElement;
    private primaryButton: HTMLButtonElement;

    private closeable: boolean = true;

    constructor(zOffset?: number) {
        this.baseElem = untemplateDivElement("modal-tmpl");
        this.baseElem = document.body.appendChild(this.baseElem);
        this.hide();

        this.baseElem.addEventListener("click", e => {
            if (this.closeable && (e.target as HTMLElement).hasAttribute("modal-bg")) {
                this.hide();
            }
        });
        (this.baseElem.querySelector("[close]") as HTMLDivElement).addEventListener("click", () => {
            if (this.closeable) {
                this.hide();
            }
        });

        // From CSS base zIndex should be higher than everything else, only need to set offset for modals on modals
        if (zOffset !== undefined) {
            const originalZ = parseInt(this.baseElem.style.zIndex, 10);
            this.baseElem.style.zIndex = `${originalZ + zOffset}`
        }

        this.primaryButton = this.baseElem.querySelector("button[primary]");
    }

    public show() {
        this.baseElem.toggleAttribute("hidden", false);
    }

    public hide() {
        this.baseElem.toggleAttribute("hidden", true);
    }

    public replaceContent(newContentElem: HTMLElement): void {
        const modalContent = (this.baseElem.querySelector("[modal-content]") as HTMLDivElement);
        if (modalContent.children.length > 0) {
            modalContent.firstElementChild.replaceWith(newContentElem);
        } else {
            modalContent.appendChild(newContentElem)
        }
    }

    public getContent(): HTMLDivElement {
        return (this.baseElem.querySelector("[modal-content]") as HTMLDivElement)
    }

    public alignContent(align: string): void {
        if (!["start", "center", "end", "left", "right"].includes(align)) {
            return;
        }
        (this.baseElem.querySelector("[modal-content]") as HTMLDivElement).style.textAlign = align;
    }

    public deconstruct() {
        this.baseElem.remove();
    }

    public replaceContentAndGetPrimaryButton(title: string, content: HTMLElement, buttonType: ButtonType, isMini: boolean = false, hide: boolean = false, closeable: boolean = true, alignContent:string = 'left'): HTMLButtonElement {
        this.alignContent(alignContent);

        (this.baseElem.querySelector("[title]") as HTMLDivElement).innerText = title;
        this.replaceContent(content);
        
        this.closeable = closeable;
        (this.baseElem.querySelector("[close]") as HTMLDivElement).toggleAttribute("hidden", !closeable);

        if (isMini) {
            this.baseElem.querySelector("[modal]").toggleAttribute("mini", true);
        }

        if (hide) {
            this.hide();
        } else {
            this.show();
        }

        RemoveEventListeners(this.primaryButton);
        this.primaryButton = this.baseElem.querySelector("button[primary]");
        this.primaryButton.toggleAttribute("ok", buttonType === ButtonType.Ok);
        this.primaryButton.toggleAttribute("submit", buttonType === ButtonType.Submit);
        this.primaryButton.toggleAttribute("close", buttonType === ButtonType.Close);
        
        const primaryButtonText = buttonType === ButtonType.Ok ? "Ok" : (buttonType === ButtonType.Submit ? "Submit" : "Close")
        this.primaryButton.innerText = primaryButtonText;

        this.primaryButton.toggleAttribute("hidden", buttonType === ButtonType.None);

        return this.primaryButton;
    }

    public setErrorStyling(isError: boolean) {
        this.baseElem.querySelector("[modal]").toggleAttribute("error", isError)
        this.baseElem.querySelector("[modal]").toggleAttribute("success", !isError)
    }
}