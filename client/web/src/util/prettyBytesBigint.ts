export function prettyBytesBigint(bytes: bigint): string {
    let unit = 0
    const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
    const threshold = BigInt(1000)

    while (bytes >= threshold) {
        bytes /= threshold
        unit += 1
    }

    return bytes.toString() + ' ' + units[unit]
}
