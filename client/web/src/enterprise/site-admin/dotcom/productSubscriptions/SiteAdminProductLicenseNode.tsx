import React from 'react'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { Text, LinkOrSpan, Label } from '@sourcegraph/wildcard'

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
    <li className="list-group-item py-3">
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
            <Text className="ml-auto mb-0">
                <small>Version {node.version}</small>
            </Text>
        </div>
        <ProductLicenseValidity license={node} className="mb-2" />
        {node.version > 1 && (
            <>
                <Label className="mb-2">
                    <Text className="mb-0">Site ID</Text>
                </Label>
                <Text className="mb-3 w-100">{node.siteID ?? <span className="text-muted">Unused</span>}</Text>
            </>
        )}
        {node.info && node.info.tags.length > 0 && (
            <>
                {hasUnknownTags(node.info.tags) && <UnknownTagWarning className="mb-2" />}
                <Label className="w-100">
                    <Text className="mb-2">Tags</Text>
                    <Text className="mb-2">
                        <ProductLicenseTags tags={node.info.tags} />
                    </Text>
                </Label>
            </>
        )}
        <Label className="w-100">
            <Text className="mb-2">License Key</Text>
            <CopyableText flex={true} text={node.licenseKey} />
        </Label>
    </li>
)
