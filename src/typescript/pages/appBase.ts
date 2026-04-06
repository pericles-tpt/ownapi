import { Modal } from "../components/modal";
import { Section } from "./sections/section";
import { elemOrParentsHaveAttribute, getChildrenAndWidthsFromHiddenElement } from "../utility/html"
import { replaceIcon } from "../utility/render";
import { isMobile } from "../utility/utlity";

const HAMBURGER_NAV_LABEL_WIDTH = 160;

export abstract class AppBase {
    protected root: HTMLElement;
    private title: HTMLElement;
    protected modal: Modal;

    private darkModeToggle: HTMLButtonElement;
    private hamburgerToggle: HTMLAnchorElement;
    private hamburgerMenu: HTMLDivElement;

    private appName: string;
    protected hamburgerMenuShown: boolean = false;

    private appSections: Section[] = [];
    private appSectionNavElems: (HTMLElement | null)[][] = [];

    protected lastSectionShown: number = 0;
    private lastSectionTitle: string;
    private lastProps: any[] = [];

    constructor(root: HTMLElement, appName: string) {
        this.root = root;
        this.title = root.querySelector("[title]");
        this.modal = new Modal();
        this.appName = appName;
        this.lastSectionTitle = appName;
        
        this.darkModeToggle = this.root.querySelector("[ld-mode]");
        this.hamburgerToggle = this.root.querySelector("[hamburger]");
        this.hamburgerMenu = this.root.querySelector("[hamburger-menu]");

        // Base event listeners
        this.darkModeToggle.addEventListener("click", (e) => this.toggleLightDarkMode(e));
        this.hamburgerToggle.addEventListener("click", () => this.toggleHamburgerMenu());
    }

    protected abstract addEventListeners(): void

    protected populateNavsAndSections(sections: Section[], showSectionProps: (() => any[] | Promise<any[]> | null)[], hamburgerMenuNavSelectors: string[], otherNavSelectors: string[][] = []) {
        if (sections.length !== hamburgerMenuNavSelectors.length || sections.length !== showSectionProps.length) {
            console.warn("unequal number of sections and section nav selectors provided");
            return;
        }

        const [maybeHamburgerMenuNavElems,] = getChildrenAndWidthsFromHiddenElement(this.hamburgerMenu, hamburgerMenuNavSelectors.filter(n => n.length > 0));
        // FIX: fsr setting all elements to maxWidth sometimes isn't enough space for the largest element to fix on one line so adding a bit here
        maybeHamburgerMenuNavElems.forEach((_, i) => {
            maybeHamburgerMenuNavElems[i].style.width = `${HAMBURGER_NAV_LABEL_WIDTH}px`;
        })
        sections.forEach((s, i) => {
            if (s === null) {
                return;
            }
            const thisSectionNavElems: HTMLElement[] = [];

            // Hamburger nav
            const hasHamburgerNavElem = hamburgerMenuNavSelectors[i].length > 0;
            let elem: HTMLElement | null = hasHamburgerNavElem ? maybeHamburgerMenuNavElems[i] : null;
            if (hasHamburgerNavElem) {
                if (elem === null) {
                    console.warn(`failed to find elem to navigation to section: ${s.getSectionTitle()}, with selector: ${hamburgerMenuNavSelectors[i]}`);
                    return;
                }
                this.attachNavListener(elem, i, showSectionProps[i]);
            }
            thisSectionNavElems.push(elem);

            // Other nav
            if (otherNavSelectors.length === hamburgerMenuNavSelectors.length) {
                otherNavSelectors[i].forEach(ns => {
                    const elem = document.querySelector(ns) as HTMLElement;
                    if (elem === null) {
                        console.warn(`failed to find elem to navigation to section: ${s.getSectionTitle()}, with selector: ${ns}`);
                        return;
                    }
                    this.attachNavListener(elem, i, showSectionProps[i]);
                    thisSectionNavElems.push(elem);
                });
            }

            this.appSections.push(s);
            this.appSectionNavElems.push(thisSectionNavElems);
        });
    }

    // Event Functions
    protected toggleLightDarkMode(e: Event) {
        const isDarkMode = elemOrParentsHaveAttribute(e.target as HTMLElement, "dark");
        
        // Button style
        this.darkModeToggle.toggleAttribute("dark", !isDarkMode);
        this.darkModeToggle.toggleAttribute("light", isDarkMode);
        
        this.darkModeToggle.innerHTML = "";
        this.darkModeToggle.appendChild(document.createElement("div"));
        replaceIcon(this.darkModeToggle.lastElementChild as HTMLElement, isDarkMode ? "fa-solid fa-sun fa-sm fa-fw" : "fa-solid fa-moon fa-sm fa-fw")
        this.darkModeToggle.append(isDarkMode ? "light mode" : "dark mode")
    
        // Page style
        document.querySelector("body").toggleAttribute("light-mode", !isDarkMode);
        document.querySelector("header").toggleAttribute("light-mode", !isDarkMode);
        document.querySelector("[hamburger-menu]").toggleAttribute("light-mode", !isDarkMode);
    }

    private toggleHamburgerMenu() {
        this.hamburgerMenuShown = !this.hamburgerMenuShown;
        this.hamburgerMenu.toggleAttribute("hidden", !this.hamburgerMenuShown);
        
        replaceIcon(this.hamburgerToggle, `fa-solid ${this.hamburgerMenuShown ? "fa-xmark" : "fa-bars"} fa-sm fa-fw`, true)        
        this.updateTitle(this.hamburgerMenuShown ? this.appName : this.lastSectionTitle)
    }

    private updateTitle(newTitle: string) {
        this.title.innerText = newTitle;
    }

    private attachNavListener(target: HTMLElement, idx: number, maybePropFunc: (() => any[] | Promise<any[]> | null)) {
        target.addEventListener("click", () => {
            if (maybePropFunc === null) {
                this.showSection(idx, []);
            } else {
                Promise.resolve(maybePropFunc()).then(resp => {
                    this.showSection(idx, resp);
                }).catch(err => {
                    console.warn("failed to retrieve items for section from promise: ", err)
                    this.showSection(idx, []);
                })
            }
            if (isMobile()) {
                this.toggleHamburgerMenu();
            }
        });
    }

    protected showSection(idx: number, props: any) {
        if (idx < 0 || idx >= this.appSections.length) {
            console.warn(`section index out of range, num sections: ${this.appSections.length}, idx: ${idx}`)
            return;
        }

        this.appSections.forEach((s, i) => {
            if (i === idx) {
                this.updateTitle(s.getSectionTitle());
                s.show();
                this.lastProps = props;
                this.lastSectionShown = idx;
                this.lastSectionTitle = s.getSectionTitle();
                return;
            }
            s.hide();
        });
    }
}