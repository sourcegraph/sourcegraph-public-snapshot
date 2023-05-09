import React from 'react'

import { mdiShieldOff } from '@mdi/js'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { Tooltip, Text, LinkOrSpan, Icon, H4 } from '@sourcegraph/wildcard'

import { CopyableText } from '../../../../components/CopyableText'
import { ProductLicenseFields } from '../../../../graphql-operations'
import { AccountName } from '../../../dotcom/productSubscriptions/AccountName'
import { ProductLicenseValidity } from '../../../dotcom/productSubscriptions/ProductLicenseValidity'
import { ProductLicenseInfoDescription } from '../../../productSubscription/ProductLicenseInfoDescription'
import { hasUnknownTags, ProductLicenseTags, UnknownTagWarning } from '../../../productSubscription/ProductLicenseTags'

export interface SiteAdminProductLicenseNodeProps {
    node: ProductLicenseFields
    showSubscription: boolean
}

/**
 * Displays a product license in a connection in the site admin area.
 */
export const SiteAdminProductLicenseNode: React.FunctionComponent<
    React.PropsWithChildren<SiteAdminProductLicenseNodeProps>
> = ({ node, showSubscription }) => (
    <li className="list-group-item py-2">
        {showSubscription && (
            <div className="mr-3 text-truncate">
                <strong>
                    License in{' '}
                    <LinkOrSpan to={node.subscription.urlForSiteAdmin} className="mr-3">
                        {node.subscription.name}
                    </LinkOrSpan>
                </strong>
                <span className="mr-3">
                    <AccountName account={node.subscription.account} />
                </span>
            </div>
        )}
        <div className="d-flex justify-content-baseline mb-2">
            {node.info && <ProductLicenseInfoDescription licenseInfo={node.info} className="mb-0" />}
            <Text className="ml-2 mb-0">
                <small className="text-muted">
                    Created <Timestamp date={node.createdAt} />
                </small>
            </Text>
        </div>
        {node.info && node.subscription.activeLicense && node.subscription.activeLicense.id === node.id ? (
            <ProductLicenseValidity licenseInfo={node.info} className="mb-2" />
        ) : (
            <Tooltip content="A newer license was generated for this subscription. This license should no longer be used.">
                <Text className="mb-2">
                    <Icon svgPath={mdiShieldOff} aria-hidden={true} className="text-warning mr-1" />
                    <strong>Inactive</strong>
                </Text>
            </Tooltip>
        )}
        {node.info && node.info.tags.length > 0 && (
            <>
                {hasUnknownTags(node.info.tags) && <UnknownTagWarning className="mb-2" />}
                <H4>Tags</H4>
                <Text className="mb-2">
                    <ProductLicenseTags tags={node.info.tags} />
                </Text>
            </>
        )}
        <CopyableText flex={true} text={node.licenseKey} />
    </li>
)
