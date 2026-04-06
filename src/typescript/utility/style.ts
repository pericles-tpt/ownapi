export function getStyle(elem: HTMLElement, propertyKey: string): string {
    return window.getComputedStyle(elem, null).getPropertyValue(propertyKey)
}