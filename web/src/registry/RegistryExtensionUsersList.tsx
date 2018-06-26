import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { FilteredConnection } from '../components/FilteredConnection'
import { createAggregateError } from '../util/errors'

const registryExtensionUserFragment = gql`
    fragment RegistryExtensionUserFields on User {
        id
        username
        url
    }
`

const RegistryExtensionUserNode: React.SFC<{
    node: GQL.IUser
}> = ({ node }) => (
    <Link to={node.url} className="list-group-item d-flex justify-content-between align-items-center">
        {node.username}
    </Link>
)

interface Props extends RouteComponentProps<{}> {
    extension: Pick<GQL.IRegistryExtension, 'id' | 'viewerHasEnabled'>
    shouldUpdateURLQuery?: boolean
    noSummaryIfAllNodesVisible?: boolean
}

class FilteredRegistryExtensionUserConnection extends FilteredConnection<GQL.IUser> {}

/**
 * Displays the users for whom an extension is enabled.
 */
export class RegistryExtensionUsersList extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return (
            <FilteredRegistryExtensionUserConnection
                listClassName="list-group list-group-flush"
                listComponent="div"
                noun="extension user"
                pluralNoun="extension users"
                queryConnection={this.queryRegistryExtensionUsers}
                // Updating when viewerHasEnabled changes makes it so that clicking "Enable/disable extension" in
                // the header immediately updates this list.
                updateOnChange={`${this.props.extension.id}:${this.props.extension.viewerHasEnabled}`}
                nodeComponent={RegistryExtensionUserNode}
                hideSearch={true}
                noSummaryIfAllNodesVisible={this.props.noSummaryIfAllNodesVisible}
                shouldUpdateURLQuery={this.props.shouldUpdateURLQuery}
                history={this.props.history}
                location={this.props.location}
            />
        )
    }

    private queryRegistryExtensionUsers = (args: { first?: number }): Observable<GQL.IUserConnection> =>
        queryGraphQL(
            gql`
                query RegistryExtensionUsers($extension: ID!, $first: Int) {
                    node(id: $extension) {
                        ... on RegistryExtension {
                            users(first: $first) {
                                nodes {
                                    ...RegistryExtensionUserFields
                                }
                                totalCount
                            }
                        }
                    }
                }
                ${registryExtensionUserFragment}
            `,
            { ...args, extension: this.props.extension.id }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node || errors) {
                    throw createAggregateError(errors)
                }
                const node = data.node as GQL.IRegistryExtension
                if (!node.users) {
                    throw createAggregateError(errors)
                }
                return node.users
            })
        )
}
