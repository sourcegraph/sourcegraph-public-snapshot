export function isValidNumber(value: unknown): value is number {
    return value !== null && typeof value === 'number' && !Number.isNaN(value) && Number.isFinite(value)
}
