import React from 'react'
import { Observable } from 'rxjs'
import { filter, map } from 'rxjs/operators'
import { View, ViewContexts } from '../../../../../shared/src/api/client/services/viewService'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../../backend/graphql'
import { NamespaceGraphsResult, NamespaceGraphsVariables } from '../../../graphql-operations'
import { pluralize } from '../../../../../shared/src/util/strings'
import { ContributableViewContainer } from '../../../../../shared/src/api/protocol'
import { isDefined } from '../../../../../shared/src/util/types'

export const namespaceGraphs = ({
    id,
}: ViewContexts[typeof ContributableViewContainer.Profile]): Observable<View | null> => {
    const graphs = requestGraphQL<NamespaceGraphsResult, NamespaceGraphsVariables>(
        gql`
            query NamespaceGraphs($id: ID!, $first: Int!) {
                node(id: $id) {
                    ... on GraphOwner {
                        graphs(first: $first) {
                            nodes {
                                name
                                url
                            }
                            totalCount
                        }
                    }
                }
            }
        `,
        { id, first: 5 }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.node?.graphs)
    )

    return graphs.pipe(
        filter(isDefined),
        map(graphs => ({
            title: `${graphs.totalCount} ${pluralize('graph', graphs.totalCount)}`,
            titleLink: '/users/sqs/graphs', // TODO(sqs)
            content: [
                {
                    reactComponent: () => <div>Graphs: ${JSON.stringify(graphs)}</div>,
                },
            ],
        }))
    )
}
