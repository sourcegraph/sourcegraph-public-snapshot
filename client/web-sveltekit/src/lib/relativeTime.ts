// in miliseconds
const units = {
    year: 24 * 60 * 60 * 1000 * 365,
    month: (24 * 60 * 60 * 1000 * 365) / 12,
    day: 24 * 60 * 60 * 1000,
    hour: 60 * 60 * 1000,
    minute: 60 * 1000,
    second: 1000,
} satisfies Partial<Record<Intl.RelativeTimeFormatUnit, number>>

const rtf = new Intl.RelativeTimeFormat('en', { numeric: 'auto' })

export function getRelativeTime(date1: Date, date2: Date = new Date()): string {
    const elapsed = date1.getTime() - date2.getTime()

    for (const unit in units) {
        if (Math.abs(elapsed) > units[unit as keyof typeof units]) {
            return rtf.format(
                Math.round(elapsed / units[unit as keyof typeof units]),
                unit as Intl.RelativeTimeFormatUnit
            )
        }
    }
    return rtf.format(Math.round(elapsed / units.second), 'second')
}
