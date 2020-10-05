import React, { useMemo } from 'react'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { requestGraphQL } from '../../../backend/graphql'
import { NamespaceAreaContext } from '../../../namespaces/NamespaceArea'
import { ListGraphsResult, ListGraphsVariables } from '../../../graphql-operations'
import { map } from 'rxjs/operators'
import { Link } from 'react-router-dom'
import PlusIcon from 'mdi-react/PlusIcon'
import { GraphList } from '../shared/graphList/GraphList'
import { GraphListItemFragmentGQL } from '../shared/graphList/GraphListItem'
import { GraphSelectionProps } from '../selector/graphSelectionProps'

interface Props extends NamespaceAreaContext, GraphSelectionProps {}

export const GraphOwnerListGraphsPage: React.FunctionComponent<Props> = ({ namespace, ...props }) => {
    const graphs = useObservable(
        useMemo(
            () =>
                requestGraphQL<ListGraphsResult, ListGraphsVariables>(
                    gql`
                        query ListGraphs($graphOwner: ID!) {
                            node(id: $graphOwner) {
                                ... on GraphOwner {
                                    graphs {
                                        nodes {
                                            ...GraphListItem
                                        }
                                        totalCount
                                    }
                                }
                            }
                        }
                        ${GraphListItemFragmentGQL}
                    `,
                    // TODO(sqs): paginate with `first`
                    { graphOwner: namespace.id }
                ).pipe(
                    map(dataOrThrowErrors),
                    map(data => data.node?.graphs)
                ),
            [namespace.id]
        )
    )

    return (
        <div className="container">
            <div className="d-flex justify-content-end mb-2">
                <Link to={`${namespace.url}/graphs/new`} className="btn btn-secondary">
                    <PlusIcon className="icon-inline" /> New graph
                </Link>
            </div>
            <GraphList {...props} graphs={graphs} />
        </div>
    )
}
