import { gql, queryGraphQL } from '@sourcegraph/webapp/dist/backend/graphql'
import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import { FilteredConnection } from '@sourcegraph/webapp/dist/components/FilteredConnection'
import { PageTitle } from '@sourcegraph/webapp/dist/components/PageTitle'
import { eventLogger } from '@sourcegraph/webapp/dist/tracking/eventLogger'
import { createAggregateError } from '@sourcegraph/webapp/dist/util/errors'
import AddIcon from 'mdi-react/AddIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import {
    siteAdminProductSubscriptionFragment,
    SiteAdminProductSubscriptionNode,
    SiteAdminProductSubscriptionNodeHeader,
    SiteAdminProductSubscriptionNodeProps,
} from './SiteAdminProductSubscriptionNode'

interface Props extends RouteComponentProps<{}> {}

class FilteredSiteAdminProductSubscriptionConnection extends FilteredConnection<
    GQL.IProductSubscription,
    Pick<SiteAdminProductSubscriptionNodeProps, 'onDidUpdate'>
> {}

/**
 * Displays the product subscriptions that have been created on Sourcegraph.com.
 */
export class SiteAdminProductSubscriptionsPage extends React.Component<Props> {
    private subscriptions = new Subscription()
    private updates = new Subject<void>()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminProductSubscriptions')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<SiteAdminProductSubscriptionNodeProps, 'onDidUpdate'> = {
            onDidUpdate: this.onDidUpdateProductSubscription,
        }

        return (
            <div className="site-admin-product-subscriptions-page">
                <PageTitle title="Product subscriptions" />
                <h2>Product subscriptions</h2>
                <div>
                    <Link to="/site-admin/dotcom/product/subscriptions/new" className="btn btn-primary">
                        <AddIcon className="icon-inline" />
                        Create product subscription
                    </Link>
                </div>
                <FilteredSiteAdminProductSubscriptionConnection
                    className="mt-3"
                    listComponent="table"
                    listClassName="table"
                    noun="product subscription"
                    pluralNoun="product subscriptions"
                    queryConnection={this.queryProductSubscriptions}
                    headComponent={SiteAdminProductSubscriptionNodeHeader}
                    nodeComponent={SiteAdminProductSubscriptionNode}
                    nodeComponentProps={nodeProps}
                    hideSearch={true}
                    updates={this.updates}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryProductSubscriptions = (args: { first?: number }): Observable<GQL.IProductSubscriptionConnection> =>
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
                ${siteAdminProductSubscriptionFragment}
            `,
            {
                first: args.first,
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
