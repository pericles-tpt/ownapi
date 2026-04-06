import { DeepCopy } from "./object";

export function GetAllArraysEqualToFirstLength(arr: any[][], returnOnFirstUnequal: boolean): any[][] {
    const ret: any[][] = [];
    if (arr.length === 0) {
        return ret
    }

    const firstArrayLength = arr[0].length
    ret.push(arr[0])
    for (let i = 1; i < arr.length; i++) {
        if (firstArrayLength !== arr[i].length) {
            if (returnOnFirstUnequal) {
                return ret
            }
            continue
        }
        ret.push(arr[i])
    }
    return ret
}

export function arraysEqualAsSets(a: any[], b: any[]) {
    if (a.length !== b.length) {
        return false
    }

    return JSON.stringify(DeepCopy(a).sort()) === JSON.stringify(DeepCopy(b).sort())
}

export function transposeMN(arr: any[][]): any[][] {
    if (arr.length === 0) {
        return [];
    }

    const ret: any[][] = [];
    for (let i = 0; i < arr[0].length; i++) {
        ret.push([]);
        for (let j = 0; j < arr.length; j++) {
            if (arr[j].length !== arr[0].length) {
                return [] // -> not an N x N array
            }
            ret[i].push(arr[j][i]);
        }
    }

    return ret;
}