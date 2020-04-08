import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'
import { FilteredConnection } from '../../../components/FilteredConnection'
import { PageTitle } from '../../../components/PageTitle'
import { eventLogger } from '../../../tracking/eventLogger'
import {
    productSubscriptionFragment,
    ProductSubscriptionNode,
    ProductSubscriptionNodeHeader,
    ProductSubscriptionNodeProps,
} from '../../dotcom/productSubscriptions/ProductSubscriptionNode'

interface Props extends RouteComponentProps<{}> {
    user: GQL.IUser
}

class FilteredProductSubscriptionConnection extends FilteredConnection<
    GQL.IProductSubscription,
    Pick<ProductSubscriptionNodeProps, 'onDidUpdate'>
> {}

/**
 * Displays the product subscriptions associated with this account.
 */
export class UserSubscriptionsProductSubscriptionsPage extends React.Component<Props> {
    private subscriptions = new Subscription()
    private updates = new Subject<void>()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSubscriptionsProductSubscriptions')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<ProductSubscriptionNodeProps, 'onDidUpdate'> = {
            onDidUpdate: this.onDidUpdateProductSubscription,
        }

        return (
            <div className="user-subscriptions-product-subscriptions-page">
                <PageTitle title="Subscriptions" />
                <div className="d-flex justify-content-between align-items-center mb-3">
                    <h2 className="mb-0">Subscriptions</h2>
                    <Link to={`${this.props.match.path}/new`} className="btn btn-primary">
                        New subscription
                    </Link>
                </div>
                <p>
                    A subscription gives you a license key to run a self-hosted Sourcegraph instance. See{' '}
                    <a href="https://about.sourcegraph.com/pricing">Sourcegraph pricing</a> for more information.
                </p>
                <FilteredProductSubscriptionConnection
                    className="mt-3"
                    listComponent="table"
                    listClassName="table"
                    noun="subscription"
                    pluralNoun="subscriptions"
                    queryConnection={this.queryLicenses}
                    headComponent={ProductSubscriptionNodeHeader}
                    nodeComponent={ProductSubscriptionNode}
                    nodeComponentProps={nodeProps}
                    updates={this.updates}
                    hideSearch={true}
                    noSummaryIfAllNodesVisible={true}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryLicenses = (args: { first?: number }): Observable<GQL.IProductSubscriptionConnection> => {
        const vars: GQL.IProductSubscriptionsOnDotcomQueryArguments = {
            first: args.first,
            account: this.props.user.id,
        }
        return queryGraphQL(
            gql`
                query ProductSubscriptions($first: Int, $account: ID) {
                    dotcom {
                        productSubscriptions(first: $first, account: $account) {
                            nodes {
                                ...ProductSubscriptionFields
                            }
                            totalCount
                            pageInfo {
                                hasNextPage
                            }
                        }
                    }
                }
                ${productSubscriptionFragment}
            `,
            vars
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.dotcom || !data.dotcom.productSubscriptions || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
                return data.dotcom.productSubscriptions
            })
        )
    }

    private onDidUpdateProductSubscription = (): void => this.updates.next()
}
