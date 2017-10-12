
export enum FilterType {
    Repo = 'repo',
    UnknownRepo = 'unknownrepo',
    RepoGroup = 'repogroup',
    File = 'file',
}

const filterTypes = new Set(Object.values(FilterType))

export type Filter = RepoGroupFilter | RepoFilter | FileFilter | UnknownRepoFilter

interface BaseFilter {
    value: string
}

export interface RepoFilter extends BaseFilter {
    type: FilterType.Repo
}

export interface UnknownRepoFilter extends BaseFilter {
    type: FilterType.UnknownRepo
}

export interface FileFilter extends BaseFilter {
    type: FilterType.File
}

export interface RepoGroupFilter extends BaseFilter {
    type: FilterType.RepoGroup
}

export interface SearchOptions {
    query: string
    filters: Filter[]
    matchCase: boolean
    matchWord: boolean
    matchRegex: boolean
}

/**
 * Builds a URL query for given SearchOptions (without leading `?`)
 */
export function buildSearchURLQuery({ query, filters, matchCase, matchRegex, matchWord }: SearchOptions): string {
    const searchParams = new URLSearchParams()
    searchParams.set('q', query)
    for (const filter of filters) {
        searchParams.append(filter.type, filter.value)
    }
    searchParams.set('matchCase', matchCase + '')
    searchParams.set('matchWord', matchWord + '')
    searchParams.set('matchRegex', matchRegex + '')
    return searchParams.toString().replace(/%2F/g, '/')
}

/**
 * Parses the SearchOptions out of URL search params
 */
export function parseSearchURLQuery(query: string): SearchOptions {
    const searchParams = new URLSearchParams(query)
    const filters: Filter[] = Array.from(searchParams.entries())
        .filter(([name]) => filterTypes.has(name))
        .map(([name, value]) => ({ type: name, value } as Filter))
    return {
        filters,
        query: searchParams.get('q') || '',
        matchCase: searchParams.get('matchCase') === 'true',
        matchWord: searchParams.get('matchWord') === 'true',
        matchRegex: searchParams.get('matchRegex') === 'true',
    }
}
