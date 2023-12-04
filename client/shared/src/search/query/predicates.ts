/* eslint-disable no-template-curly-in-string */
import { type Completion, resolveFieldAlias } from './filters'

interface Access {
    name: string
    fields?: Access[]
}

/**
 * Represents recognized predicate accesses associated with fields. The
 * data structure is a tree, where nodes are lists to preserve ordering
 * for autocomplete suggestions.
 */
export const PREDICATES: Access[] = [
    {
        name: 'repo',
        fields: [
            {
                name: 'contains',
                fields: [
                    { name: 'file' },
                    { name: 'path' },
                    { name: 'content' },
                    {
                        name: 'commit',
                        fields: [{ name: 'after' }],
                    },
                ],
            },
            {
                name: 'has',
                fields: [
                    { name: 'file' },
                    { name: 'path' },
                    { name: 'content' },
                    {
                        name: 'commit',
                        fields: [{ name: 'after' }],
                    },
                    { name: 'description' },
                    { name: 'tag' },
                    { name: 'key' },
                    { name: 'meta' },
                    { name: 'topic' },
                ],
            },
        ],
    },
    {
        name: 'file',
        fields: [
            {
                name: 'contains',
                fields: [{ name: 'content' }],
            },
            {
                name: 'has',
                fields: [{ name: 'content' }, { name: 'owner' }],
            },
        ],
    },
]

/** Represents a predicate's components corresponding to the syntax path(parameters). */
export interface Predicate {
    path: string[]
    parameters: string
}

/** Returns the access tree for a predicate path. */
export const resolveAccess = (path: string[], tree: Access[]): Access[] | undefined => {
    if (path.length === 0) {
        return tree
    }

    // repo:contains() and file:contains() are not supported
    if (path.length === 1 && path[0] === 'contains') {
        return undefined
    }

    const subtree = tree.find(value => value.name === path[0])
    if (!subtree) {
        return undefined
    }
    if (!subtree.fields) {
        return []
    }
    return resolveAccess(path.slice(1), subtree.fields)
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
export const scanPredicate = (field: string, value: string): Predicate | undefined => {
    const match = value.match(/^[.a-z]+/i)
    if (!match) {
        return undefined
    }
    const name = match[0]
    const path = name.split('.')
    // Remove negation from the field for lookup
    if (field.startsWith('-')) {
        field = field.slice(1)
    }
    field = resolveFieldAlias(field)
    const access = resolveAccess([field, ...path], PREDICATES)
    if (!access) {
        return undefined
    }
    const rest = value.slice(name.length)
    const parameters = scanBalancedParens(rest)
    if (!parameters) {
        return undefined
    }
    if (!parameters.startsWith('(') || !parameters.endsWith(')')) {
        return undefined
    }

    return { path, parameters }
}

export const predicateCompletion = (field: string): Completion[] => {
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
    return []
}
