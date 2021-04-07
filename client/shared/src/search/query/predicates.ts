// Represents recognized predicates associated with fields.
export const PREDICATES: Record<string, string[]> = {
    ['repo']: ['contains'],
    ['r']: ['contains'],
}

// Represents a predicates components corresponding to the syntax name(parameters).
export interface Predicate {
    name: string
    parameters: string
}

export const scanBalancedParens = (input: string): string | undefined => {
    let adjustedStart = 0
    let balanced = 0
    let current = ''
    const result: string[] = []

    const nextChar = (): void => {
        current = input[adjustedStart]
        adjustedStart += 1
    }

    while (input[adjustedStart] !== undefined) {
        nextChar()
        if (current === '(') {
            balanced += 1
            result.push(current)
        } else if (current === ')') {
            balanced -= 1
            result.push(current)
        } else if (current === '\\') {
            if (input[adjustedStart] !== undefined) {
                nextChar() // consume escaped
                result.push('\\', current)
                continue
            }
            result.push(current)
        } else {
            result.push(current)
        }
    }

    if (balanced !== 0) {
        return undefined
    }
    return result.join('')
}

/**
 * Scans predicate syntax of the form field:name(parameters) and
 * returns the name and parameters components. It checks that:
 *
 * (1) The (field, name) pair is a recognized predicate.
 * (2) The parameters value is well-balanced.
 */
export const scanPredicate = (field: string, value: string): Predicate | undefined => {
    const match = value.match(/^[\da-z]+/i)
    if (!(match && PREDICATES[field] && PREDICATES[field].some(pred => pred === match[0]))) {
        return undefined
    }
    const name = match[0]
    const rest = value.slice(name.length)
    if (!rest.startsWith('(') || !rest.endsWith(')')) {
        return undefined
    }
    const parameters = scanBalancedParens(rest)
    if (!parameters) {
        return undefined
    }

    return { name, parameters }
}
