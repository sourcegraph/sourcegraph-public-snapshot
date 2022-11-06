const ONE_KILO_BYTE = 1000
const ONE_MEGA_BYTE = 1_000_000

export function humanizeSize(size: number): string {
    if (size > ONE_MEGA_BYTE) {
        const estimatedSize = size / ONE_MEGA_BYTE
        return `${estimatedSize.toFixed(1)}MB`
    }

    if (size > ONE_KILO_BYTE) {
        const estimatedSize = size / ONE_KILO_BYTE
        return `${estimatedSize.toFixed(1)}KB`
    }

    return `${size}B`
}
