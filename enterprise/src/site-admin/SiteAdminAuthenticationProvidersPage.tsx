import SettingsIcon from 'mdi-react/SettingsIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../../../packages/webapp/src/backend/graphql'
import * as GQL from '../../../packages/webapp/src/backend/graphqlschema'
import { FilteredConnection } from '../../../packages/webapp/src/components/FilteredConnection'
import { PageTitle } from '../../../packages/webapp/src/components/PageTitle'
import { eventLogger } from '../../../packages/webapp/src/tracking/eventLogger'
import { createAggregateError } from '../../../packages/webapp/src/util/errors'

interface AuthProviderNodeProps {
    /** The auth provider to display in this item. */
    node: GQL.IAuthProvider
}

/** Whether to show experimental auth features. */
export const authExp = localStorage.getItem('authExp') !== null

class AuthProviderNode extends React.PureComponent<AuthProviderNodeProps> {
    public render(): JSX.Element | null {
        return (
            <li className="list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    <div className="mr-2">
                        <strong>{this.props.node.displayName}</strong>{' '}
                        <span className="badge badge-secondary">{this.props.node.serviceType}</span>
                        <br />
                        {(this.props.node.serviceID || this.props.node.clientID) && (
                            <small className="text-muted">
                                {this.props.node.serviceID}
                                {this.props.node.clientID && <> &mdash; {this.props.node.clientID}</>}
                            </small>
                        )}
                    </div>
                    {authExp && (
                        <div className="text-nowrap">
                            {this.props.node.authenticationURL && (
                                <a className="btn btn-secondary" href={this.props.node.authenticationURL}>
                                    Authenticate
                                </a>
                            )}
                        </div>
                    )}
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
export class SiteAdminAuthenticationProvidersPage extends React.Component<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminAuthentication')
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-authentication-page">
                <PageTitle title="Authentication providers - Admin" />
                <h2>Authentication providers</h2>
                <p>
                    Authentication providers allow users to sign into Sourcegraph. See{' '}
                    <a href="https://about.sourcegraph.com/docs/config/authentication">authentication documentation</a>{' '}
                    about configuring single-sign-on (SSO) via SAML and OpenID Connect.
                </p>
                <div>
                    <Link to="/site-admin/configuration" className="btn btn-secondary">
                        <SettingsIcon className="icon-inline" /> Configure auth providers
                    </Link>
                </div>
                <FilteredAuthProviderConnection
                    className="mt-3"
                    listClassName="list-group list-group-flush"
                    noun="authentication provider"
                    pluralNoun="authentication providers"
                    queryConnection={this.queryAuthProviders}
                    nodeComponent={AuthProviderNode}
                    hideSearch={true}
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
