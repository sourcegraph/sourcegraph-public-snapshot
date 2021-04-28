import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'

import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import { createAggregateError } from '@sourcegraph/shared/src/util/errors'

import { requestGraphQL } from '../../../backend/graphql'
import { FilteredConnection } from '../../../components/FilteredConnection'
import { PageTitle } from '../../../components/PageTitle'
import {
    UserAreaUserFields,
    ExternalAccountFields,
    ExternalAccountsConnectionFields,
    UserExternalAccountsResult,
    UserExternalAccountsVariables,
} from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'

import {
    ExternalAccountNode,
    ExternalAccountNodeProps,
    externalAccountsConnectionFragment,
} from './ExternalAccountNode'

interface Props extends RouteComponentProps<{}> {
    user: UserAreaUserFields
}

/**
 * Displays the external accounts (from authentication providers) associated with the user's account.
 */
export class UserSettingsExternalAccountsPage extends React.Component<Props> {
    private subscriptions = new Subscription()
    private externalAccountUpdates = new Subject<void>()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSettingsExternalAccounts')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nodeProps: Omit<ExternalAccountNodeProps, 'node'> = {
            onDidUpdate: this.onDidUpdateExternalAccount,
            showUser: false,
        }

        return (
            <div className="user-settings-external-accounts-page">
                <PageTitle title="External accounts" />
                <h2>External accounts</h2>
                <FilteredConnection<ExternalAccountFields, Omit<ExternalAccountNodeProps, 'node'>>
                    className="list-group list-group-flush mt-3"
                    noun="external account"
                    pluralNoun="external accounts"
                    queryConnection={this.queryUserExternalAccounts}
                    nodeComponent={ExternalAccountNode}
                    nodeComponentProps={nodeProps}
                    updates={this.externalAccountUpdates}
                    hideSearch={true}
                    noSummaryIfAllNodesVisible={true}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryUserExternalAccounts = (args: { first?: number }): Observable<ExternalAccountsConnectionFields> =>
        requestGraphQL<UserExternalAccountsResult, UserExternalAccountsVariables>(
            gql`
                query UserExternalAccounts($user: ID!, $first: Int) {
                    node(id: $user) {
                        ... on User {
                            externalAccounts(first: $first) {
                                ...ExternalAccountsConnectionFields
                            }
                        }
                    }
                }
                ${externalAccountsConnectionFragment}
            `,
            { user: this.props.user.id, first: args.first ?? null }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node) {
                    throw createAggregateError(errors)
                }
                const user = data.node
                if (!user.externalAccounts) {
                    throw createAggregateError(errors)
                }
                return user.externalAccounts
            })
        )

    private onDidUpdateExternalAccount = (): void => this.externalAccountUpdates.next()
}
