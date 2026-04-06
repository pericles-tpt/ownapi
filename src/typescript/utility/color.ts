const GOOD_CONTRAST_RATIO = 10;

// Source: https://www.w3.org/TR/WCAG20/#relativeluminancedef
export function getRGBLuminance(rgb: number[]): (number | null) {
    if (!validRGB(rgb)) {
        return null
    }

    const copy: number[]= [];

    rgb.forEach(c => {
        const cd = c / 255
        copy.push(cd <= 0.03928 ? cd / 12.92 : (((cd + 0.055) / 1.055) ** 2.4))
    })

    return 0.2126 * copy[0] + 0.7152 * copy[1] + 0.0722 * copy[2]
}

export function interpolateColors(rgbA: number[], rgbB: number[], currNum: number, maxNum: number): (number[] | null) {
    if (!(validRGB(rgbA) && (validRGB(rgbB)))) {
        return null
    }

    const bScale = (currNum / maxNum)
    const res = [
        Math.floor(rgbA[0] + (rgbA[0] > rgbB[0] ? -1 : 1) * (bScale * Math.abs(rgbA[0] - rgbB[0]))),
        Math.floor(rgbA[1] + (rgbA[1] > rgbB[1] ? -1 : 1) * (bScale * Math.abs(rgbA[1] - rgbB[1]))),
        Math.floor(rgbA[2] + (rgbA[2] > rgbB[2] ? -1 : 1) * (bScale * Math.abs(rgbA[2] - rgbB[2]))),
    ]

    return res
}

export function getGoodContrastGrayscaleColor(otherColorLuminance: number): (number[] | null) {
    let thisColorLuminance = ((otherColorLuminance + 0.05) / GOOD_CONTRAST_RATIO) - 0.05
    if (otherColorLuminance < 0.5) {
        thisColorLuminance = (GOOD_CONTRAST_RATIO * (otherColorLuminance + 0.05)) - 0.05
    }

    return [
        Math.floor(255 * thisColorLuminance),
        Math.floor(255 * thisColorLuminance),
        Math.floor(255 * thisColorLuminance),
    ]
}

// Needs to have 3 values, each in range (0, 255)
export function validRGB(vals: number[]): boolean {
    if (vals.length != 3) {
        return false
    }

    vals.forEach(v =>{
        if (v < 0 || v > 255) {
            return false
        }
    })

    return true
}

const AVAILABLE_COLORS = ["blue", "green", "yellow", "red", "orange", "purple", "black", "gray", "pink", "brown"];
export function GetNDifferentColors(n: number): string[] {
    const ret = new Array<string>(n);
    for (let i = 0; i < n; i++) {
        ret[i] = AVAILABLE_COLORS[i % AVAILABLE_COLORS.length];
    }
    return ret;
}

// source: https://stackoverflow.com/a/56266358
export function isValidColor(strColor: string) {
    const s = new Option().style;
    s.color = strColor;
    return s.color !== '';
}