import React, { useEffect } from 'react'

import { mdiPlus } from '@mdi/js'

import { Button, Link, Icon, PageHeader, Container, Tooltip } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import type { SiteAdminProductSubscriptionFields } from '../../../../graphql-operations'
import { canWriteLicenseManagement } from '../../../../rbac/check'
import { eventLogger } from '../../../../tracking/eventLogger'

import { queryProductSubscriptions } from './backend'
import {
    SiteAdminProductSubscriptionNode,
    SiteAdminProductSubscriptionNodeHeader,
    type SiteAdminProductSubscriptionNodeProps,
} from './SiteAdminProductSubscriptionNode'

import styles from './SiteAdminCreateProductSubscriptionPage.module.scss'

interface Props {
    authenticatedUser: AuthenticatedUser
}

/**
 * Displays the product subscriptions that have been created on Sourcegraph.com.
 */
export const SiteAdminProductSubscriptionsPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    authenticatedUser,
}) => {
    useEffect(() => eventLogger.logViewEvent('SiteAdminProductSubscriptions'), [])

    return (
        <div className="site-admin-product-subscriptions-page">
            <PageTitle title="Product subscriptions" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Product subscriptions' }]}
                actions={
                    <Tooltip content={!canWriteLicenseManagement(authenticatedUser) ? 'No permissions' : undefined}>
                        <Button
                            disabled={!canWriteLicenseManagement(authenticatedUser)}
                            to="./new"
                            variant="primary"
                            as={Link}
                        >
                            <Icon aria-hidden={true} svgPath={mdiPlus} />
                            Create product subscription
                        </Button>
                    </Tooltip>
                }
                className="mb-3"
            />

            <Container>
                <FilteredConnection<SiteAdminProductSubscriptionFields, SiteAdminProductSubscriptionNodeProps>
                    listComponent="table"
                    listClassName="table"
                    contentWrapperComponent={ListContentWrapper}
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

const ListContentWrapper: React.FunctionComponent<React.PropsWithChildren<{}>> = ({ children }) => (
    <div className={styles.contentWrapper}>{children}</div>
)
