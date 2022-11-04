import { useCallback, useState } from 'react'

import stringify from 'fast-json-stable-stringify'
import { flatten, sortBy } from 'lodash'
import LRU from 'lru-cache'

import { createAggregateError, ErrorLike } from '@sourcegraph/common'
import { Range as ExtensionRange, Position as ExtensionPosition } from '@sourcegraph/extension-api-types'
import { getDocumentNode } from '@sourcegraph/http-client'
import { LanguageSpec } from '@sourcegraph/shared/src/codeintel/legacy-extensions/language-specs/language-spec'
import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'

import { getWebGraphQLClient } from '../backend/graphql'
import { CodeIntelSearch2Variables } from '../graphql-operations'

import { Location, buildSearchBasedLocation, split } from './location'
import { CODE_INTEL_SEARCH_QUERY, LOCAL_CODE_INTEL_QUERY } from './ReferencesPanelQueries'
import {
    definitionQuery,
    isExternalPrivateSymbol,
    isSourcegraphDotCom,
    referencesQuery,
    SearchResult,
    searchResultToResults,
    searchWithFallback,
} from './searchBased'
import { SettingsGetter } from './settings'
import { sortByProximity } from './sort'
import { isDefined } from './util/helpers'

type LocationHandler = (locations: Location[]) => void

interface UseSearchBasedCodeIntelResult {
    fetch: (onReferences: LocationHandler, onDefinitions: LocationHandler) => void
    fetchReferences: (onReferences: LocationHandler) => void
    loading: boolean
    error?: ErrorLike
}

interface UseSearchBasedCodeIntelOptions {
    repo: string
    commit: string

    path: string
    position: ExtensionPosition
    searchToken: string
    fileContent: string

    spec: LanguageSpec

    isFork: boolean
    isArchived: boolean

    getSetting: SettingsGetter

    filter?: string
}

export const useSearchBasedCodeIntel = (options: UseSearchBasedCodeIntelOptions): UseSearchBasedCodeIntelResult => {
    const [loadingReferences, setLoadingReferences] = useState(false)
    const [referencesError, setReferencesError] = useState<ErrorLike | undefined>()

    const [loadingDefinitions, setLoadingDefinitions] = useState(false)
    const [definitionsError, setDefinitionsError] = useState<ErrorLike | undefined>()

    const fetchReferences = useCallback(
        (onReferences: LocationHandler) => {
            setLoadingReferences(true)

            searchBasedReferences(options)
                .then(references => {
                    onReferences(references)
                    setLoadingReferences(false)
                })
                .catch(error => {
                    setReferencesError(error)
                    setLoadingReferences(false)
                })
        },
        [options]
    )

    const fetch = useCallback(
        (onReferences: LocationHandler, onDefinitions: LocationHandler) => {
            fetchReferences(onReferences)

            setLoadingDefinitions(true)
            searchBasedDefinitions(options)
                .then(definitions => {
                    onDefinitions(definitions)
                    setLoadingDefinitions(false)
                })
                .catch(error => {
                    setDefinitionsError(error)
                    setLoadingDefinitions(false)
                })
        },
        [options, fetchReferences]
    )

    const errors = [definitionsError, referencesError].filter(isDefined)
    return {
        fetch,
        fetchReferences,
        loading: loadingReferences || loadingDefinitions,
        error: createAggregateError(errors),
    }
}

// searchBasedReferences is 90% copy&paste from code-intel-extension's
export async function searchBasedReferences({
    repo,
    isFork,
    isArchived,
    commit,
    searchToken,
    path,
    position,
    fileContent,
    spec,
    getSetting,
    filter,
}: UseSearchBasedCodeIntelOptions): Promise<Location[]> {
    const filterReferences = (results: Location[]): Location[] =>
        filter ? results.filter(location => location.file.includes(filter)) : results

    const symbol = await findSymbol({ repository: repo, commit, path, row: position.line, column: position.character })
    if (symbol?.refs) {
        return symbol.refs.map(reference => ({
            repo,
            file: path,
            content: fileContent,
            commitID: commit,
            range: rangeToExtensionRange(reference),
            url: toPrettyBlobURL({
                filePath: path,
                revision: commit,
                repoName: repo,
                commitID: commit,
                position: {
                    line: reference.row + 1,
                    character: reference.column + 1,
                },
            }),
            lines: split(fileContent),
            precise: false,
        }))
    }

    const queryTerms = referencesQuery({ searchToken, path, fileExts: spec.fileExts })
    const queryArguments = {
        repo,
        isFork,
        isArchived,
        commit,
        queryTerms,
        filterReferences,
    }

    const doSearch = (negateRepoFilter: boolean): Promise<Location[]> =>
        searchWithFallback(
            args => searchAndFilterReferences({ queryTerms: args.queryTerms, filterReferences }),
            queryArguments,
            negateRepoFilter,
            getSetting
        )

    // Perform a search in the current git tree
    const sameRepoReferences = doSearch(false)

    // Perform an indexed search over all _other_ repositories. This
    // query is ineffective on DotCom as we do not keep repositories
    // in the index permanently.
    const remoteRepoReferences = isSourcegraphDotCom() ? Promise.resolve([]) : doSearch(true)

    // Resolve then merge all references and sort them by proximity
    // to the current text document path.
    const referenceChunk = [sameRepoReferences, remoteRepoReferences]
    const mergedReferences = flatten(await Promise.all(referenceChunk))
    return sortByProximity(mergedReferences, location.pathname)
}

export async function searchBasedDefinitions({
    repo,
    isFork,
    isArchived,
    commit,
    searchToken,
    fileContent,
    path,
    position,
    spec,
    getSetting,
    filter,
}: UseSearchBasedCodeIntelOptions): Promise<Location[]> {
    const symbol = await findSymbol({ repository: repo, commit, path, row: position.line, column: position.character })
    if (symbol?.def) {
        return [
            {
                repo,
                file: path,
                content: fileContent,
                commitID: commit,
                range: rangeToExtensionRange(symbol.def),
                url: toPrettyBlobURL({
                    filePath: path,
                    revision: commit,
                    repoName: repo,
                    commitID: commit,
                    position: {
                        line: symbol.def.row + 1,
                        character: symbol.def.column + 1,
                    },
                }),
                lines: split(fileContent),
                precise: false,
            },
        ]
    }

    const filterDefinitions = (results: Location[]): Location[] => {
        const filteredByName = filter ? results.filter(location => location.file.includes(filter)) : results
        return spec?.filterDefinitions
            ? spec.filterDefinitions<Location>(filteredByName, {
                  repo,
                  fileContent,
                  filePath: path,
              })
            : filteredByName
    }

    // Construct base definition query without scoping terms
    const queryTerms = definitionQuery({ searchToken, path, fileExts: spec.fileExts })
    const queryArguments = {
        repo,
        isFork,
        isArchived,
        commit,
        path,
        fileContent,
        filterDefinitions,
        queryTerms,
    }

    const doSearch = (negateRepoFilter: boolean): Promise<Location[]> =>
        searchWithFallback(
            args => searchAndFilterDefinitions({ spec, path, filterDefinitions, queryTerms: args.queryTerms }),
            queryArguments,
            negateRepoFilter,
            getSetting
        )

    // Perform a search in the current git tree
    const sameRepoDefinitions = doSearch(false)

    // Return any local location definitions first
    const results = await sameRepoDefinitions
    if (results.length > 0) {
        return results
    }

    // Fallback to definitions found in any other repository. This performs
    // an indexed search over all repositories. Do not do this on the DotCom
    // instance as we are unlikely to have indexed the relevant definition
    // and we'd end up jumping to what would seem like a random line of code.
    return isSourcegraphDotCom() ? Promise.resolve([]) : doSearch(true)
}

/**
 * Perform a search query for definitions. Returns results converted to locations,
 * filtered by the language's definition filter, and sorted by proximity to the
 * current text document path.
 *
 * @param api The GraphQL API instance.
 * @param args Parameter bag.
 */
async function searchAndFilterDefinitions({
    spec,
    path,
    filterDefinitions,
    queryTerms,
}: {
    /** The LanguageSpec of the language in which we're searching */
    spec: LanguageSpec
    /** The file we're in */
    path: string
    /** The function used to filter definitions. */
    filterDefinitions: (results: Location[]) => Location[]
    /** The terms of the search query. */
    queryTerms: string[]
}): Promise<Location[]> {
    // Perform search and perform pre-filtering before passing it
    // off to the language spec for the proper filtering pass.
    const result = await executeSearchQuery(queryTerms)
    const preFilteredResults = result
        .flatMap(searchResultToResults)
        .filter(result => !isExternalPrivateSymbol(spec, path, result))
        .map(buildSearchBasedLocation)

    // Filter results based on language spec
    const filteredResults = filterDefinitions(preFilteredResults)

    return sortByProximity(filteredResults, location.pathname)
}

async function searchAndFilterReferences({
    queryTerms,
    filterReferences,
}: {
    /** The terms of the search query. */
    queryTerms: string[]
    /** The function used to filter definitions. */
    filterReferences: (results: Location[]) => Location[]
}): Promise<Location[]> {
    const result = await executeSearchQuery(queryTerms)
    const references = result.flatMap(searchResultToResults).map(buildSearchBasedLocation)
    return filterReferences ? filterReferences(references) : references
}

async function executeSearchQuery(terms: string[]): Promise<SearchResult[]> {
    interface Response {
        search: {
            results: {
                limitHit: boolean
                results: (SearchResult | undefined)[]
            }
        }
    }
    const client = await getWebGraphQLClient()
    const result = await client.query<Response, CodeIntelSearch2Variables>({
        query: getDocumentNode(CODE_INTEL_SEARCH_QUERY),
        variables: {
            query: terms.join(' '),
        },
    })

    if (result.error) {
        throw createAggregateError([result.error])
    }

    return result.data.search.results.results.filter(isDefined)
}

const findSymbol = async (
    repositoryCommitPathPosition: RepositoryCommitPathPosition
): Promise<LocalSymbol | undefined> => {
    const payload = await fetchLocalCodeIntelPayload(repositoryCommitPathPosition)
    if (!payload) {
        return
    }

    for (const symbol of payload.symbols) {
        if (isInRange(repositoryCommitPathPosition, symbol.def)) {
            return symbol
        }

        for (const reference of symbol.refs ?? []) {
            if (isInRange(repositoryCommitPathPosition, reference)) {
                return symbol
            }
        }
    }

    return undefined
}

const cache = <Arguments extends unknown[], V>(
    func: (...args: Arguments) => V,
    options: LRU.Options<string, V>
): ((...args: Arguments) => V) => {
    const lru = new LRU<string, V>(options)
    return (...args) => {
        const key = stringify(args)
        if (lru.has(key)) {
            // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
            return lru.get(key)!
        }
        const value = func(...args)
        lru.set(key, value)
        return value
    }
}

const fetchLocalCodeIntelPayload = cache(
    async (repositoryCommitPath: RepositoryCommitPath): Promise<LocalCodeIntelPayload | undefined> => {
        const client = await getWebGraphQLClient()
        type LocalCodeIntelResponse = GenericBlobResponse<{ localCodeIntel: string }>
        const result = await client.query<LocalCodeIntelResponse, RepositoryCommitPath>({
            query: getDocumentNode(LOCAL_CODE_INTEL_QUERY),
            variables: repositoryCommitPath,
        })

        if (result.error) {
            throw createAggregateError([result.error])
        }

        const payloadString = result.data.repository?.commit?.blob?.localCodeIntel
        if (!payloadString) {
            return undefined
        }

        const payload = JSON.parse(payloadString) as LocalCodeIntelPayload
        if (!payload) {
            return undefined
        }

        for (const symbol of payload.symbols) {
            if (symbol.refs) {
                symbol.refs = sortBy(symbol.refs, reference => reference.row)
            }
        }

        return payload
    },
    { max: 10 }
)

interface RepositoryCommitPath {
    repository: string
    commit: string
    path: string
}

type RepositoryCommitPathPosition = RepositoryCommitPath & Position

type LocalCodeIntelPayload = {
    symbols: LocalSymbol[]
} | null

interface LocalSymbol {
    hover?: string
    def: LocalRange
    refs?: LocalRange[]
}

interface LocalRange {
    row: number
    column: number
    length: number
}

interface Position {
    row: number
    column: number
}

const isInRange = (position: Position, range: LocalRange): boolean => {
    if (position.row !== range.row) {
        return false
    }
    if (position.column < range.column) {
        return false
    }
    if (position.column > range.column + range.length) {
        return false
    }
    return true
}

/** The response envelope for all blob queries. */
interface GenericBlobResponse<R> {
    repository: { commit: { blob: R | null } | null } | null
}

const rangeToExtensionRange = (range: LocalRange): ExtensionRange => ({
    start: {
        line: range.row,
        character: range.column,
    },
    end: {
        line: range.row,
        character: range.column + range.length,
    },
})
