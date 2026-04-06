export function getRandomInt(max: number) {
    return Math.floor(Math.random() * max);
}

export function Round(num: number, dp: number) {
    const mag = 10**dp;
    return Math.round(num * mag) / mag;
}