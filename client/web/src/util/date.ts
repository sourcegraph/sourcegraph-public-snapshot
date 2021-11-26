import { formatDistanceStrict, parseISO } from 'date-fns'
import { enUS } from 'date-fns/locale'

const formatDistanceLocale = {
    xDays: '{{count}}d',
    xHours: '{{count}}h',
    xMinutes: '{{count}}m',
    xMonths: '{{count}}mo',
    xSeconds: '{{count}}s',
    xYears: '{{count}}y',
}

const formatDistanceAbbr = (token: keyof typeof formatDistanceLocale, count: number): string =>
    formatDistanceLocale[token].replace('{{count}}', String(count))

export function formatDistanceShortStrict(date: number | Date | string): string {
    return formatDistanceStrict(typeof date === 'string' ? parseISO(date) : date, Date.now(), {
        locale: {
            ...enUS,
            formatDistance: (token, count) =>
                token === 'xSeconds' && count === 1 ? 'now' : formatDistanceAbbr(token, count),
        },
    })
}
