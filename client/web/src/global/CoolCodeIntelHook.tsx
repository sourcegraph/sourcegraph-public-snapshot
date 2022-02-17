import { ApolloError, QueryResult, WatchQueryFetchPolicy } from '@apollo/client'

import { getDocumentNode, GraphQLResult, useQuery } from '@sourcegraph/http-client'
import { asGraphQLResult } from '@sourcegraph/web/src/components/FilteredConnection/utils'

import { getWebGraphQLClient } from '../backend/graphql'
import { ConnectionQueryArguments } from '../components/FilteredConnection'
import {
    GetPreciseCodeIntelVariables,
    LoadAdditionalReferencesResult,
    LoadAdditionalReferencesVariables,
    RefPanelLsifDataFields,
} from '../graphql-operations'

import { LOAD_ADDITIONAL_REFERENCES_QUERY } from './CoolCodeIntelQueries'

export interface UsePreciseCodeIntelResult {
    lsifData?: RefPanelLsifDataFields
    error?: ApolloError
    loading: boolean

    referencesHasNextPage: boolean
    implementationsHasNextPage: boolean
    fetchMoreReferences: () => void
    fetchMoreImplementations: () => void
}

interface UsePreciseCodeIntelConfig {
    /** Set if query variables should be updated in and derived from the URL */
    useURL?: boolean
    /** Allows modifying how the query interacts with the Apollo cache */
    fetchPolicy?: WatchQueryFetchPolicy
    /** Set to enable polling of all the nodes currently loaded in the connection */
    pollInterval?: number
}

interface UsePreciseCodeIntelParameters<TResult> {
    query: string
    variables: GetPreciseCodeIntelVariables & ConnectionQueryArguments
    getConnection: (result: GraphQLResult<TResult>) => RefPanelLsifDataFields
    options?: UsePreciseCodeIntelConfig
}

export const usePreciseCodeIntel = <TResult,>({
    query,
    variables,
    getConnection: getLsifDataFromGraphQLResult,
    options,
}: UsePreciseCodeIntelParameters<TResult>): UsePreciseCodeIntelResult => {
    const { data, error, loading, fetchMore } = useQuery<
        TResult,
        GetPreciseCodeIntelVariables & ConnectionQueryArguments
    >(query, {
        variables,
        notifyOnNetworkStatusChange: true, // Ensures loading state is updated on `fetchMore`
        fetchPolicy: options?.fetchPolicy,
    })

    /**
     * Map over Apollo results to provide type-compatible `GraphQLResult`s for consumers.
     * This ensures good interoperability between `FilteredConnection` and `useConnection`.
     */
    const getLsifData = ({ data, error }: Pick<QueryResult<TResult>, 'data' | 'error'>): RefPanelLsifDataFields => {
        const result = asGraphQLResult({ data, errors: error?.graphQLErrors || [] })
        return getLsifDataFromGraphQLResult(result)
    }

    const lsifData = data ? getLsifData({ data, error }) : undefined

    // TODO: This does not work
    const fetchMoreReferences = async (): Promise<void> => {
        const cursor = lsifData?.references.pageInfo?.endCursor
        if (!cursor) {
            return
        }

        const client = await getWebGraphQLClient()
        const { data: additionalData, error } = await client.query<
            LoadAdditionalReferencesResult,
            LoadAdditionalReferencesVariables
        >({
            query: getDocumentNode(LOAD_ADDITIONAL_REFERENCES_QUERY),
            variables: {
                ...variables,
                ...{ afterReferences: cursor },
            },
        })

        if (error) {
            console.log('error', error)
            return
        }

        if (!additionalData || !additionalData.repository?.commit?.blob?.lsif) {
            console.log('empty')
            return
        }

        const lsif = additionalData.repository?.commit?.blob?.lsif
        console.log('current references', lsifData.references.nodes)
        console.log('additional references', lsif.references.nodes)
    }

    const fetchMoreImplementations = async (): Promise<void> => {
        const cursor = lsifData?.implementations.pageInfo?.endCursor

        await fetchMore({
            variables: {
                ...variables,
                ...{ afterImplementations: cursor },
            },
            updateQuery: (previousResult, { fetchMoreResult }) => {
                if (!fetchMoreResult) {
                    return previousResult
                }

                const previousData = getLsifData({ data: previousResult })
                const previousImplementationNodes = previousData.implementations.nodes
                const previousReferencesNodes = previousData.references.nodes

                const fetchMoreData = getLsifData({ data: fetchMoreResult })
                fetchMoreData.implementations.nodes.unshift(...previousImplementationNodes)
                fetchMoreData.references.nodes = previousReferencesNodes

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
