import { DeepCopy } from "../../utility/object";

export abstract class BaseForm<State> {
    protected root: HTMLDivElement;
    protected isCreateForm: boolean;

    private originalState: State | null;
    protected newState: State;

    constructor(root: HTMLDivElement, state: State | null) {
        this.root = root;
        this.reload(state);
        this.addEventListeners();
    }

    public reload(state: State | null): void {
        this.originalState = state;
        this.isCreateForm = this.originalState === null;
        this.newState = this.isCreateForm ? this.getDefaultState() : DeepCopy(state);

        this.initialisePropertiesAndHtml();

        this.updateFormFromState();
    }

    public reset(): void {
        this.reload(null);
    }

    public getState(): State | null {
        this.updateStateFromForm();
        if (!this.valid()) {
            return null
        }

        return this.newState;
    }

    public getRootElem(): HTMLDivElement {
        return this.root;
    }

    // NOTE: Copy below to subclass
    protected abstract getDefaultState(): State;
    // Attach/set initial state for form elements
    protected abstract initialisePropertiesAndHtml(): void;
    // Populate form elements from `state`
    protected abstract updateFormFromState(): void;
    // Update `state` from form elements
    protected abstract updateStateFromForm(): void;
    // Add event listeners for elements that CHANGE HTML STATE
    protected abstract addEventListeners(): void;
    protected abstract valid(): boolean;

    // HELPERS
}