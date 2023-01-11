/**
 * Function used to print a number of bytes in a more readable format.
 *
 * It returns an integer value for sizes of magnitudes of bytes, kb and MB.
 * For gigabytes and further it also includes 2 numbers after a decimal point if the size is not integral.
 *
 * @param bytes         Size in bytes.
 *
 * @returns Pretty printed size.
 */
export function prettyBytesBigint(bytes: bigint): string {
    let unit = 0
    const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
    const threshold = BigInt(1000)
    // two decimal points
    let decimal = 0

    while (bytes >= threshold) {
        // `bytes % threshold` is always less than 1000, so the conversion to Number is safe
        decimal = Number(bytes % threshold)
        // taking first two digits
        decimal = Math.floor(decimal / 10)
        bytes /= threshold
        unit += 1
    }
    const resultNumber = bytes.toString() + (decimal !== 0 && unit > 2 ? `.${decimal.toString().padStart(2, '0')}` : '')
    return resultNumber + ' ' + units[unit]
}
