import { ApolloError, QueryResult, WatchQueryFetchPolicy } from '@apollo/client'

import { dataOrThrowErrors, getDocumentNode, useQuery } from '@sourcegraph/http-client'
import { asGraphQLResult } from '@sourcegraph/web/src/components/FilteredConnection/utils'

import { ConnectionQueryArguments } from '../components/FilteredConnection'
import {
    UsePreciseCodeIntelForPositionVariables,
    UsePreciseCodeIntelForPositionResult,
    PreciseCodeIntelForLocationFields,
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

    implementationsHasNextPage: boolean
    fetchMoreImplementations: () => void
}

interface UsePreciseCodeIntelConfig {
    /** Allows modifying how the query interacts with the Apollo cache */
    fetchPolicy?: WatchQueryFetchPolicy
}

interface UsePreciseCodeIntelParameters {
    variables: UsePreciseCodeIntelForPositionVariables & ConnectionQueryArguments
    options?: UsePreciseCodeIntelConfig
}

export const usePreciseCodeIntel = ({
    variables,
    options,
}: UsePreciseCodeIntelParameters): UsePreciseCodeIntelResult => {
    const { data, error, loading, fetchMore } = useQuery<
        UsePreciseCodeIntelForPositionResult,
        UsePreciseCodeIntelForPositionVariables & ConnectionQueryArguments
    >(USE_PRECISE_CODE_INTEL_FOR_POSITION_QUERY, {
        variables,
        notifyOnNetworkStatusChange: true,
        fetchPolicy: options?.fetchPolicy,
    })

    const lsifData = data ? getLsifData({ data, error }) : undefined

    const fetchMoreReferences = async (): Promise<void> => {
        const cursor = lsifData?.references.pageInfo?.endCursor

        await fetchMore({
            query: getDocumentNode(LOAD_ADDITIONAL_REFERENCES_QUERY),
            variables: {
                ...variables,
                ...{ afterReferences: cursor },
            },
            updateQuery: (previousResult, { fetchMoreResult }) => {
                if (!fetchMoreResult) {
                    return previousResult
                }

                const previousData = getLsifData({ data: previousResult })
                const previousReferencesNodes = previousData.references.nodes

                const fetchMoreData = getLsifData({ data: fetchMoreResult })
                fetchMoreData.implementations = previousData.implementations
                fetchMoreData.definitions = previousData.definitions
                fetchMoreData.hover = previousData.hover
                fetchMoreData.references.nodes.unshift(...previousReferencesNodes)

                return fetchMoreResult
            },
        })
    }

    const fetchMoreImplementations = async (): Promise<void> => {
        const cursor = lsifData?.implementations.pageInfo?.endCursor

        await fetchMore({
            query: getDocumentNode(LOAD_ADDITIONAL_IMPLEMENTATIONS_QUERY),
            variables: {
                ...variables,
                ...{ afterImplementations: cursor },
            },
            updateQuery: (previousResult, { fetchMoreResult }) => {
                if (!fetchMoreResult) {
                    return previousResult
                }

                const previousData = getLsifData({ data: previousResult })
                const previousImplementationsNodes = previousData.implementations.nodes

                const fetchMoreData = getLsifData({ data: fetchMoreResult })
                fetchMoreData.references = previousData.references
                fetchMoreData.definitions = previousData.definitions
                fetchMoreData.hover = previousData.hover
                fetchMoreData.implementations.nodes.unshift(...previousImplementationsNodes)

                return fetchMoreResult
            },
        })
    }

    return {
        lsifData,
        loading,
        error,
        fetchMoreReferences,
        fetchMoreImplementations,
        referencesHasNextPage: lsifData ? lsifData.references.pageInfo.endCursor !== null : false,
        implementationsHasNextPage: lsifData ? lsifData.implementations.pageInfo.endCursor !== null : false,
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
            hover: null,
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
