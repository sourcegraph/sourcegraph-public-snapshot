/* eslint-disable no-template-curly-in-string */
import { Completion, resolveFieldAlias } from './filters'

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
                    { name: 'content' },
                    {
                        name: 'commit',
                        fields: [{ name: 'after' }],
                    },
                ],
            },
            {
                name: 'dependencies',
            },
            {
                name: 'deps',
            },

            {
                name: 'dependents',
            },
            {
                name: 'revdeps',
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
                label: 'contains.file(...)',
                insertText: 'contains.file(${1:CHANGELOG})',
                asSnippet: true,
            },
            {
                label: 'contains.content(...)',
                insertText: 'contains.content(${1:TODO})',
                asSnippet: true,
            },
            {
                label: 'contains(...)',
                insertText: 'contains(file:${1:CHANGELOG} content:${2:fix})',
                asSnippet: true,
            },
            {
                label: 'contains.commit.after(...)',
                insertText: 'contains.commit.after(${1:1 month ago})',
                asSnippet: true,
            },
            {
                label: 'deps(...)',
                insertText: 'deps(${1})',
                asSnippet: true,
            },
            {
                label: 'dependencies(...)',
                insertText: 'dependencies(${1})',
                asSnippet: true,
            },

            {
                label: 'revdeps(...)',
                insertText: 'revdeps(${1})',
                asSnippet: true,
            },
            {
                label: 'dependents(...)',
                insertText: 'dependents(${1})',
                asSnippet: true,
            },
        ]
    }
    return []
}
