// Copied from https://github.com/cenfun/nice-ticks/blob/f310f77280fae72706f8b1f956d01029f0c4f133/src/index.js
// The published package didn't import properly in webpack, but it's very small.

function toStr(str: number): string {
    return '' + str
}

// return float part length
function fLen(n: number): number {
    const s = toStr(n)
    const a = s.split('.')
    if (a.length > 1) {
        return a[1].length
    }
    return 0
}
// return float as int
function fInt(n: number): number {
    const s = toStr(n)
    return parseInt(s.replace('.', ''), 10)
}

function add(n1: number, n2: number): number {
    const r1 = fLen(n1)
    const r2 = fLen(n2)
    if (r1 + r2 === 0) {
        return n1 + n2
    }
    const m = Math.pow(10, Math.max(r1, r2))
    return (Math.round(n1 * m) + Math.round(n2 * m)) / m
}

function mul(n1: number, n2: number): number {
    const r1 = fLen(n1)
    const r2 = fLen(n2)
    if (r1 + r2 === 0) {
        return n1 * n2
    }
    const m1 = fInt(n1)
    const m2 = fInt(n2)
    return (m1 * m2) / Math.pow(10, r1 + r2)
}

function nice(x: number, round: boolean): number {
    const exp = Math.floor(Math.log(x) / Math.log(10))
    const f = x / Math.pow(10, exp)
    let nf
    if (round) {
        if (f < 1.5) {
            nf = 1
        } else if (f < 3) {
            nf = 2
        } else if (f < 7) {
            nf = 5
        } else {
            nf = 10
        }
    } else if (f <= 1) {
        nf = 1
    } else if (f <= 2) {
        nf = 2
    } else if (f <= 5) {
        nf = 5
    } else {
        nf = 10
    }
    return nf * Math.pow(10, exp)
}

const toNum = function (num: number | string): number {
    if (typeof num !== 'number') {
        num = parseFloat(num)
    }
    if (isNaN(num)) {
        num = 0
    }
    return num
}

/**
 * Calculate nice ticks for a chart axis.
 *
 * @param min The lowest data point.
 * @param max The highest data point.
 * @param num The number of desired ticks.
 * @returns An array of tick values.
 */
export const niceTicks = function (min: number, max: number, num = 4): number[] {
    min = toNum(min)
    max = toNum(max)
    num = toNum(num)

    if (min === max) {
        max = min + 1
    } else if (min > max) {
        const n = min
        min = max
        max = n
    }

    const r = nice(max - min, false)
    const d = nice(r / (num - 1), true)
    const s = mul(Math.floor(min / d), d)
    const e = mul(Math.ceil(max / d), d)
    const arr = []
    let v = s
    while (v <= e) {
        arr.push(v)
        v = add(v, d)
    }
    return arr
}
