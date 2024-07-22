/* eslint-disable no-template-curly-in-string */
import { type Completion, resolveFieldAlias, FilterType } from './filters'

interface PredicateDefinition {
    field: string
    name: string
}

// PREDICATES is a registry of predicates, grouped by field they belong to
export const PREDICATES: PredicateDefinition[] = [
    { field: 'repo', name: 'contains.file' },
    { field: 'repo', name: 'contains.path' },
    { field: 'repo', name: 'contains.content' },
    { field: 'repo', name: 'contains.commit.after' },
    { field: 'repo', name: 'contains.commit.after' },
    { field: 'repo', name: 'has' },
    { field: 'repo', name: 'has.file' },
    { field: 'repo', name: 'has.path' },
    { field: 'repo', name: 'has.content' },
    { field: 'repo', name: 'has.commit.after' },
    { field: 'repo', name: 'has.description' },
    { field: 'repo', name: 'has.tag' },
    { field: 'repo', name: 'has.key' },
    { field: 'repo', name: 'has.meta' },
    { field: 'repo', name: 'has.topic' },
    { field: 'file', name: 'contains.content' },
    { field: 'file', name: 'has.content' },
    { field: 'file', name: 'has.owner' },
    { field: 'rev', name: 'at.time' },
]

/** Represents a predicate's components corresponding to the syntax path(parameters). */
export interface PredicateInstance extends PredicateDefinition {
    parameters: string
}

// scans a string up to closing parentheses. Examples:
// - `foo` succeeds, parentheses are absent, so it is vacuously balanced
// - `foo(...)` succeeds up to the closing `)`
// - `foo(...))` succeeds up to the first `)`, which is recognized as the closing paren
// - `foo(` does not succeed, it is not balanced
// - `foo)` does not succeed, it is not balanced
const scanBalancedParens = (input: string): string | undefined => {
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
            if (balanced === 0) {
                // we've reached a closing parenthesis where the string is balanced
                break
            }
        } else if (current === '\\') {
            if (input[adjustedStart] !== undefined) {
                nextChar() // consume escaped
                result.push('\\', current)
                continue
            }
            result.push(current)
        } else if (current.match(/^\s+/) && balanced <= 0) {
            // Whitespace signals we've reached the end of this parenthesized expr.
            break
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
 * Scans predicate syntax of the form field:foo.bar(parameters) and
 * returns the name and parameters components. It checks that:
 *
 * (1) The (field, name) pair is a recognized predicate.
 * (2) The parameters value is well-balanced.
 */
export const scanPredicate = (field: string, value: string): PredicateInstance | undefined => {
    const match = value.match(/^[.a-z]+/i)
    if (!match) {
        return undefined
    }
    const name = match[0]
    // Remove negation from the field for lookup
    if (field.startsWith('-')) {
        field = field.slice(1)
    }
    field = resolveFieldAlias(field)
    const predicate = PREDICATES.find(predicate => predicate.field === field && predicate.name === name)
    if (!predicate) {
        return undefined
    }
    const rest = value.slice(predicate.name.length)
    const parameters = scanBalancedParens(rest)
    if (!parameters) {
        return undefined
    }
    if (!parameters.startsWith('(') || !parameters.endsWith(')')) {
        return undefined
    }

    return { ...predicate, parameters }
}

export const predicateCompletion = (field: FilterType): Completion[] => {
    if (field === 'repo') {
        return [
            {
                label: 'has.path(...)',
                insertText: 'has.path(${1:CHANGELOG})',
                asSnippet: true,
                description: 'Search only inside repositories that contain matching file paths',
            },
            {
                label: 'has.content(...)',
                insertText: 'has.content(${1:TODO})',
                asSnippet: true,
                description: 'Search only inside repositories that contain matching file contents ',
            },
            {
                label: 'has.file(...)',
                insertText: 'has.file(path:${1:CHANGELOG} content:${2:fix})',
                asSnippet: true,
                description: 'Search only in repositories that contain matching file paths and contents',
            },
            {
                label: 'has.topic(...)',
                insertText: 'has.topic(${1})',
                description: 'Search only inside repositories that have a matching GitHub/GitLab topic',
                asSnippet: true,
            },
            {
                label: 'has.commit.after(...)',
                insertText: 'has.commit.after(${1:1 month ago})',
                asSnippet: true,
                description: 'Search only in repositories that have been committed to since then',
            },
            {
                label: 'has.description(...)',
                insertText: 'has.description(${1})',
                asSnippet: true,
                description: 'Search only inside repositories whose description matches',
            },
            {
                label: 'has.meta(...)',
                insertText: 'has.meta(${1:key}:${2:value})',
                description:
                    'Search only inside repositories having ({key}:{value}) pair, or ({key}) with any value or ({key}:) with no value metadata',
                asSnippet: true,
            },
        ]
    }
    if (field === 'file') {
        return [
            {
                label: 'has.content(...)',
                insertText: 'has.content(${1:TODO})',
                asSnippet: true,
                description: 'Search only inside files whose contents match a pattern',
            },
            {
                label: 'has.owner(...)',
                insertText: 'has.owner(${1})',
                asSnippet: true,
                description: 'Search only inside files that have a specific owner',
            },
            {
                label: 'has.contributor(...)',
                insertText: 'has.contributor(${1})',
                asSnippet: true,
                description: 'Search only inside files that have a contributor that matches a pattern',
            },
        ]
    }
    if (field === 'rev') {
        return [
            {
                label: 'at.time(...)',
                insertText: 'at.time(${1:1 year ago})',
                asSnippet: true,
                description:
                    'Search repos at a specific time in history. Optionally, a base revision can be specified as a second parameter like rev:at.time(yesterday, my-branch)',
            },
        ]
    }
    return []
}
