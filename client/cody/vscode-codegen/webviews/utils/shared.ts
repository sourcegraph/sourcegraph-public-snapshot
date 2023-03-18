export function getShortTimestamp(): string {
    const date = new Date()
    return `${padTimePart(date.getHours())}:${padTimePart(date.getMinutes())}`
}

function padTimePart(timePart: number): string {
    return timePart < 10 ? `0${timePart}` : timePart.toString()
}
