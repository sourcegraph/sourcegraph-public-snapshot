import * as yaml from 'js-yaml/dist/js-yaml'

export enum FilterType {
    Repo = 'repo',
    RepoGroup = 'repogroup',
    File = 'file',
    FileGlob = 'fileglob'
}

export type Filter = RepoGroupFilter | RepoFilter | FileFilter | FileGlobFilter

export interface RepoFilter {
    type: FilterType.Repo
    repoPath: string
}

export interface FileFilter {
    type: FilterType.File
    filePath: string
}

export interface FileGlobFilter {
    type: FilterType.FileGlob
    glob: string
}

export interface RepoGroupFilter {
    type: FilterType.RepoGroup
    name: string
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
    const filterString = yaml.safeDump(filters, {
        flowLevel: 0,
        indent: 0,
        lineWidth: Infinity,
        noCompatMode: true,
        sortKeys: true,
        noRefs: true,
        condenseFlow: true
    }).trim()
    searchParams.set('filters', filterString)
    searchParams.set('matchCase', matchCase + '')
    searchParams.set('matchWord', matchWord + '')
    searchParams.set('matchCase', matchRegex + '')
    // Unescape some characters that are technically reserved but no modern browser has issues with
    // Escape spaces to + instead of %20
    const querystring = searchParams.toString()
        .replace(/%20/g, '+')
        .replace(/%5B/g, '[')
        .replace(/%5D/g, ']')
        .replace(/%7B/g, '{')
        .replace(/%7D/g, '}')
        .replace(/%2C/g, ',')
        .replace(/%3A/g, ':')
        .replace(/%2F/g, '/')
    return querystring
}

/**
 * Parses the SearchOptions out of URL search params
 */
export function parseSearchURLQuery(query: string): SearchOptions {
    const searchParams = new URLSearchParams(query)
    return {
        query: searchParams.get('q') || '',
        filters: yaml.safeLoad(searchParams.get('filters') || '[]'),
        matchCase: searchParams.get('matchCase') === 'true',
        matchWord: searchParams.get('matchWord') === 'true',
        matchRegex: searchParams.get('matchCase') === 'true'
    }
}
