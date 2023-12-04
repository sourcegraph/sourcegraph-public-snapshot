import type { Literal } from './token'

interface Access {
    name: string
    fields?: Access[]
}

export const SELECTORS: Access[] = [
    {
        name: 'repo',
    },
    {
        name: 'file',
        fields: [{ name: 'directory' }, { name: 'path' }, { name: 'owners' }],
    },
    {
        name: 'content',
    },
    {
        name: 'symbol',
        fields: [
            { name: 'file' },
            { name: 'module' },
            { name: 'namespace' },
            { name: 'package' },
            { name: 'class' },
            { name: 'method' },
            { name: 'property' },
            { name: 'field' },
            { name: 'constructor' },
            { name: 'enum' },
            { name: 'interface' },
            { name: 'function' },
            { name: 'variable' },
            { name: 'constant' },
            { name: 'string' },
            { name: 'number' },
            { name: 'boolean' },
            { name: 'array' },
            { name: 'object' },
            { name: 'key' },
            { name: 'null' },
            { name: 'enum-member' },
            { name: 'struct' },
            { name: 'event' },
            { name: 'operator' },
            { name: 'type-parameter' },
        ],
    },
    {
        name: 'commit',
        fields: [{ name: 'diff', fields: [{ name: 'added' }, { name: 'removed' }] }],
    },
]
const kinds = new Set(SELECTORS.map(value => value.name))

/**
 * Returns true if the provided select value has additional subfields.
 */
export const selectorHasFields = (value: string): boolean => {
    const parts = value.split('.')
    let fields: Access[] | undefined = SELECTORS

    for (const part of parts) {
        fields = fields?.find(value => value.name === part && value.fields)?.fields
    }
    return !!fields
}

/**
 * Returns all paths rooted at a {@link selector} up to {@param depth}.
 */
export const selectDiscreteValues = (selectors: Access[], depth: number): string[] => {
    if (depth < 0) {
        return []
    }
    const paths: string[] = []
    for (const entry of selectors) {
        paths.push(`${entry.name}`)
        if (entry.fields) {
            paths.push(...selectDiscreteValues(entry.fields, depth - 1).map(value => `${entry.name}.` + value))
        }
    }
    return paths
}

export const selectorCompletion = (value: Literal | undefined): string[] => {
    if (!value) {
        return selectDiscreteValues(SELECTORS, 0)
    }

    if (value.value.endsWith('.') || value.value.split('.').length > 1 || kinds.has(value.value)) {
        // Resolve completions to greater depth for `foo.` if the value is `foo.` or `foo.bar`.
        const kind = value.value.split('.')[0]
        return selectDiscreteValues(
            SELECTORS.filter(value => value.name === kind),
            2
        )
    }
    return selectDiscreteValues(SELECTORS, 0)
}
