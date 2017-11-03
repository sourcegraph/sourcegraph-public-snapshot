import parse from 'date-fns/parse'

export function parseCommitDateString(dateString: string): Date {
    return parse(
        dateString,
        'YYYY-MM-DD HH:mm:ss ZZ',
        new Date()
    )
}
