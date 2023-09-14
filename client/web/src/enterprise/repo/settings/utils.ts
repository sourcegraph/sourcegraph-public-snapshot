export const formatDuration = (duration: number): string => {
    if (duration > 1) {
        return `${duration.toFixed(2)}s`
    }

    const durationInMs = duration * 1000
    return `${durationInMs.toFixed(2)}ms`
}
