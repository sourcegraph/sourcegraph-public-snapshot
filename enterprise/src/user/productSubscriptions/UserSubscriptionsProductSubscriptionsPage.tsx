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
                <h2>Subscriptions</h2>
                <div>
                    <Link to={`${this.props.match.path}/new`} className="btn btn-primary">
                        New subscription
                    </Link>
                </div>
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

    private queryLicenses = (args: { first?: number }): Observable<GQL.IProductSubscriptionConnection> =>
        queryGraphQL(
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
            {
                first: args.first,
                account: this.props.user.id,
            } as GQL.IProductSubscriptionsOnDotcomQueryArguments
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.dotcom || !data.dotcom.productSubscriptions || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
                return data.dotcom.productSubscriptions
            })
        )

    private onDidUpdateProductSubscription = () => this.updates.next()
}
