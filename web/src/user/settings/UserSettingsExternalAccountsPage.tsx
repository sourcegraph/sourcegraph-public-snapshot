import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { FilteredConnection } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { createAggregateError } from '../../util/errors'
import { externalAccountFragment, ExternalAccountNode, ExternalAccountNodeProps } from './ExternalAccountNode'

interface Props extends RouteComponentProps<{}> {
    user: GQL.IUser
}

/** We fake a XyzConnection type because our GraphQL API doesn't have one (or need one) for external accounts. */
interface ExternalAccountConnection {
    nodes: GQL.IExternalAccount[]
    totalCount: number
}

class FilteredExternalAccountConnection extends FilteredConnection<
    GQL.IExternalAccount,
    Pick<ExternalAccountNodeProps, 'onDidUpdate'>
> {}

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
        const nodeProps: Pick<ExternalAccountNodeProps, 'onDidUpdate'> = {
            onDidUpdate: this.onDidUpdateExternalAccount,
        }

        return (
            <div className="user-settings-external-accounts-page">
                <PageTitle title="Connected accounts" />
                <h2>Connected accounts</h2>
                <FilteredExternalAccountConnection
                    className="list-group list-group-flush mt-3"
                    noun="connected account"
                    pluralNoun="connected accounts"
                    queryConnection={this.queryUserExternalAccounts}
                    nodeComponent={ExternalAccountNode}
                    nodeComponentProps={nodeProps}
                    updates={this.externalAccountUpdates}
                    hideFilter={true}
                    noSummaryIfAllNodesVisible={true}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryUserExternalAccounts = (args: {}): Observable<ExternalAccountConnection> =>
        queryGraphQL(
            gql`
                query UserExternalAccounts($user: ID!) {
                    node(id: $user) {
                        ... on User {
                            externalAccounts {
                                ...ExternalAccountFields
                            }
                        }
                    }
                }
                ${externalAccountFragment}
            `,
            { user: this.props.user.id }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node) {
                    throw createAggregateError(errors)
                }
                const user = data.node as GQL.IUser
                if (!user.externalAccounts) {
                    throw createAggregateError(errors)
                }
                return {
                    nodes: user.externalAccounts,
                    totalCount: user.externalAccounts.length,
                }
            })
        )

    private onDidUpdateExternalAccount = () => this.externalAccountUpdates.next()
}
