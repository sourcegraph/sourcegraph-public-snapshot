import * as React from 'react'

import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'
import { Badge, Link, H2, Text, AnchorLink, Button } from '@sourcegraph/wildcard'

import { queryGraphQL } from '../../backend/graphql'
import { FilteredConnection } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'

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
                        <Badge variant="secondary">{this.props.node.serviceType}</Badge>
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
                                <Button to={this.props.node.authenticationURL} variant="secondary" as={AnchorLink}>
                                    Authenticate
                                </Button>
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
                <H2>Authentication providers</H2>
                <Text>
                    Authentication providers allow users to sign into Sourcegraph. See{' '}
                    <Link to="/help/admin/auth">authentication documentation</Link> about configuring single-sign-on
                    (SSO) via SAML and OpenID Connect. Configure authentication providers in the{' '}
                    <Link to="/help/admin/config/site_config">site configuration</Link>.
                </Text>
                <FilteredAuthProviderConnection
                    className="list-group list-group-flush mt-3"
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
