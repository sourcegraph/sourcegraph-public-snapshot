import { gql, queryGraphQL } from '@sourcegraph/webapp/dist/backend/graphql'
import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import { FilteredConnection } from '@sourcegraph/webapp/dist/components/FilteredConnection'
import { PageTitle } from '@sourcegraph/webapp/dist/components/PageTitle'
import { eventLogger } from '@sourcegraph/webapp/dist/tracking/eventLogger'
import { createAggregateError } from '@sourcegraph/webapp/dist/util/errors'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import {
    externalAccountFragment,
    ExternalAccountNode,
    ExternalAccountNodeProps,
} from '../user/account/ExternalAccountNode'

interface Props extends RouteComponentProps<{}> {}

class FilteredExternalAccountConnection extends FilteredConnection<
    GQL.IExternalAccount,
    Pick<ExternalAccountNodeProps, 'onDidUpdate' | 'showUser'>
> {}

interface FilterParams {
    user?: GQL.ID
    serviceType?: string
    serviceID?: string
    clientID?: string
}

/**
 * Displays the external accounts (from authentication providers) associated with the user's account.
 */
export class SiteAdminExternalAccountsPage extends React.Component<Props> {
    private subscriptions = new Subscription()
    private externalAccountUpdates = new Subject<void>()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminExternalAccounts')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<ExternalAccountNodeProps, 'onDidUpdate' | 'showUser'> = {
            onDidUpdate: this.onDidUpdateExternalAccount,
            showUser: true,
        }

        return (
            <div className="user-settings-external-accounts-page">
                <PageTitle title="External accounts" />
                <h2>External user accounts</h2>
                <p>
                    An external account (on an <Link to="/site-admin/auth/providers">authentication provider</Link>) is
                    linked to a Sourcegraph user when it's used to sign into Sourcegraph.
                </p>
                <div>
                    <Link to="/site-admin/auth/providers" className="btn btn-secondary">
                        View auth providers
                    </Link>
                </div>
                <FilteredExternalAccountConnection
                    className="list-group list-group-flush mt-3"
                    noun="external user account"
                    pluralNoun="external user accounts"
                    queryConnection={this.queryExternalAccounts}
                    nodeComponent={ExternalAccountNode}
                    nodeComponentProps={nodeProps}
                    updates={this.externalAccountUpdates}
                    hideSearch={true}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryExternalAccounts = (
        args: {
            first?: number
        } & FilterParams
    ): Observable<GQL.IExternalAccountConnection> =>
        queryGraphQL(
            gql`
                query ExternalAccounts(
                    $first: Int
                    $user: ID
                    $serviceType: String
                    $serviceID: String
                    $clientID: String
                ) {
                    site {
                        externalAccounts(
                            first: $first
                            user: $user
                            serviceType: $serviceType
                            serviceID: $serviceID
                            clientID: $clientID
                        ) {
                            nodes {
                                ...ExternalAccountFields
                            }
                            totalCount
                            pageInfo {
                                hasNextPage
                            }
                        }
                    }
                }
                ${externalAccountFragment}
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.site || !data.site.externalAccounts) {
                    throw createAggregateError(errors)
                }
                return data.site.externalAccounts
            })
        )

    private onDidUpdateExternalAccount = () => this.externalAccountUpdates.next()
}
