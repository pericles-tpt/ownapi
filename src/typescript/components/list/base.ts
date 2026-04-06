export abstract class BaseList<Item> {
    protected root: HTMLElement;
    private messageElem: HTMLDivElement;
    private readonly itemsElem: HTMLDivElement;

    private state: Item[] = [];

    private itemsMatchFunc: (a: Item, itemsToDelete: Item[]) => boolean;

    constructor(root: HTMLElement, items: Item[], itemsMatchFunc: (a: Item, itemsToMatch: Item[]) => boolean) {
        this.root = root;
        this.root.toggleAttribute("items-list", true);

        const messageElem = document.createElement("div");
        messageElem.toggleAttribute("message", true);
        messageElem.toggleAttribute("hidden", true);
        this.messageElem = messageElem;
        this.root.appendChild(this.messageElem);
        const itemsElem = document.createElement("div");
        itemsElem.toggleAttribute("items", true);
        this.itemsElem = itemsElem;
        this.root.appendChild(this.itemsElem);

        this.itemsMatchFunc = itemsMatchFunc;

        this.replace(items);
    }

    public add(items: Item[], addToBottom: boolean = true) {
        const itemElems: HTMLElement[] = []
        items.forEach(it => {
            itemElems.push(this.generateHtmlFromItem(it));
            if (addToBottom) {
                this.state.push(it);
            } else {
                this.state.unshift(it);
            }
        })
        
        if (addToBottom) {
            this.itemsElem.append(...itemElems);
        } else {
            this.itemsElem.prepend(...itemElems);
        }
    }

    public delete(items: Item[]) {
        const newItems: Item[] = []
        this.state.forEach(it => {
            if (!this.itemsMatchFunc(it, items)) {
                newItems.push(it);
            }
        })
        this.replace(newItems);
        this.maybeShowEmptyMessage();
    }

    public modify(items: Item[]) {
        this.state.forEach((it, i) => {
            if (!this.itemsMatchFunc(it, items)) {
                this.state[i] = it;
            }
        });
        this.replace(this.state);
        this.maybeShowEmptyMessage();
    }

    public replace(newState: Item[]) {
        this.state = newState;
        this.itemsElem.innerHTML = "";
        this.state.forEach(it => {
            const elem = this.generateHtmlFromItem(it);
            this.itemsElem.appendChild(elem);
        })
        this.maybeShowEmptyMessage();
    }

    public filter(searchTerms: string[], preFilter: (it: Item) => boolean | null = null) {
        const itemsToHide: number[] = [];
        this.state.forEach((it, i) => {
            const content = this.getItemContent(it);
            let match = false;
            for (let j = 0; j < searchTerms.length; j++) {
                const matchesPrefilter = preFilter === null || preFilter(it);
                if (matchesPrefilter && content.includes(searchTerms[j])) {
                    match = true;
                    break;
                }
            }
            if (!match) {
                itemsToHide.push(i);
            }
        });

        let currHideIdx = itemsToHide.length === 0 ? -1 : 0;
        for (let i = 0; i < this.itemsElem.children.length; i++) {
            const hide = currHideIdx > -1 && currHideIdx < itemsToHide.length && itemsToHide[currHideIdx] === i;
            this.itemsElem.children[i].toggleAttribute("hidden", hide);
            if (hide) {
                currHideIdx++;
            }
        }
        if (itemsToHide.length === this.state.length) {
            this.showHideNoneFoundMessage(true);
        } else {
            this.showHideNoneFoundMessage(false);
        }
    }

    public getItemsElem(): HTMLDivElement {
        return this.itemsElem;
    }

    private maybeShowEmptyMessage() {
        const show = this.state.length === 0;
        if (show) {
            this.messageElem.innerText = "empty"
        }
        this.messageElem.toggleAttribute("hidden", !show);
        this.itemsElem.toggleAttribute("hidden", show);
    }

    private showHideNoneFoundMessage(show: boolean) {
        this.messageElem.innerText = "no items found from search";
        this.messageElem.toggleAttribute("hidden", !show);
        this.itemsElem.toggleAttribute("hidden", show);
    }

    protected abstract generateHtmlFromItem(it: Item): HTMLElement;
    protected abstract getItemContent(it: Item): string;
}