import { AppBase } from "./appBase";

const SECTION_ORDER: string[] = [];

// App - handles section navigation and inter-section communication
export class App extends AppBase {
    // SECTIONS - Render HTML for state
    private sectionOrderMap: Map<string, number>  = new Map<string, number>();

    constructor(root: HTMLElement) {
        super(root, "notes app");
        
        // SECTIONS
        // this.inputSection      = new InputSection(root.querySelector("[input-section]"));
        SECTION_ORDER.forEach((s, i) => {
            this.sectionOrderMap.set(s, i);
        });
        super.populateNavsAndSections(
        [            
            // this.inputSection,
        ], 
        [
            // null,
        ],
        [
            // "a[input]",
        ]);
        
        // Navigate to initial section - input, activate or login
        let initialSection = this.sectionOrderMap.get("input")
        this.showSection(initialSection, null);
        
        this.addEventListeners();
    }

    protected addEventListeners() {
        console.error("UNIMPLEMENTED")
    }

    public goToLogin() {
        this.showSection(this.sectionOrderMap.get("login"), null);
    }
}