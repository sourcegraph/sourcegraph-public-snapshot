type Func<Args extends unknown[]> = (...args: Args) => void

export function debounce<Args extends unknown[]>(
    callback: (...args: Args) => void,
    durationMs: number = 10
): Func<Args> {
    let timeoutId: NodeJS.Timeout | null = null

    const callable = (...args: Args): void => {
        if (timeoutId !== null) {
            clearTimeout(timeoutId)
        }

        timeoutId = setTimeout(() => {
            callback(...args)
        }, durationMs)
    }

    return callable as unknown as Func<Args>
}
