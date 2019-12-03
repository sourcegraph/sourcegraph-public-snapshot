import * as H from 'history'
import AddIcon from 'mdi-react/AddIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { map, tap } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { mutateGraphQL, queryGraphQL } from '../../../../backend/graphql'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { eventLogger } from '../../../../tracking/eventLogger'

interface UserCreateSubscriptionNodeProps {
    /**
     * The user to display in this list item.
     */
    node: GQL.IUser

    /**
     * Browser history, used to redirect the user to the new subscription after one is successfully created.
     */
    history: H.History
}

class UserCreateSubscriptionNode extends React.PureComponent<UserCreateSubscriptionNodeProps> {
    private createProductSubscription = (): Observable<
        Pick<GQL.IProductSubscription, 'id' | 'name' | 'url' | 'urlForSiteAdmin'>
    > =>
        createProductSubscription({ accountID: this.props.node.id }).pipe(
            tap(({ url, urlForSiteAdmin }) => this.props.history.push(urlForSiteAdmin || url))
        )

    public render(): JSX.Element | null {
        return (
            <li className="list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    <div>
                        <Link to={`/users/${this.props.node.username}`}>{this.props.node.username}</Link>{' '}
                        <span className="text-muted">
                            ({this.props.node.emails.filter(email => email.isPrimary).map(email => email.email)})
                        </span>
                    </div>
                    <div>
                        <button
                            type="button"
                            className="btn btn-sm btn-secondary"
                            onClick={this.createProductSubscription}
                        >
                            <AddIcon className="icon-inline" /> Create new subscription
                        </button>
                    </div>
                </div>
            </li>
        )
    }
}

class FilteredUserConnection extends FilteredConnection<GQL.IUser, Pick<UserCreateSubscriptionNodeProps, 'history'>> {}

interface Props extends RouteComponentProps<{}> {
    authenticatedUser: GQL.IUser
}

/**
 * Creates a product subscription for an account based on information provided in the displayed form.
 *
 * For use on Sourcegraph.com by Sourcegraph teammates only.
 */
export class SiteAdminCreateProductSubscriptionPage extends React.Component<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminCreateProductSubscription')
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<UserCreateSubscriptionNodeProps, 'history'> = {
            history: this.props.history,
        }
        return (
            <div className="site-admin-create-product-subscription-page">
                <PageTitle title="Create product subscription" />
                <h2>Create product subscription</h2>
                <FilteredUserConnection
                    className="list-group list-group-flush mt-3"
                    noun="user"
                    pluralNoun="users"
                    queryConnection={queryAccounts}
                    nodeComponent={UserCreateSubscriptionNode}
                    nodeComponentProps={nodeProps}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }
}

function queryAccounts(args: { first?: number; query?: string }): Observable<GQL.IUserConnection> {
    return queryGraphQL(
        gql`
            query ProductSubscriptionAccounts($first: Int, $query: String) {
                users(first: $first, query: $query) {
                    nodes {
                        id
                        username
                        emails {
                            email
                            verified
                            isPrimary
                        }
                    }
                }
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.users)
    )
}

function createProductSubscription(
    args: GQL.ICreateProductSubscriptionOnDotcomMutationArguments
): Observable<Pick<GQL.IProductSubscription, 'id' | 'name' | 'url' | 'urlForSiteAdmin'>> {
    return mutateGraphQL(
        gql`
            mutation CreateProductSubscription($accountID: ID!) {
                dotcom {
                    createProductSubscription(accountID: $accountID) {
                        id
                        name
                        urlForSiteAdmin
                    }
                }
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.dotcom.createProductSubscription)
    )
}
