import { Modal } from "../../components/modal";

export abstract class Section {
    protected root: HTMLElement;
    private title: string;
    protected modal: Modal;

    constructor(root: HTMLElement, title: string) {
        this.root = root;
        this.title = title;
        this.modal = new Modal();
    }

    
    public getSectionTitle(): string {
        return this.title;
    }
    
    public show() {
        this.root.toggleAttribute("hidden", false);
    }
    public hide() {
        this.root.toggleAttribute("hidden", true);
    }
}