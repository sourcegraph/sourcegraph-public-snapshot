import { useCallback, useState } from 'react'

import { flatten } from 'lodash'

import { createAggregateError, ErrorLike } from '@sourcegraph/common'
import { getDocumentNode } from '@sourcegraph/http-client'

import { getWebGraphQLClient } from '../backend/graphql'
import { CodeIntelSearchVariables } from '../graphql-operations'

import { LanguageSpec } from './language-specs/languagespec'
import { Location, buildSearchBasedLocation } from './location'
import { CODE_INTEL_SEARCH_QUERY } from './ReferencesPanelQueries'
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

async function searchReferences(terms: string[]): Promise<Location[]> {
    const result = await executeSearchQuery(terms)

    return result.flatMap(searchResultToResults).map(buildSearchBasedLocation)
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
    const result = await client.query<Response, CodeIntelSearchVariables>({
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
