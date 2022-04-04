import { useCallback, useState } from 'react'

import { flatten } from 'lodash'

import { createAggregateError, ErrorLike } from '@sourcegraph/common'
import { getDocumentNode } from '@sourcegraph/http-client'

import { getWebGraphQLClient } from '../backend/graphql'
import { CodeIntelSearchResult, CodeIntelSearchVariables } from '../graphql-operations'

import { LanguageSpec } from './language-specs/languagespec'
import { Location, buildSearchBasedLocation } from './location'
import { CODE_INTEL_SEARCH_QUERY } from './ReferencesPanelQueries'
import { isSourcegraphDotCom, referencesQuery, searchWithFallback } from './searchBased'
import { SettingsGetter } from './settings'
import { sortByProximity } from './sort'
import { searchResultsToLocations } from './useCodeIntel'

type LocationHandler = (locations: Location[]) => void
interface UseSearchBasedCodeIntelResult {
    fetch: (onReferences: LocationHandler) => void
    loading: boolean
    error?: ErrorLike
}

interface UseSearchBasedCodeIntelOptions {
    repo: string
    commit: string

    path: string
    searchToken: string

    spec: LanguageSpec

    isFork: boolean
    isArchived: boolean

    getSetting: SettingsGetter
}

export const useSearchBasedCodeIntel = (options: UseSearchBasedCodeIntelOptions): UseSearchBasedCodeIntelResult => {
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState()

    const fetch = useCallback(
        (onReferences: LocationHandler) => {
            setLoading(true)
            searchBasedReferences(options)
                .then(references => {
                    onReferences(references)
                    setLoading(false)
                })
                .catch(error => {
                    setError(error)
                    setLoading(false)
                })
        },
        [options]
    )

    return { fetch, loading, error }
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

async function searchReferences(terms: string[]): Promise<Location[]> {
    const query = terms.join(' ')
    const result = await executeQuery(query)

    return searchResultsToLocations(result).map(buildSearchBasedLocation)
}

async function executeQuery(query: string): Promise<CodeIntelSearchResult> {
    const client = await getWebGraphQLClient()
    const result = await client.query<CodeIntelSearchResult, CodeIntelSearchVariables>({
        query: getDocumentNode(CODE_INTEL_SEARCH_QUERY),
        variables: {
            query,
        },
    })

    if (result.error) {
        throw createAggregateError([result.error])
    }

    return result.data
}
