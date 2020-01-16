import { SearchSuggestion } from '../../../../shared/src/graphql/schema'
import { Filter } from './parser'

export interface FilterDefinition {
    aliases: string[]
    description: string
    discreteValues?: string[]
    suggestions?: SearchSuggestion['__typename'] | string[]
    default?: string
}

const LANGUAGES: string[] = [
    'c',
    'cpp',
    'csharp',
    'css',
    'go',
    'graphql',
    'haskell',
    'html',
    'java',
    'javascript',
    'json',
    'lua',
    'markdown',
    'php',
    'powershell',
    'python',
    'r',
    'ruby',
    'sass',
    'swift',
    'typescript',
]

export const FILTERS: readonly FilterDefinition[] = [
    {
        aliases: ['r', 'repo'],
        description: 'Include only results from repositories matching the given regex pattern.',
        suggestions: 'Repository',
    },
    {
        aliases: ['-r', '-repo'],
        description: 'Exclude results from repositories matching the given regex pattern.',
        suggestions: 'Repository',
    },
    {
        aliases: ['f', 'file'],
        description: 'Include only results from files matching the given regex pattern.',
        suggestions: 'File',
    },
    {
        aliases: ['-f', '-file'],
        description: 'Exclude results from files matching the given regex pattern.',
        suggestions: 'File',
    },
    {
        aliases: ['repogroup'],
        description: 'group-name (include results from the named group)',
    },
    {
        aliases: ['repohasfile'],
        description: 'regex-pattern (include results from repos that contain a matching file)',
        suggestions: 'File',
    },
    {
        aliases: ['-repohasfile'],
        description: 'regex-pattern (exclude results from repositories that contain a matching file)',
        suggestions: 'File',
    },
    {
        aliases: ['repohascommitafter'],
        description: '"string specifying time frame" (filter out stale repositories without recent commits)',
    },
    {
        aliases: ['type'],
        description: 'Limit results to the specified type.',
        discreteValues: ['code', 'diff', 'commit', 'symbol'],
    },
    {
        aliases: ['case'],
        description: 'Treat the search pattern as case-sensitive.',
        discreteValues: ['yes', 'no'],
        default: 'no',
    },
    {
        aliases: ['lang'],
        description: 'Include only results from the given language',
        discreteValues: LANGUAGES,
    },
    {
        aliases: ['-lang'],
        description: 'Exclude results from the given language',
        discreteValues: LANGUAGES,
    },
    {
        aliases: ['fork'],
        discreteValues: ['yes', 'no', 'only'],
        description: 'Fork',
    },
    {
        aliases: ['archived'],
        description: 'Archived',
    },
    {
        aliases: ['count'],
        description: 'Number of results to fetch (integer)',
    },
    {
        aliases: ['timeout'],
        description: 'Duration before timeout',
    },
]

/**
 * Returns the {@link FilterDefinition} for the given filterType if it exists, or `undefined` otherwise.
 */
export const getFilterDefinition = (filterType: string): FilterDefinition | undefined =>
    FILTERS.find(({ aliases }) => aliases.some(a => a === filterType))

/**
 * Validates a filter given its type and value.
 */
export const validateFilter = (
    filterType: string,
    filterValue: Filter['filterValue']
): { valid: true } | { valid: false; reason: string } => {
    const definition = getFilterDefinition(filterType)
    if (!definition) {
        return { valid: false, reason: 'Invalid filter type.' }
    }
    if (
        definition.discreteValues &&
        (!filterValue ||
            filterValue.token.type !== 'literal' ||
            !definition.discreteValues.includes(filterValue.token.value))
    ) {
        return {
            valid: false,
            reason: `Invalid filter value, expected one of: ${definition.discreteValues.join(', ')}.`,
        }
    }
    return { valid: true }
}
