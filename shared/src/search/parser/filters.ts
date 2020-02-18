import { Filter } from './parser'
import { SearchSuggestion } from '../../graphql/schema'
import {
    FilterTypes,
    isNegatedFilter,
    resolveNegatedFilter,
    NegatableFilter,
    isNegatableFilter,
    isFilterType,
} from '../interactive/util'
import { Omit } from 'utility-types'

interface BaseFilterDefinition {
    alias?: string
    description: string
    discreteValues?: string[]
    suggestions?: SearchSuggestion['__typename'] | string[]
    default?: string
}

interface NegatableFilterDefinition extends Omit<BaseFilterDefinition, 'description'> {
    negatable: true
    description: (negated: boolean) => string
}

export type FilterDefinition = BaseFilterDefinition | NegatableFilterDefinition

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

export const FILTERS: Record<NegatableFilter, NegatableFilterDefinition> &
    Record<Exclude<FilterTypes, NegatableFilter>, BaseFilterDefinition> = {
    [FilterTypes.after]: {
        description: 'Commits made after a certain date',
    },
    [FilterTypes.archived]: {
        description: 'Include results from archived repositories.',
    },
    [FilterTypes.author]: {
        description: 'The author of a commit',
    },
    [FilterTypes.before]: {
        description: 'Commits made before a certain date',
    },
    [FilterTypes.case]: {
        description: 'Treat the search pattern as case-sensitive.',
        discreteValues: ['yes', 'no'],
        default: 'no',
    },
    [FilterTypes.content]: {
        description:
            'Explicitly overrides the search pattern. Used for explicitly delineating the search pattern to search for in case of clashes.',
    },
    [FilterTypes.count]: {
        description: 'Number of results to fetch (integer)',
    },
    [FilterTypes.file]: {
        alias: 'f',
        negatable: true,
        description: negated =>
            `${negated ? 'Exclude' : 'Include only'} results from files matching the given regex pattern.`,
        suggestions: 'File',
    },
    [FilterTypes.fork]: {
        discreteValues: ['yes', 'no', 'only'],
        description: 'Include results from forked repositories.',
    },
    [FilterTypes.lang]: {
        negatable: true,
        description: negated => `${negated ? 'Exclude' : 'Include only'} results from the given language`,
        suggestions: LANGUAGES,
    },
    [FilterTypes.message]: {
        description: 'Commits with messages matching a certain string',
    },
    [FilterTypes.patterntype]: {
        discreteValues: ['regexp', 'literal', 'structural'],
        description: 'The pattern type (regexp, literal, structural) in use',
    },
    [FilterTypes.repo]: {
        alias: 'r',
        negatable: true,
        description: negated =>
            `${negated ? 'Exclude' : 'Include only'} results from repositories matching the given regex pattern.`,
        suggestions: 'Repository',
    },
    [FilterTypes.repogroup]: {
        description: 'group-name (include results from the named group)',
    },
    [FilterTypes.repohascommitafter]: {
        description: '"string specifying time frame" (filter out stale repositories without recent commits)',
    },
    [FilterTypes.repohasfile]: {
        negatable: true,
        description: negated =>
            `${negated ? 'Exclude' : 'Include only'} results from repos that contain a matching file`,
    },
    [FilterTypes.timeout]: {
        description: 'Duration before timeout',
    },
    [FilterTypes.type]: {
        description: 'Limit results to the specified type.',
        discreteValues: ['code', 'diff', 'commit', 'symbol', 'repo', 'path'],
    },
}

/**
 * Returns the {@link FilterDefinition} for the given filterType if it exists, or `undefined` otherwise.
 */
export const resolveFilter = (
    filterType: string
):
    | { type: NegatableFilter; negated: boolean; definition: NegatableFilterDefinition }
    | { type: Exclude<FilterTypes, NegatableFilter>; definition: BaseFilterDefinition }
    | undefined => {
    filterType = filterType.toLowerCase()
    if (isNegatedFilter(filterType)) {
        const type = resolveNegatedFilter(filterType)
        return {
            type,
            definition: FILTERS[type],
            negated: true,
        }
    }
    if (isFilterType(filterType)) {
        if (isNegatableFilter(filterType)) {
            return {
                type: filterType,
                definition: FILTERS[filterType],
                negated: false,
            }
        }
        if (FILTERS[filterType]) {
            return { type: filterType, definition: FILTERS[filterType] }
        }
    }
    for (const [type, definition] of Object.entries(FILTERS as Record<FilterTypes, FilterDefinition>)) {
        if (definition.alias && filterType === definition.alias) {
            return {
                type: type as Exclude<FilterTypes, NegatableFilter>,
                definition: definition as BaseFilterDefinition,
            }
        }
    }
    return undefined
}

/**
 * Validates a filter given its type and value.
 */
export const validateFilter = (
    filterType: string,
    filterValue: Filter['filterValue']
): { valid: true } | { valid: false; reason: string } => {
    const typeAndDefinition = resolveFilter(filterType)
    if (!typeAndDefinition) {
        return { valid: false, reason: 'Invalid filter type.' }
    }
    const { definition } = typeAndDefinition
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
