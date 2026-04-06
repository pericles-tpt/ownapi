import { Round } from "./math";

const MILLI_TO_SEC = 1000;
const MILLI_TO_MIN = MILLI_TO_SEC * 60;
const MILLI_TO_HOUR = MILLI_TO_MIN * 60;
const MILLI_TO_DAY = MILLI_TO_HOUR * 24;

const unitToMicrosecondsMap = new Map<string, number>(
    [
        ['us', 1],
        ['ms', 1000],
        ['s' , 1000_000],
        ['m', 1000_000 * 60],
        ['h', 1000_000 * 60 * 60],
        ['d', 1000_000 * 60 * 60 * 24],
    ]
)

export function milliToDurationString(nv: number): string {
    const stringParts: string[] = [];
    
    const days = Math.floor(nv / MILLI_TO_DAY);
    nv -= days * MILLI_TO_DAY;
    if (days > 0) {
        stringParts.push(`${days}`.padStart(2, "0"))
    }
    const hours = Math.floor(nv / MILLI_TO_HOUR);
    nv -= hours * MILLI_TO_HOUR;
    if (hours > 0 || stringParts.length > 0) {
        stringParts.push(`${hours}`.padStart(2, "0"))
    }
    const minutes = Math.floor(nv / MILLI_TO_MIN);
    nv -= minutes * MILLI_TO_MIN;
    if (minutes > 0 || stringParts.length > 0) {
        stringParts.push(`${minutes}`.padStart(2, "0"))
    }
    if (minutes < 1) {
        stringParts.push("00")
    }
    const seconds = Math.floor(nv / MILLI_TO_SEC);
    nv -= seconds * MILLI_TO_SEC;
    stringParts.push(`${seconds}`.padStart(2, "0"))
    const milliseconds = nv;

    return stringParts.join(":") + "." + `${milliseconds}`.padStart(3, "0")
}

export function formatGoDatetime(datetime: string): string {
    const d = new Date(datetime);
    return d.toLocaleString();
}

export function numMinutesToTimeString(minutes: number): string {
    return `${Math.floor(minutes / 60)}:${(minutes % 60) < 10 ? `0${(minutes % 60)}` : (minutes % 60)}`
}

export function getISOStringWithTZ(date: Date): string {
    const dateStringNoTZ = date.toISOString().split(".")[0];
    const tzOffsetMinutes = (date.getTimezoneOffset() * -1);
    const tzString = (tzOffsetMinutes > 0 ? "+" : "-") + numMinutesToTimeString(Math.abs(tzOffsetMinutes))

    return `${dateStringNoTZ}${tzString}`;
}

export function getDatetimeLocalValidTimeString(date: Date): string {
    // Source: https://stackoverflow.com/a/61082536
    const ret = new Date(date);
    ret.setMinutes(ret.getMinutes() - ret.getTimezoneOffset(), ret.getSeconds(), ret.getMilliseconds());
    return ret.toISOString().slice(0,16);
}

export function getMSPreciseTimeString(date: Date): string {
    // Source: https://stackoverflow.com/a/61082536
    const ret = new Date(date);
    ret.setMinutes(ret.getMinutes() - ret.getTimezoneOffset(), ret.getSeconds(), ret.getMilliseconds());
    return ret.toISOString().slice(0,23);
}

export function before(a: Date, b: Date): boolean {
    return a.getTime() < b.getTime()
}
export function after(a: Date, b: Date): boolean {
    return a.getTime() > b.getTime()
}
export function equal(a: Date, b: Date): boolean {
    return new Date(a).getTime() === new Date(b).getTime();
}
export function elapsedMS(since: Date): number {
    return new Date().getTime() - since.getTime();
}

export function getTimingStringFromUs(us: number): string {
    let unit = 'd';
    let denom = unitToMicrosecondsMap.get(unit);
    if (us < 10**3) {
        unit = 'us';
    } else if (us < 10**6) {
        unit = 'ms';
    } else if (us < 10**6 * 60) {
        unit = 's';
    } else if (us < 10**6 * 60 * 60) {
        unit = 'm';
    } else if (us < 10**6 * 60 * 60 * 24) {
        unit = 'h';
    } else {
        unit = 'd';
    }
    denom = unitToMicrosecondsMap.get(unit);
    return `${Round(us / denom, 2)}${unit}`
}