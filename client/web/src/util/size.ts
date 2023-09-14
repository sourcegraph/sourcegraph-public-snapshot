const ONE_KILO_BYTE = 1_000
const ONE_MEGA_BYTE = 1_000_000
const ONE_GIGA_BYTE = 1_000_000_000
const ONE_TERA_BYTE = 1_000_000_000_000

/**
 * Converts a byte size to a human readable string.
 *
 * @param size - The size in bytes to convert.
 * @returns A human readable string representing the size (e.g. "1.23MB").
 */
export function humanizeSize(size: number): string {
    if (size > ONE_TERA_BYTE) {
        const estimatedSize = size / ONE_TERA_BYTE
        return `${estimatedSize.toFixed(2)}TB`
    }

    if (size > ONE_GIGA_BYTE) {
        const estimatedSize = size / ONE_GIGA_BYTE
        return `${estimatedSize.toFixed(2)}GB`
    }

    if (size > ONE_MEGA_BYTE) {
        const estimatedSize = size / ONE_MEGA_BYTE
        return `${estimatedSize.toFixed(2)}MB`
    }

    if (size > ONE_KILO_BYTE) {
        const estimatedSize = size / ONE_KILO_BYTE
        return `${estimatedSize.toFixed(2)}KB`
    }

    return `${size}B`
}
