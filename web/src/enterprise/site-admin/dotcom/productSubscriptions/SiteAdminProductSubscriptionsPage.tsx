import AddIcon from 'mdi-react/AddIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { queryGraphQL } from '../../../../backend/graphql'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { eventLogger } from '../../../../tracking/eventLogger'
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
                <div className="d-flex justify-content-between align-items-center mt-3 mb-3">
                    <h2 className="mb-0">Product subscriptions</h2>
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
                    updates={this.updates}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryProductSubscriptions = (args: {
        first?: number
        query?: string
    }): Observable<GQL.IProductSubscriptionConnection> =>
        queryGraphQL(
            gql`
                query ProductSubscriptions($first: Int, $account: ID, $query: String) {
                    dotcom {
                        productSubscriptions(first: $first, account: $account, query: $query) {
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
                query: args.query,
            } as GQL.IProductSubscriptionsOnDotcomQueryArguments
        ).pipe(
            map(dataOrThrowErrors),
            map(data => data.dotcom.productSubscriptions)
        )

    private onDidUpdateProductSubscription = (): void => this.updates.next()
}
