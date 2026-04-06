export function DeepCopy(object: any): any {
    return JSON.parse(JSON.stringify(object));
}

export function recurseStringifyJSON(json: any): string {
    let ret = ""
    if (Array.isArray(json)) {
        (json as any[]).forEach(elem => {
            ret += recurseStringifyJSON(elem)
        })
    } else {
        ret += JSON.stringify(json)
    }
    return ret
}