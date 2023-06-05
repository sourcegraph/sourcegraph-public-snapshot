import React, { useEffect } from 'react'

import { H2, Text } from '@sourcegraph/wildcard'

import { FilteredConnection } from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { ProductLicenseFields } from '../../../../graphql-operations'
import { eventLogger } from '../../../../tracking/eventLogger'

import { queryLicenses } from './backend'
import { SiteAdminProductLicenseNode, SiteAdminProductLicenseNodeProps } from './SiteAdminProductLicenseNode'

interface Props {}

/**
 * Displays the product licenses that have been created on Sourcegraph.com.
 */
export const SiteAdminProductLicensesPage: React.FunctionComponent<React.PropsWithChildren<Props>> = () => {
    useEffect(() => eventLogger.logViewEvent('SiteAdminProductLicenses'), [])

    const nodeProps: Pick<SiteAdminProductLicenseNodeProps, 'showSubscription'> = {
        showSubscription: true,
    }

    return (
        <div className="site-admin-product-subscriptions-page">
            <PageTitle title="Product subscriptions" />
            <H2>License key lookup</H2>
            <Text>Find matching licenses and their associated product subscriptions.</Text>
            <FilteredConnection<ProductLicenseFields, Pick<SiteAdminProductLicenseNodeProps, 'showSubscription'>>
                className="list-group list-group-flush mt-3"
                noun="product license"
                pluralNoun="product licenses"
                queryConnection={queryLicenses}
                nodeComponent={SiteAdminProductLicenseNode}
                nodeComponentProps={nodeProps}
                emptyElement={<span className="text-muted mt-2">Enter a partial license key to find matches.</span>}
                noSummaryIfAllNodesVisible={true}
                autoFocus={true}
            />
        </div>
    )
}
