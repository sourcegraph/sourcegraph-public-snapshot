import { useCallback, useEffect, useRef, useState } from 'react'

import { ApolloError, QueryResult } from '@apollo/client'

import {
    appendLineRangeQueryParameter,
    appendSubtreeQueryParameter,
    toPositionOrRangeQueryParameter,
} from '@sourcegraph/common'
import { dataOrThrowErrors, useLazyQuery, useQuery } from '@sourcegraph/http-client'

import { ConnectionQueryArguments } from '../components/FilteredConnection'
import { asGraphQLResult } from '../components/FilteredConnection/utils'
import {
    UsePreciseCodeIntelForPositionVariables,
    UsePreciseCodeIntelForPositionResult,
    LoadAdditionalReferencesResult,
    LoadAdditionalReferencesVariables,
    LoadAdditionalImplementationsResult,
    LoadAdditionalImplementationsVariables,
    CodeIntelSearchResult,
    CodeIntelSearchVariables,
    LocationFields,
} from '../graphql-operations'

import { LanguageSpec } from './language-specs/languagespec'
import { Location, buildPreciseLocation, buildSearchBasedLocation } from './location'
import {
    LOAD_ADDITIONAL_IMPLEMENTATIONS_QUERY,
    LOAD_ADDITIONAL_REFERENCES_QUERY,
    CODE_INTEL_SEARCH_QUERY,
    USE_PRECISE_CODE_INTEL_FOR_POSITION_QUERY,
} from './ReferencesPanelQueries'
import { definitionQuery, referencesQuery } from './searchBased'
import { sortByProximity } from './sort'

interface CodeIntelData {
    references: {
        endCursor: string | null
        nodes: Location[]
    }
    implementations: {
        endCursor: string | null
        nodes: Location[]
    }
    definitions: {
        endCursor: string | null
        nodes: Location[]
    }
}

export interface UseCodeIntelResult {
    data?: CodeIntelData
    error?: ApolloError
    loading: boolean

    referencesHasNextPage: boolean
    fetchMoreReferences: () => void
    fetchMoreReferencesLoading: boolean

    implementationsHasNextPage: boolean
    fetchMoreImplementations: () => void
    fetchMoreImplementationsLoading: boolean
}

interface UseCodeIntelParameters {
    variables: UsePreciseCodeIntelForPositionVariables & ConnectionQueryArguments
    searchToken: string
    spec: LanguageSpec
    fileContent: string
}

export const useCodeIntel = ({
    variables,
    searchToken,
    spec,
    fileContent,
}: UseCodeIntelParameters): UseCodeIntelResult => {
    const [codeIntelData, setCodeIntelData] = useState<CodeIntelData>()
    const filterDefinitions = useCallback(
        (results: Location[]) =>
            spec?.filterDefinitions
                ? spec.filterDefinitions<Location>(results, {
                      repo: variables.repository,
                      fileContent,
                      filePath: variables.path,
                  })
                : results,
        [spec, fileContent, variables.path, variables.repository]
    )

    const fellBackToSearchBased = useRef(false)
    const shouldFetchPrecise = useRef(true)
    useEffect(() => {
        // We need to fetch again if the variables change
        shouldFetchPrecise.current = true
    }, [
        variables.repository,
        variables.commit,
        variables.path,
        variables.line,
        variables.character,
        variables.filter,
        variables.firstReferences,
        variables.firstImplementations,
    ])

    const [fetchSearchBasedReferences, fetchSearchBasedReferencesResult] = useLazyQuery<
        CodeIntelSearchResult,
        CodeIntelSearchVariables
    >(CODE_INTEL_SEARCH_QUERY, {
        fetchPolicy: 'no-cache',
        onCompleted: result => {
            const newReferences = searchResultsToLocations(result).map(buildSearchBasedLocation)
            const sorted = sortByProximity(newReferences, location.pathname)

            const previousData = codeIntelData
            if (!previousData) {
                setCodeIntelData({
                    implementations: { endCursor: null, nodes: [] },
                    definitions: { endCursor: null, nodes: [] },
                    references: {
                        endCursor: null,
                        nodes: sorted,
                    },
                })
            } else {
                setCodeIntelData({
                    implementations: previousData.implementations,
                    definitions: previousData.definitions,
                    references: {
                        endCursor: null,
                        nodes: sorted,
                    },
                })
            }
        },
    })

    const [fetchSearchBasedDefinitions, fetchSearchBasedDefinitionsResult] = useLazyQuery<
        CodeIntelSearchResult,
        CodeIntelSearchVariables
    >(CODE_INTEL_SEARCH_QUERY, {
        fetchPolicy: 'no-cache',
        onCompleted: result => {
            const newDefinitions = searchResultsToLocations(result).map(buildSearchBasedLocation)
            const sorted = sortByProximity(newDefinitions, location.pathname)
            // Definitions are filtered based on the LanguageSpec
            const filteredDefinitions = filterDefinitions(sorted)

            const previousData = codeIntelData
            if (!previousData) {
                setCodeIntelData({
                    implementations: { endCursor: null, nodes: [] },
                    references: { endCursor: null, nodes: [] },
                    definitions: {
                        endCursor: null,
                        nodes: filteredDefinitions,
                    },
                })
            } else {
                setCodeIntelData({
                    implementations: previousData.implementations,
                    references: previousData.references,
                    definitions: {
                        endCursor: null,
                        nodes: filteredDefinitions,
                    },
                })
            }
        },
    })

    const fetchSearchBasedReferencesForToken = useCallback(
        (searchToken: string) => {
            const terms = referencesQuery({ searchToken, path: variables.path, fileExts: spec.fileExts })
            const query = terms.join(' ')
            return fetchSearchBasedReferences({ variables: { query } })
        },
        [fetchSearchBasedReferences, variables.path, spec.fileExts]
    )

    const fetchSearchBasedDefinitionsForToken = useCallback(
        (searchToken: string) => {
            const query = definitionQuery({ searchToken, path: variables.path, fileExts: spec.fileExts }).join(' ')
            return fetchSearchBasedDefinitions({ variables: { query } })
        },
        [fetchSearchBasedDefinitions, variables.path, spec.fileExts]
    )

    const { error, loading } = useQuery<
        UsePreciseCodeIntelForPositionResult,
        UsePreciseCodeIntelForPositionVariables & ConnectionQueryArguments
    >(USE_PRECISE_CODE_INTEL_FOR_POSITION_QUERY, {
        variables,
        notifyOnNetworkStatusChange: false,
        fetchPolicy: 'no-cache',
        skip: !shouldFetchPrecise,
        onCompleted: result => {
            if (shouldFetchPrecise.current) {
                shouldFetchPrecise.current = false

                const lsifData = result ? getLsifData({ data: result }) : undefined
                if (lsifData) {
                    setCodeIntelData(lsifData)
                } else {
                    console.info('No LSIF data. Falling back to search-based code intelligence.')
                    fellBackToSearchBased.current = true

                    // eslint-disable-next-line @typescript-eslint/no-floating-promises
                    fetchSearchBasedDefinitionsForToken(searchToken)
                    // eslint-disable-next-line @typescript-eslint/no-floating-promises
                    fetchSearchBasedReferencesForToken(searchToken)
                }
            }
        },
    })

    const [fetchAdditionalReferences, additionalReferencesResult] = useLazyQuery<
        LoadAdditionalReferencesResult,
        LoadAdditionalReferencesVariables & ConnectionQueryArguments
    >(LOAD_ADDITIONAL_REFERENCES_QUERY, {
        fetchPolicy: 'no-cache',
        onCompleted: result => {
            const previousData = codeIntelData

            const newReferenceData = result.repository?.commit?.blob?.lsif?.references

            if (!previousData || !newReferenceData) {
                return
            }

            setCodeIntelData({
                implementations: previousData.implementations,
                definitions: previousData.definitions,
                references: {
                    endCursor: newReferenceData.pageInfo.endCursor,
                    nodes: [...previousData.references.nodes, ...newReferenceData.nodes.map(buildPreciseLocation)],
                },
            })
        },
    })

    const [fetchAdditionalImplementations, additionalImplementationsResult] = useLazyQuery<
        LoadAdditionalImplementationsResult,
        LoadAdditionalImplementationsVariables & ConnectionQueryArguments
    >(LOAD_ADDITIONAL_IMPLEMENTATIONS_QUERY, {
        fetchPolicy: 'no-cache',
        onCompleted: result => {
            const previousData = codeIntelData

            const newImplementationsData = result.repository?.commit?.blob?.lsif?.implementations

            if (!previousData || !newImplementationsData) {
                return
            }

            setCodeIntelData({
                references: previousData.references,
                definitions: previousData.definitions,
                implementations: {
                    endCursor: newImplementationsData.pageInfo.endCursor,
                    nodes: [
                        ...previousData.implementations.nodes,
                        ...newImplementationsData.nodes.map(buildPreciseLocation),
                    ],
                },
            })
        },
    })

    const fetchMoreReferences = (): void => {
        const cursor = codeIntelData?.references.endCursor || null

        if (cursor !== null) {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            fetchAdditionalReferences({
                variables: {
                    ...variables,
                    ...{ afterReferences: cursor },
                },
            })
        }
    }

    const fetchMoreImplementations = (): void => {
        const cursor = codeIntelData?.implementations.endCursor || null

        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        fetchAdditionalImplementations({
            variables: {
                ...variables,
                ...{ afterImplementations: cursor },
            },
        })
    }

    const combinedLoading =
        loading ||
        (fellBackToSearchBased.current &&
            (fetchSearchBasedReferencesResult.loading || fetchSearchBasedDefinitionsResult.loading))

    const combinedError = error || fetchSearchBasedReferencesResult.error || fetchSearchBasedDefinitionsResult.error

    return {
        data: codeIntelData,
        loading: combinedLoading,

        error: combinedError,

        fetchMoreReferences,
        fetchMoreReferencesLoading: additionalReferencesResult.loading,
        referencesHasNextPage: codeIntelData ? codeIntelData.references.endCursor !== null : false,

        fetchMoreImplementations,
        implementationsHasNextPage: codeIntelData ? codeIntelData.implementations.endCursor !== null : false,
        fetchMoreImplementationsLoading: additionalImplementationsResult.loading,
    }
}

const getLsifData = ({
    data,
    error,
}: Pick<QueryResult<UsePreciseCodeIntelForPositionResult>, 'data' | 'error'>): CodeIntelData | undefined => {
    const result = asGraphQLResult({ data, errors: error?.graphQLErrors || [] })

    const extractedData = dataOrThrowErrors(result)

    // If there weren't any errors and we just didn't receive any data
    if (!extractedData || !extractedData.repository?.commit?.blob?.lsif) {
        return undefined
    }

    const lsif = extractedData.repository?.commit?.blob?.lsif

    return {
        implementations: {
            endCursor: lsif.implementations.pageInfo.endCursor,
            nodes: lsif.implementations.nodes.map(buildPreciseLocation),
        },
        references: {
            endCursor: lsif.references.pageInfo.endCursor,
            nodes: lsif.references.nodes.map(buildPreciseLocation),
        },
        definitions: {
            endCursor: lsif.definitions.pageInfo.endCursor,
            nodes: lsif.definitions.nodes.map(buildPreciseLocation),
        },
    }
}

function searchResultsToLocations(result: CodeIntelSearchResult): LocationFields[] {
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
