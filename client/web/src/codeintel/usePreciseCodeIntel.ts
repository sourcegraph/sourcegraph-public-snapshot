import { useEffect, useRef, useState } from 'react'

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
    PreciseCodeIntelForLocationFields,
    LoadAdditionalReferencesResult,
    LoadAdditionalReferencesVariables,
    LoadAdditionalImplementationsResult,
    LoadAdditionalImplementationsVariables,
    CodeIntelSearchResult,
    CodeIntelSearchVariables,
    LocationFields,
} from '../graphql-operations'

import { Location } from './ReferencesPanel'
import {
    LOAD_ADDITIONAL_IMPLEMENTATIONS_QUERY,
    LOAD_ADDITIONAL_REFERENCES_QUERY,
    LOAD_ADDITIONAL_REFERENCES_SEARCH_BASED_QUERY,
    USE_PRECISE_CODE_INTEL_FOR_POSITION_QUERY,
} from './ReferencesPanelQueries'

interface CodeIntelData {
    references: Location[]
    implementations: Location[]
    definitions: Location[]
}

export interface UsePreciseCodeIntelResult {
    lsifData?: PreciseCodeIntelForLocationFields
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
    const [referenceData, setReferenceData] = useState<PreciseCodeIntelForLocationFields>()
    const [attemptedSearchReferences, setAttemptedSearchReferences] = useState<boolean>(false)

    const shouldFetch = useRef(true)
    useEffect(() => {
        // We need to fetch again if the variables change
        shouldFetch.current = true
    }, [variables])

    const { error, loading } = useQuery<
        UsePreciseCodeIntelForPositionResult,
        UsePreciseCodeIntelForPositionVariables & ConnectionQueryArguments
    >(USE_PRECISE_CODE_INTEL_FOR_POSITION_QUERY, {
        variables,
        notifyOnNetworkStatusChange: false,
        fetchPolicy: 'no-cache',
        skip: !shouldFetch,
        onCompleted: result => {
            if (shouldFetch.current) {
                const lsifData = result ? getLsifData({ data: result }) : undefined
                setReferenceData(lsifData)
                shouldFetch.current = false
            }
        },
    })

    const [fetchSearchBasedReferences, additionalReferencesResultSEARCH] = useLazyQuery<
        CodeIntelSearchResult,
        CodeIntelSearchVariables
    >(LOAD_ADDITIONAL_REFERENCES_SEARCH_BASED_QUERY, {
        fetchPolicy: 'no-cache',
        onCompleted: result => {
            if (!result || !result.search) {
                console.log('no results ')
                return
            }

            const searchResults = result.search.results.results
                .filter(value => value !== undefined)
                .filter(result => result.__typename === 'FileMatch')
            console.log('searchResults', searchResults)
            // const newReferenceData = result.repository?.commit?.blob?.lsif?.references

            const newReferences: LocationFields[] = []
            for (const result of searchResults) {
                if (result.__typename !== 'FileMatch') {
                    continue
                }

                for (const lineMatch of result.lineMatches) {
                    const positionOrRangeQueryParameter = toPositionOrRangeQueryParameter({
                        // TODO: only using first offset?
                        position: { line: lineMatch.lineNumber, character: lineMatch.offsetAndLengths[0][0] },
                    })
                    const url = appendLineRangeQueryParameter(
                        appendSubtreeQueryParameter(result.file.url),
                        positionOrRangeQueryParameter
                    )
                    newReferences.push({
                        url,
                        resource: {
                            path: result.file.path,
                            content: result.file.content,
                            repository: result.repository,
                            commit: {
                                oid: result.file.commit.oid,
                            },
                        },
                        range: {
                            start: {
                                line: lineMatch.lineNumber,
                                character: lineMatch.offsetAndLengths[0][0],
                            },
                            end: {
                                line: lineMatch.lineNumber,
                                character: lineMatch.offsetAndLengths[0][2],
                            },
                        },
                    })
                }

                const symbolReferences = result.symbols.map(symbol => ({
                    url: symbol.location.url,
                    resource: {
                        path: symbol.location.resource.path,
                        content: '',
                        commit: {
                            oid: result.file.commit.oid,
                        },
                        repository: result.repository,
                    },
                    range: symbol.location.range,
                }))
                for (const symbolReference of symbolReferences) {
                    newReferences.push(symbolReference)
                }
            }
            console.log('newReferences', newReferences)

            const previousData = referenceData
            if (!previousData) {
                console.log('here')
                setReferenceData({
                    implementations: {
                        pageInfo: { endCursor: null },
                        nodes: [],
                    },
                    definitions: {
                        pageInfo: { endCursor: null },
                        nodes: [],
                    },
                    references: {
                        pageInfo: { endCursor: null },
                        nodes: newReferences,
                    },
                })
            } else {
                console.log('there')
                setReferenceData({
                    implementations: previousData.implementations,
                    definitions: previousData.definitions,
                    references: {
                        pageInfo: { endCursor: null },
                        nodes: [...previousData.references.nodes, ...newReferences],
                    },
                })
            }
            // if (!previousData || !newReferenceData) {
            //     return
            // }

            // setReferenceData({
            //     implementations: previousData.implementations,
            //     definitions: previousData.definitions,
            //     references: {
            //         ...newReferenceData,
            //         nodes: [...previousData.references.nodes, ...newReferenceData.nodes],
            //     },
            // })
        },
    })

    const [fetchAdditionalReferences, additionalReferencesResult] = useLazyQuery<
        LoadAdditionalReferencesResult,
        LoadAdditionalReferencesVariables & ConnectionQueryArguments
    >(LOAD_ADDITIONAL_REFERENCES_QUERY, {
        fetchPolicy: 'no-cache',
        onCompleted: result => {
            console.log('fetch additional references')
            const previousData = referenceData

            const newReferenceData = result.repository?.commit?.blob?.lsif?.references

            if (!previousData || !newReferenceData) {
                return
            }

            setReferenceData({
                implementations: previousData.implementations,
                definitions: previousData.definitions,
                references: {
                    ...newReferenceData,
                    nodes: [...previousData.references.nodes, ...newReferenceData.nodes],
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
            const previousData = referenceData

            const newImplementationsData = result.repository?.commit?.blob?.lsif?.implementations

            if (!previousData || !newImplementationsData) {
                return
            }

            setReferenceData({
                references: previousData.references,
                definitions: previousData.definitions,
                implementations: {
                    ...newImplementationsData,
                    nodes: [...previousData.implementations.nodes, ...newImplementationsData.nodes],
                },
            })
        },
    })

    const fetchMoreReferences = (): void => {
        console.log('fetchMoreReferences')
        const cursor = referenceData?.references.pageInfo?.endCursor || null

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
        const cursor = referenceData?.implementations.pageInfo?.endCursor || null

        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        fetchAdditionalImplementations({
            variables: {
                ...variables,
                ...{ afterImplementations: cursor },
            },
        })
    }

    console.log('referenceData', referenceData)
    return {
        lsifData: referenceData,
        loading,
        error,

        fetchMoreReferences,
        fetchMoreReferencesLoading: additionalReferencesResult.loading || additionalReferencesResultSEARCH.loading,
        referencesHasNextPage: referenceData
            ? referenceData.references.pageInfo.endCursor !== null || attemptedSearchReferences === false
            : false,

        fetchMoreImplementations,
        implementationsHasNextPage: referenceData ? referenceData.implementations.pageInfo.endCursor !== null : false,
        fetchMoreImplementationsLoading: additionalImplementationsResult.loading,
    }
}

const getLsifData = ({
    data,
    error,
}: Pick<QueryResult<UsePreciseCodeIntelForPositionResult>, 'data' | 'error'>): PreciseCodeIntelForLocationFields => {
    const result = asGraphQLResult({ data, errors: error?.graphQLErrors || [] })

    const extractedData = dataOrThrowErrors(result)

    // If there weren't any errors and we just didn't receive any data
    if (!extractedData || !extractedData.repository?.commit?.blob?.lsif) {
        return {
            definitions: {
                nodes: [],
                pageInfo: {
                    endCursor: null,
                },
            },
            references: {
                nodes: [],
                pageInfo: {
                    endCursor: null,
                },
            },
            implementations: {
                nodes: [],
                pageInfo: {
                    endCursor: null,
                },
            },
        }
    }

    const lsif = extractedData.repository?.commit?.blob?.lsif

    return lsif
}
