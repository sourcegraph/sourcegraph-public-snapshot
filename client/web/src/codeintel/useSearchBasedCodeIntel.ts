import { useCallback, useState } from 'react'

import { flatten } from 'lodash'

import {
    appendLineRangeQueryParameter,
    appendSubtreeQueryParameter,
    createAggregateError,
    ErrorLike,
    toPositionOrRangeQueryParameter,
} from '@sourcegraph/common'
import { getDocumentNode } from '@sourcegraph/http-client'

import { getWebGraphQLClient } from '../backend/graphql'
import { CodeIntelSearchResult, CodeIntelSearchVariables, LocationFields } from '../graphql-operations'

import { LanguageSpec } from './language-specs/languagespec'
import { Location, buildSearchBasedLocation } from './location'
import { CODE_INTEL_SEARCH_QUERY } from './ReferencesPanelQueries'
import { definitionQuery, isSourcegraphDotCom, referencesQuery, searchWithFallback } from './searchBased'
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
    searchToken: string
    fileContent: string

    spec: LanguageSpec

    isFork: boolean
    isArchived: boolean

    getSetting: SettingsGetter
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
    spec,
    getSetting,
}: UseSearchBasedCodeIntelOptions): Promise<Location[]> {
    const queryTerms = referencesQuery({ searchToken, path, fileExts: spec.fileExts })
    const queryArguments = {
        repo,
        isFork,
        isArchived,
        commit,
        queryTerms,
    }

    const doSearch = (negateRepoFilter: boolean): Promise<Location[]> =>
        searchWithFallback(args => searchReferences(args.queryTerms), queryArguments, negateRepoFilter, getSetting)

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
    spec,
    getSetting,
}: UseSearchBasedCodeIntelOptions): Promise<Location[]> {
    const filterDefinitions = (results: Location[]): Location[] =>
        spec?.filterDefinitions
            ? spec.filterDefinitions<Location>(results, {
                  repo,
                  fileContent,
                  filePath: path,
              })
            : results

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
            args => searchAndFilterDefinitions({ filterDefinitions, queryTerms: args.queryTerms }),
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
    filterDefinitions,
    queryTerms,
}: {
    /** The function used to filter definitions. */
    filterDefinitions: (results: Location[]) => Location[]
    /** The terms of the search query. */
    queryTerms: string[]
}): Promise<Location[]> {
    // Perform search and perform pre-filtering before passing it
    // off to the language spec for the proper filtering pass.
    const result = await executeSearchQuery(queryTerms)
    const preFilteredResults = searchResultsToLocations(result).map(buildSearchBasedLocation)

    // TODO: This needs to be ported
    // const preFilteredResults = searchResults.filter(result => !isExternalPrivateSymbol(doc, path, result))

    // Filter results based on language spec
    const filteredResults = filterDefinitions(preFilteredResults)

    return sortByProximity(filteredResults, location.pathname)
}

async function searchReferences(terms: string[]): Promise<Location[]> {
    const result = await executeSearchQuery(terms)

    return searchResultsToLocations(result).map(buildSearchBasedLocation)
}

async function executeSearchQuery(terms: string[]): Promise<CodeIntelSearchResult> {
    const client = await getWebGraphQLClient()
    const result = await client.query<CodeIntelSearchResult, CodeIntelSearchVariables>({
        query: getDocumentNode(CODE_INTEL_SEARCH_QUERY),
        variables: {
            query: terms.join(' '),
        },
    })

    if (result.error) {
        throw createAggregateError([result.error])
    }

    return result.data
}

export function searchResultsToLocations(result: CodeIntelSearchResult): LocationFields[] {
    if (!result || !result.search) {
        return []
    }

    const searchResults = result.search.results.results
        .filter(value => value !== undefined)
        .filter(result => result.__typename === 'FileMatch')

    const newReferences: LocationFields[] = []
    for (const result of searchResults) {
        if (result.__typename !== 'FileMatch') {
            continue
        }

        const resource = {
            path: result.file.path,
            content: result.file.content,
            repository: result.repository,
            commit: {
                oid: result.file.commit.oid,
            },
        }

        for (const lineMatch of result.lineMatches) {
            const positionOrRangeQueryParameter = toPositionOrRangeQueryParameter({
                // TODO: only using first offset?
                position: { line: lineMatch.lineNumber + 1, character: lineMatch.offsetAndLengths[0][0] + 1 },
            })
            const url = appendLineRangeQueryParameter(
                appendSubtreeQueryParameter(result.file.url),
                positionOrRangeQueryParameter
            )
            newReferences.push({
                url,
                resource,
                range: {
                    start: {
                        line: lineMatch.lineNumber,
                        character: lineMatch.offsetAndLengths[0][0],
                    },
                    end: {
                        line: lineMatch.lineNumber,
                        character: lineMatch.offsetAndLengths[0][0] + lineMatch.offsetAndLengths[0][1],
                    },
                },
            })
        }

        const symbolReferences = result.symbols.map(symbol => ({
            url: symbol.location.url,
            resource,
            range: symbol.location.range,
        }))
        for (const symbolReference of symbolReferences) {
            newReferences.push(symbolReference)
        }
    }

    return newReferences
}
