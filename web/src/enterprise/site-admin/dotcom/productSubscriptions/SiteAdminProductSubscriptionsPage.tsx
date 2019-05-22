import AddIcon from 'mdi-react/AddIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { Tab, TabsWithLocalStorageViewStatePersistence } from '../../../../../../shared/src/components/Tabs'
import { gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../../../shared/src/util/errors'
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

type SubscriptionsDisplays = 'by-created-at' | 'by-expires-at'

interface State {
    tab: SubscriptionsDisplays
}

/**
 * Displays the product subscriptions that have been created on Sourcegraph.com.
 */
export class SiteAdminProductSubscriptionsPage extends React.Component<Props, State> {
    public state: State = { tab: 'by-created-at' }
    private static TABS: Tab<SubscriptionsDisplays>[] = [
        { id: 'by-created-at', label: 'Sort by latest created' },
        { id: 'by-expires-at', label: 'Sort by license expiration' },
    ]
    private static LAST_TAB_STORAGE_KEY = 'site-admin-product-subscriptions-last-tab'

    private subscriptions = new Subscription()
    private updates = new Subject<void>()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminProductSubscriptions')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private tabSelected = (tab: SubscriptionsDisplays) => this.setState({ tab })

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
                <TabsWithLocalStorageViewStatePersistence
                    tabs={SiteAdminProductSubscriptionsPage.TABS}
                    storageKey={SiteAdminProductSubscriptionsPage.LAST_TAB_STORAGE_KEY}
                    tabClassName="tab-bar__tab--h5like"
                    onSelectTab={this.tabSelected}
                >
                    <FilteredSiteAdminProductSubscriptionConnection
                        key="by-created-at"
                        className="mt-3"
                        listComponent="table"
                        listClassName="table"
                        noun="product subscription"
                        pluralNoun="product subscriptions"
                        queryConnection={this.queryProductSubscriptionsByCreatedAt}
                        headComponent={SiteAdminProductSubscriptionNodeHeader}
                        nodeComponent={SiteAdminProductSubscriptionNode}
                        nodeComponentProps={nodeProps}
                        hideSearch={true}
                        updates={this.updates}
                        history={this.props.history}
                        location={this.props.location}
                    />
                    <FilteredSiteAdminProductSubscriptionConnection
                        key="by-expires-at"
                        className="mt-3"
                        listComponent="table"
                        listClassName="table"
                        noun="product subscription"
                        pluralNoun="product subscriptions"
                        queryConnection={this.queryProductSubscriptionsByExpiresAt}
                        headComponent={SiteAdminProductSubscriptionNodeHeader}
                        nodeComponent={SiteAdminProductSubscriptionNode}
                        nodeComponentProps={nodeProps}
                        hideSearch={true}
                        updates={this.updates}
                        history={this.props.history}
                        location={this.props.location}
                    />
                </TabsWithLocalStorageViewStatePersistence>
            </div>
        )
    }

    private queryProductSubscriptions = (orderBy: GQL.SubscriptionOrderBy) => (args: {
        first?: number
    }): Observable<GQL.IProductSubscriptionConnection> =>
        queryGraphQL(
            gql`
                query ProductSubscriptions($first: Int, $account: ID, $orderBy: SubscriptionOrderBy) {
                    dotcom {
                        productSubscriptions(first: $first, account: $account, orderBy: $orderBy) {
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
                orderBy,
            } as GQL.IProductSubscriptionsOnDotcomQueryArguments
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.dotcom || !data.dotcom.productSubscriptions || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
                return data.dotcom.productSubscriptions
            })
        )
    private queryProductSubscriptionsByCreatedAt = this.queryProductSubscriptions(
        GQL.SubscriptionOrderBy.SUBSCRIPTION_CREATED_AT
    )
    private queryProductSubscriptionsByExpiresAt = this.queryProductSubscriptions(
        GQL.SubscriptionOrderBy.SUBSCRIPTION_ACTIVE_LICENSE_EXPIRES_AT
    )

    private onDidUpdateProductSubscription = () => this.updates.next()
}
