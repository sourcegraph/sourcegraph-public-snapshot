import classNames from 'classnames'
import React from 'react'

import { gql, useQuery } from '@sourcegraph/http-client'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { CatalogExplorerGraphResult, CatalogExplorerGraphVariables } from '../../../../../graphql-operations'
import { ComponentFiltersProps } from '../../../core/component-query'
import { CatalogGraph, CATALOG_GRAPH_FRAGMENT } from '../../catalog-graph/CatalogGraph'

interface Props extends Pick<ComponentFiltersProps, 'filters'> {
    highlightID?: Scalars['ID']
    queryScope?: string
    className?: string
    errorClassName?: string
}

export const CatalogExplorerGraph: React.FunctionComponent<Props> = ({
    filters,
    highlightID,
    queryScope,
    className,
    errorClassName,
}) => {
    const { data, error, loading } = useQuery<CatalogExplorerGraphResult, CatalogExplorerGraphVariables>(
        gql`
            query CatalogExplorerGraph($query: String!) {
                graph(query: $query) {
                    ...CatalogGraphFields
                }
            }

            ${CATALOG_GRAPH_FRAGMENT}
        `,
        {
            variables: {
                query: `${queryScope || ''} ${filters.query || ''}`,
            },

            // Cache this data but always re-request it in the background when we revisit
            // this page to pick up newer changes.
            fetchPolicy: 'cache-and-network',

            // For subsequent requests while this page is open, make additional network
            // requests; this is necessary for `refetch` to actually use the network. (see
            // https://github.com/apollographql/apollo-client/issues/5515)
            nextFetchPolicy: 'network-only',
        }
    )

    return error ? (
        <div className={classNames('alert alert-error', errorClassName)}>Error loading graph</div>
    ) : loading && !data ? (
        <LoadingSpinner className="icon-inline" />
    ) : !data || !data.graph ? (
        <div className={classNames('alert alert-error', errorClassName)}>Catalog graph is not available</div>
    ) : data.graph.nodes.length > 0 ? (
        <CatalogGraph
            graph={{
                ...data.graph,
                edges: data.graph.edges.filter(edge => edge.type === 'DEPENDS_ON'),
            }}
            activeNodeID={highlightID}
            className={className}
        />
    ) : (
        <p className={classNames('text-muted mb-0', errorClassName)}>Empty graph</p>
    )
}
