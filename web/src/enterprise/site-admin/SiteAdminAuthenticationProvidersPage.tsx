import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../shared/src/util/errors'
import { queryGraphQL } from '../../backend/graphql'
import { FilteredConnection } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'

interface AuthProviderNodeProps {
    /** The auth provider to display in that item. */
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
                        <strong>{that.props.node.displayName}</strong>{' '}
                        <span className="badge badge-secondary">{that.props.node.serviceType}</span>
                        <br />
                        {(that.props.node.serviceID || that.props.node.clientID) && (
                            <small className="text-muted">
                                {that.props.node.serviceID}
                                {that.props.node.clientID && <> &mdash; {that.props.node.clientID}</>}
                            </small>
                        )}
                    </div>
                    {authExp && (
                        <div className="text-nowrap">
                            {that.props.node.authenticationURL && (
                                <a className="btn btn-secondary" href={that.props.node.authenticationURL}>
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
                <div className="d-flex justify-content-between align-items-center mt-3 mb-1">
                    <h2 className="mb-0">Authentication providers</h2>
                </div>
                <p>
                    Authentication providers allow users to sign into Sourcegraph. See{' '}
                    <a href="https://docs.sourcegraph.com/admin/auth">authentication documentation</a> about configuring
                    single-sign-on (SSO) via SAML and OpenID Connect. Configure authentication providers in the{' '}
                    <a href="https://docs.sourcegraph.com/admin/config/site_config">site configuration</a>.
                </p>
                <FilteredAuthProviderConnection
                    className="list-group list-group-flush mt-3"
                    noun="authentication provider"
                    pluralNoun="authentication providers"
                    queryConnection={that.queryAuthProviders}
                    nodeComponent={AuthProviderNode}
                    hideSearch={true}
                    history={that.props.history}
                    location={that.props.location}
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
