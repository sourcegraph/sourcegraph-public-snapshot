import { useCallback, useEffect, useRef, useState } from 'react'

import { ApolloError, QueryResult } from '@apollo/client'

import {
    appendLineRangeQueryParameter,
    appendSubtreeQueryParameter,
    toPositionOrRangeQueryParameter,
} from '@sourcegraph/common'
import { dataOrThrowErrors, useLazyQuery, useQuery } from '@sourcegraph/http-client'
import { asGraphQLResult } from '@sourcegraph/web/src/components/FilteredConnection/utils'

import { ConnectionQueryArguments } from '../components/FilteredConnection'
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

import { Location, buildPreciseLocation, buildSearchBasedLocation } from './location'
import {
    LOAD_ADDITIONAL_IMPLEMENTATIONS_QUERY,
    LOAD_ADDITIONAL_REFERENCES_QUERY,
    LOAD_ADDITIONAL_REFERENCES_SEARCH_BASED_QUERY,
    USE_PRECISE_CODE_INTEL_FOR_POSITION_QUERY,
} from './ReferencesPanelQueries'
import { definitionQuery, referencesQuery } from './searchBased'

interface CodeIntelResults {
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

export interface UsePreciseCodeIntelResult {
    results?: CodeIntelResults
    error?: ApolloError
    loading: boolean

    referencesHasNextPage: boolean
    fetchMoreReferences: () => void
    fetchMoreReferencesLoading: boolean

    implementationsHasNextPage: boolean
    fetchMoreImplementations: () => void
    fetchMoreImplementationsLoading: boolean
}

interface UsePreciseCodeIntelParameters {
    variables: UsePreciseCodeIntelForPositionVariables & ConnectionQueryArguments
}

export const usePreciseCodeIntel = ({ variables }: UsePreciseCodeIntelParameters): UsePreciseCodeIntelResult => {
    const [codeIntelResults, setCodeIntelResults] = useState<CodeIntelResults>()

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
    >(LOAD_ADDITIONAL_REFERENCES_SEARCH_BASED_QUERY, {
        fetchPolicy: 'no-cache',
        onCompleted: result => {
            const newReferences = searchResultsToLocations(result).map(buildSearchBasedLocation)

            const previousData = codeIntelResults
            if (!previousData) {
                setCodeIntelResults({
                    implementations: {
                        endCursor: null,
                        nodes: [],
                    },
                    definitions: {
                        endCursor: null,
                        nodes: [],
                    },
                    references: {
                        endCursor: null,
                        nodes: newReferences,
                    },
                })
            } else {
                setCodeIntelResults({
                    implementations: previousData.implementations,
                    definitions: previousData.definitions,
                    references: {
                        endCursor: null,
                        nodes: [...previousData.references.nodes, ...newReferences],
                    },
                })
            }
        },
    })

    const [fetchSearchBasedDefinitions, fetchSearchBasedDefinitionsResult] = useLazyQuery<
        CodeIntelSearchResult,
        CodeIntelSearchVariables
    >(LOAD_ADDITIONAL_REFERENCES_SEARCH_BASED_QUERY, {
        fetchPolicy: 'no-cache',
        onCompleted: result => {
            const newDefinitions = searchResultsToLocations(result).map(buildSearchBasedLocation)

            const previousData = codeIntelResults
            if (!previousData) {
                setCodeIntelResults({
                    implementations: { endCursor: null, nodes: [] },
                    references: { endCursor: null, nodes: [] },
                    definitions: {
                        endCursor: null,
                        nodes: newDefinitions,
                    },
                })
            } else {
                setCodeIntelResults({
                    implementations: previousData.implementations,
                    references: previousData.references,
                    definitions: {
                        endCursor: null,
                        nodes: [...previousData.definitions.nodes, ...newDefinitions],
                    },
                })
            }
        },
    })

    const fetchSearchBasedReferencesForToken = useCallback(() => {
        const terms = referencesQuery({ searchToken: 'copyRouteConf', path: variables.path, fileExts: ['go'] })
        const query = terms.join(' ')
        return fetchSearchBasedReferences({ variables: { query } })
    }, [fetchSearchBasedReferences, variables.path])

    const fetchSearchBasedDefinitionsForToken = useCallback(() => {
        const query = definitionQuery({ searchToken: 'GetRoute', path: variables.path, fileExts: ['go'] }).join(' ')
        return fetchSearchBasedDefinitions({ variables: { query } })
    }, [fetchSearchBasedDefinitions, variables.path])

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
                    setCodeIntelResults(lsifData)
                } else {
                    console.info('No LSIF data. Falling back to search-based code intelligence.')
                    fellBackToSearchBased.current = true
                    fetchSearchBasedDefinitionsForToken()
                    fetchSearchBasedReferencesForToken()
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
            console.log('fetch additional references')
            const previousData = codeIntelResults

            const newReferenceData = result.repository?.commit?.blob?.lsif?.references

            if (!previousData || !newReferenceData) {
                return
            }

            setCodeIntelResults({
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
            const previousData = codeIntelResults

            const newImplementationsData = result.repository?.commit?.blob?.lsif?.implementations

            if (!previousData || !newImplementationsData) {
                return
            }

            setCodeIntelResults({
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
        console.log('fetchMoreReferences')
        const cursor = codeIntelResults?.references.endCursor || null

        if (cursor === null && attemptedSearchReferences === false) {
            setAttemptedSearchReferences(true)
            fetchSearchBasedReferences({
                variables: {
                    // TODO: fix all of this
                    query:
                        'GetRoute repo:github\\.com\\/gorilla\\/mux$ type:file patternType:regexp count:500 case:yes',
                },
            })
        } else if (cursor !== null) {
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
        const cursor = codeIntelResults?.implementations.endCursor || null

        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        fetchAdditionalImplementations({
            variables: {
                ...variables,
                ...{ afterImplementations: cursor },
            },
        })
    }

    console.log('fellBackToSearchBased', fellBackToSearchBased.current)
    const combinedLoading =
        loading ||
        (fellBackToSearchBased.current &&
            (fetchSearchBasedReferencesResult.loading || fetchSearchBasedDefinitionsResult.loading))

    const combinedError = error || fetchSearchBasedReferencesResult.error || fetchSearchBasedDefinitionsResult.error

    return {
        results: codeIntelResults,
        loading: combinedLoading,

        error: combinedError,

        fetchMoreReferences,
        fetchMoreReferencesLoading: additionalReferencesResult.loading,
        referencesHasNextPage: codeIntelResults ? codeIntelResults.references.endCursor !== null : false,

        fetchMoreImplementations,
        implementationsHasNextPage: codeIntelResults ? codeIntelResults.implementations.endCursor !== null : false,
        fetchMoreImplementationsLoading: additionalImplementationsResult.loading,
    }
}

const getLsifData = ({
    data,
    error,
}: Pick<QueryResult<UsePreciseCodeIntelForPositionResult>, 'data' | 'error'>): CodeIntelResults | undefined => {
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
    console.log('searchResults', searchResults)
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
            console.log('lineMatch', lineMatch)
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
