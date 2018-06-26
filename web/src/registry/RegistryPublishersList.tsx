import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { FilteredConnection } from '../components/FilteredConnection'
import { createAggregateError } from '../util/errors'
import { registryPublisherFragment, RegistryPublisherNode } from './RegistryPublisherNode'

interface Props extends RouteComponentProps<{}> {}

class FilteredRegistryPublisherConnection extends FilteredConnection<GQL.RegistryPublisher> {}

/**
 * Displays registry publishers.
 */
export const RegistryPublishersList: React.SFC<Props> = props => (
    <FilteredRegistryPublisherConnection
        listClassName="list-group list-group-flush"
        listComponent="div"
        noun="publishers"
        pluralNoun="publishers"
        queryConnection={queryRegistryPublishers}
        nodeComponent={RegistryPublisherNode}
        hideSearch={true}
        noSummaryIfAllNodesVisible={true}
        history={props.history}
        location={props.location}
    />
)

function queryRegistryPublishers(args: { first?: number }): Observable<GQL.IRegistryPublisherConnection> {
    return queryGraphQL(
        gql`
            query RegistryPublishers($first: Int) {
                extensionRegistry {
                    publishers(first: $first) {
                        nodes {
                            ...RegistryPublisherFields
                        }
                        totalCount
                        pageInfo {
                            hasNextPage
                        }
                    }
                }
            }
            ${registryPublisherFragment}
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.extensionRegistry || !data.extensionRegistry.publishers || errors) {
                throw createAggregateError(errors)
            }
            return data.extensionRegistry.publishers
        })
    )
}
