const SYSTEM_TABLE_NAMES = ["drafts"]

export function consistentData(data: string[][]): boolean {
    if (data.length == 0) {
        return true
    }

    let numRowArgsEqual = true;
    const numRowArgsFirst = data[0].length
    data.forEach(args =>{
        numRowArgsEqual &&= (args.length === numRowArgsFirst)
    }) 

    return numRowArgsEqual;
}

export function getEnumKeys(e: any): string[] {
    return Object.keys(e).filter((v) => isNaN(Number(v)))
}

// Source: https://dev.to/nombrekeff/download-file-from-blob-21ho
export function DownloadBlobOrUrl(content: Blob | string, name: string): void {
    let blobUrl = `${content}`;
    if (typeof content !== "string") {
        // Convert your blob into a Blob URL (a special url that points to an object in the browser's memory)
        blobUrl = URL.createObjectURL(content as Blob);
    }
    
    // Create a link element
    const link = document.createElement("a");
    
    // Set link's href to point to the Blob URL
    link.href = blobUrl;
    link.download = name;
    
    // Append link to the body
    document.body.appendChild(link);
    
    // Dispatch click event on the link
    // This is necessary as link.click() does not work on the latest firefox
    link.dispatchEvent(
        new MouseEvent('click', { 
        bubbles: true, 
        cancelable: true, 
        view: window 
        })
    );
    
    // Remove link from body
    document.body.removeChild(link);
}

export function DownloadTextFile(content: string, name: string): void {
    const blobUrl = URL.createObjectURL(new Blob([content], {type: "text/plain"}));
    
    // Create a link element
    const link = document.createElement("a");
    
    // Set link's href to point to the Blob URL
    link.href = blobUrl;
    link.download = name;
    
    // Append link to the body
    document.body.appendChild(link);
    
    // Dispatch click event on the link
    // This is necessary as link.click() does not work on the latest firefox
    link.dispatchEvent(
        new MouseEvent('click', { 
        bubbles: true, 
        cancelable: true, 
        view: window 
        })
    );
    
    // Remove link from body
    document.body.removeChild(link);
}

export function GetSystemTableNames(): string[] {
    return SYSTEM_TABLE_NAMES
}

export function isMobile(): boolean {
    const screenWidth = screen.width;
    return screenWidth > 300 && screenWidth < 950;
}

export function isOnline(): boolean {
    // TODO: Implement this
    return true;
}

export function sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
}