import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { createAggregateError } from '../util/errors'

interface AuthProviderNodeProps {
    /** The auth provider to display in this item. */
    node: GQL.IAuthProvider
}

class AuthProviderNode extends React.PureComponent<AuthProviderNodeProps> {
    public render(): JSX.Element | null {
        return (
            <li className="list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    <div className="mr-2">
                        <strong>{this.props.node.displayName}</strong>{' '}
                        <span className="badge badge-secondary">{this.props.node.serviceType}</span>
                    </div>
                    <div className="text-nowrap">
                        {this.props.node.authenticationURL && (
                            <a className="btn btn-secondary" href={this.props.node.authenticationURL}>
                                Authenticate
                            </a>
                        )}
                    </div>
                </div>
            </li>
        )
    }
}

const authProviderFragment = gql`
    fragment AuthProviderFields on AuthProvider {
        serviceType
        serviceID
        clientID
        displayName
        isBuiltin
        authenticationURL
    }
`

interface Props extends RouteComponentProps<{}> {}

class FilteredAuthProviderConnection extends FilteredConnection<GQL.IAuthProvider> {}

/**
 * A page displaying the auth providers in site configuration.
 */
export class SiteAdminAuthenticationPage extends React.Component<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminAuthentication')
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-authentication-page">
                <PageTitle title="Authentication - Admin" />
                <h2>Authentication</h2>
                <FilteredAuthProviderConnection
                    className="mt-3"
                    listClassName="list-group list-group-flush"
                    noun="authentication provider"
                    pluralNoun="authentication providers"
                    queryConnection={this.queryAuthProviders}
                    nodeComponent={AuthProviderNode}
                    noShowMore={true}
                    shouldUpdateURLQuery={false}
                    hideFilter={true}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryAuthProviders = (args: {}): Observable<GQL.IAuthProviderConnection> =>
        queryGraphQL(
            gql`
                query AuthProviders {
                    site {
                        authProviders {
                            nodes {
                                ...AuthProviderFields
                            }
                            totalCount
                            pageInfo {
                                hasNextPage
                            }
                        }
                    }
                }
                ${authProviderFragment}
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.site || !data.site.authProviders || errors) {
                    throw createAggregateError(errors)
                }
                return data.site.authProviders
            })
        )
}
