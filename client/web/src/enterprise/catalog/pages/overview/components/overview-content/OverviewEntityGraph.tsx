import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'

import { CatalogGraphResult, CatalogGraphVariables } from '../../../../../../graphql-operations'
import { EntityGraph } from '../../../../components/entity-graph/EntityGraph'

import { CATALOG_GRAPH } from './gql'

interface Props {}

export const OverviewEntityGraph: React.FunctionComponent<Props> = () => {
    const { data, error, loading } = useQuery<CatalogGraphResult, CatalogGraphVariables>(CATALOG_GRAPH, {
        // Cache this data but always re-request it in the background when we revisit
        // this page to pick up newer changes.
        fetchPolicy: 'cache-and-network',

        // For subsequent requests while this page is open, make additional network
        // requests; this is necessary for `refetch` to actually use the network. (see
        // https://github.com/apollographql/apollo-client/issues/5515)
        nextFetchPolicy: 'network-only',
    })

    return error ? (
        <p>Error loading graph</p>
    ) : loading && !data ? (
        <LoadingSpinner className="icon-inline" />
    ) : !data || !data.catalog.graph ? (
        <p>Catalog graph is not available</p>
    ) : (
        <EntityGraph graph={data.catalog.graph} />
    )
}
