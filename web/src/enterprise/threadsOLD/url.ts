/**
 * Constructs and returns a new threads query by merging values into an existing query.
 */
export function threadsQueryWithValues(query: string, values: { [field: string]: string[] | null }): string {
    const tokens = parse(query).tokens.filter(({ field }) => !field || !(field in values))
    const newTokens = Object.keys(values).flatMap(field => {
        const c = values[field]
        return c ? c.map(value => ({ field, value })) : []
    })
    return [...tokens, ...newTokens]
        .map(({ field, value }) => (field ? `${field}:${value}` : value))
        .filter(v => !!v)
        .join(' ')
}

/**
 * Reports whether the specified threads query contains all of the values.
 */
export function threadsQueryMatches(query: string, values: { [field: string]: string }): boolean {
    const { fieldValues } = parse(query)
    for (const [field, value] of Object.entries(values)) {
        const queryValues = fieldValues.get(field)
        if (!queryValues || !queryValues.includes(value)) {
            return false
        }
    }
    return true
}

/** Returns the URL to the threads query. */
export function urlToThreadsQuery(query: string): string {
    const params = new URLSearchParams({ q: query })
    return `/threads?${params.toString()}`
}

interface Token {
    field?: string
    value: string
}

function parse(
    query: string
): {
    tokens: Token[]
    fieldValues: Map<string, string[]>
} {
    const tokens = query.split(/\s+/g).map(s => {
        const i = s.indexOf(':')
        return i === -1 ? { value: s } : { field: s.slice(0, i), value: s.slice(i + 1) }
    })

    const fieldValues = new Map<string, string[]>()
    for (const token of tokens) {
        if (token.field) {
            const values = fieldValues.get(token.field)
            if (values) {
                values.push(token.value)
            } else {
                fieldValues.set(token.field, [token.value])
            }
        }
    }

    return { tokens, fieldValues }
}
