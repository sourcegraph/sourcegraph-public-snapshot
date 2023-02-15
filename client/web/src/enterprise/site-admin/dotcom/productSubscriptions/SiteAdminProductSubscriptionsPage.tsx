import React, { useEffect } from 'react'

import { mdiPlus } from '@mdi/js'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { Button, Link, Icon, H2 } from '@sourcegraph/wildcard'

import { queryGraphQL } from '../../../../backend/graphql'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import {
    ProductSubscriptionsDotComResult,
    ProductSubscriptionsDotComVariables,
    SiteAdminProductSubscriptionFields,
} from '../../../../graphql-operations'
import { eventLogger } from '../../../../tracking/eventLogger'

import {
    siteAdminProductSubscriptionFragment,
    SiteAdminProductSubscriptionNode,
    SiteAdminProductSubscriptionNodeHeader,
    SiteAdminProductSubscriptionNodeProps,
} from './SiteAdminProductSubscriptionNode'

interface Props {}

/**
 * Displays the product subscriptions that have been created on Sourcegraph.com.
 */
export const SiteAdminProductSubscriptionsPage: React.FunctionComponent<React.PropsWithChildren<Props>> = () => {
    useEffect(() => eventLogger.logViewEvent('SiteAdminProductSubscriptions'), [])
    return (
        <div className="site-admin-product-subscriptions-page">
            <PageTitle title="Product subscriptions" />
            <div className="d-flex justify-content-between align-items-center mb-3">
                <H2 className="mb-0">Product subscriptions</H2>
                <Button to="/site-admin/dotcom/product/subscriptions/new" variant="primary" as={Link}>
                    <Icon aria-hidden={true} svgPath={mdiPlus} />
                    Create product subscription
                </Button>
            </div>
            <FilteredConnection<SiteAdminProductSubscriptionFields, SiteAdminProductSubscriptionNodeProps>
                className="mt-3"
                listComponent="table"
                listClassName="table"
                noun="product subscription"
                pluralNoun="product subscriptions"
                queryConnection={queryProductSubscriptions}
                headComponent={SiteAdminProductSubscriptionNodeHeader}
                nodeComponent={SiteAdminProductSubscriptionNode}
            />
        </div>
    )
}

function queryProductSubscriptions(args: {
    first?: number
    query?: string
}): Observable<ProductSubscriptionsDotComResult['dotcom']['productSubscriptions']> {
    return queryGraphQL<ProductSubscriptionsDotComResult>(
        gql`
            query ProductSubscriptionsDotCom($first: Int, $account: ID, $query: String) {
                dotcom {
                    productSubscriptions(first: $first, account: $account, query: $query) {
                        nodes {
                            ...SiteAdminProductSubscriptionFields
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
        } as Partial<ProductSubscriptionsDotComVariables>
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.dotcom.productSubscriptions)
    )
}
