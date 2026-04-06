// DATA CHECKS
export function getFirstLineArgs(content: string, separator: string): string[] {
    const lines = content.split("\n").filter(str => {return str !== ""})
    return lines.length === 0 ? [] : splitOnSeparator(lines[0], separator)
}

export function getAllLineArgs(content: string, argSeparator: string, lineSeparator: string): string[][] {
    const ret: string[][] = [];
    const lines: string[] = content.split(lineSeparator).filter(str => {return str !== ""})
    lines.forEach(l => {
        // TODO: Add quotes back in the future?
        const quotesRemovedLine = l.replace(/"/g, "");
        ret.push(splitOnSeparator(quotesRemovedLine, argSeparator).filter(str => {return str !== ""}))
    })
    return ret;
}

export function RemoveWhitespaceAndEmptyArgs(args: string[]) {
    args.forEach(a => {
		a.trim()
	})
	args.filter(a => {
		return a !== ""
	})

    return args
}

export function RemoveDigitsFromString(s: string) {
    const nonDigitsArray = s.split("").filter(c => {
        return isNaN(parseInt(c));
    });
    return nonDigitsArray.join("")
}

/*
Split data entry lines on their separator.

Importantly handles quotes when using the [SPACE] as a separator
*/
export function splitOnSeparator(line: string, separator: string): string[] {
    let words: string[] = [];
    line.split(separator).forEach(w => {
        words.push(w.trim());
    })
    if (separator === " ") {
        const stitchedQuoteSections: string[] = [];
        let startStitch = false;
        let stitchBuffer: string[] = [];
        words.forEach(w => {
            let curr = w

            if (w[0] === "\"") {
                startStitch = true
                stitchBuffer.push(w.slice(1, w.length).trim())
                return
            } else if (w[w.length-1] === "\"") {
                stitchBuffer.push(w.slice(0, w.length-1).trim())
                curr = stitchBuffer.join(" ")
                stitchBuffer = [];
                startStitch = false
            }

            if (startStitch) {
                stitchBuffer.push(w.trim())
            } else {
                stitchedQuoteSections.push(curr.trim())
            }
        })
        words = stitchedQuoteSections
    }

    return words
}

export function Truncate(s: string, maxLength: number): string {
    if (s.length > maxLength) {
        return s.slice(0, maxLength - 3) + "..."
    }
    return s
}

export function getListAsString(elems: string[]): string {
    return elems.length === 1 ? elems[0] : `${elems.slice(0, elems.length - 1).join(", ")} and ${elems[elems.length - 1]}`
}

// Source: https://stackoverflow.com/a/13627586
export function getOrdinal(num: number): string {
    let j = num % 10,
        k = num % 100;
    if (j === 1 && k !== 11) {
        return num + "st";
    }
    if (j === 2 && k !== 12) {
        return num + "nd";
    }
    if (j === 3 && k !== 13) {
        return num + "rd";
    }
    return num + "th";
}

/*
    - spaces replaced with underscores
    - lowercase
*/
export function convertToFileName(s: string): string {
    return s.toLowerCase().replace(/[\s,\t,\n]+/g,"_")
}