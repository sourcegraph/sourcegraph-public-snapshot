import React, { useEffect } from 'react'

import { mdiPlus } from '@mdi/js'

import { Button, Link, Icon, PageHeader, Container } from '@sourcegraph/wildcard'

import { FilteredConnection } from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import type { SiteAdminProductSubscriptionFields } from '../../../../graphql-operations'
import { eventLogger } from '../../../../tracking/eventLogger'

import { queryProductSubscriptions } from './backend'
import {
    SiteAdminProductSubscriptionNode,
    SiteAdminProductSubscriptionNodeHeader,
    type SiteAdminProductSubscriptionNodeProps,
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
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Product subscriptions' }]}
                actions={
                    <Button to="/site-admin/dotcom/product/subscriptions/new" variant="primary" as={Link}>
                        <Icon aria-hidden={true} svgPath={mdiPlus} />
                        Create product subscription
                    </Button>
                }
                className="mb-3"
            />

            <Container>
                <FilteredConnection<SiteAdminProductSubscriptionFields, SiteAdminProductSubscriptionNodeProps>
                    listComponent="table"
                    listClassName="table"
                    noun="product subscription"
                    pluralNoun="product subscriptions"
                    queryConnection={queryProductSubscriptions}
                    headComponent={SiteAdminProductSubscriptionNodeHeader}
                    nodeComponent={SiteAdminProductSubscriptionNode}
                />
            </Container>
        </div>
    )
}
