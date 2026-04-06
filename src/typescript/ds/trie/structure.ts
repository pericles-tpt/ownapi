export const idxCharLookup = [
    "!","\"","#","$","%","&","'","(",")","*","+",",","-",".","/","0","1","2","3","4","5","6","7",
    "8","9",":",";","<","=",">","?","@","[", "]","^","_","`","a","b","c","d","e","f","g","h","i",
    "j","k","l","m","n","o","p","q","r","s","t","u","v","w","x","y","z","{","|","}","~"," "
];

export const charIdxLookup = new Map<string, number>();

idxCharLookup.forEach((c, i) => {
    charIdxLookup.set(c, i);
})

const numChars = charIdxLookup.size;

export type Trie = {
    isPopulatedAt: number[];
    nodes: [string, Trie][],
    wordEnd: WordEnd,
}

export type WordEnd = {
    wordEnd: boolean;
    arrIndex: number;
}

// NOTE: Currently just for "select"s
export type OnFieldPath = {
    index: number,
    wordEnd: boolean
}

export function getDefaultTrie(): Trie {
    return {
        isPopulatedAt: new Array<number>(numChars).fill(-1),
        nodes: [],
        wordEnd: {
            arrIndex: -1,
            wordEnd: false,
        },
    }
}

export function aLessThanOnCharIdxLookup(a: string, b: string): boolean {
    for (let i = 0; i < a.length; i++) {
        const an = charIdxLookup.get(a.charAt(i)) as number;
        const bn = charIdxLookup.get(b.charAt(i)) as number;
        if (an < bn) {
            return true
        } else if (an > bn) {
            return false
        }
    }
    return false
}