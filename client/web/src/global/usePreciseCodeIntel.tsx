import { ApolloError, QueryResult } from '@apollo/client'
import { useEffect, useRef, useState } from 'react'

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
} from '../graphql-operations'

import {
    LOAD_ADDITIONAL_IMPLEMENTATIONS_QUERY,
    LOAD_ADDITIONAL_REFERENCES_QUERY,
    USE_PRECISE_CODE_INTEL_FOR_POSITION_QUERY,
} from './CoolCodeIntelQueries'

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

    const [fetchAdditionalReferences, additionalReferencesResult] = useLazyQuery<
        LoadAdditionalReferencesResult,
        LoadAdditionalReferencesVariables & ConnectionQueryArguments
    >(LOAD_ADDITIONAL_REFERENCES_QUERY, {
        fetchPolicy: 'no-cache',
        onCompleted: result => {
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
        const cursor = referenceData?.references.pageInfo?.endCursor || null

        fetchAdditionalReferences({
            variables: {
                ...variables,
                ...{ afterReferences: cursor },
            },
        })
    }

    const fetchMoreImplementations = (): void => {
        const cursor = referenceData?.implementations.pageInfo?.endCursor || null

        fetchAdditionalImplementations({
            variables: {
                ...variables,
                ...{ afterImplementations: cursor },
            },
        })
    }

    return {
        lsifData: referenceData,
        loading,
        error,

        fetchMoreReferences,
        fetchMoreReferencesLoading: additionalReferencesResult.loading,
        referencesHasNextPage: referenceData ? referenceData.references.pageInfo.endCursor !== null : false,

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
