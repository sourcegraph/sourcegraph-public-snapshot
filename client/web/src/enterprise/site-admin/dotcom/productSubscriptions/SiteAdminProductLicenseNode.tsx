import * as React from 'react'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { gql } from '@sourcegraph/http-client'
import { Tooltip, LinkOrSpan } from '@sourcegraph/wildcard'

import { CopyableText } from '../../../../components/CopyableText'
import { ProductLicenseFields } from '../../../../graphql-operations'
import { AccountName } from '../../../dotcom/productSubscriptions/AccountName'
import { ProductLicenseValidity } from '../../../dotcom/productSubscriptions/ProductLicenseValidity'
import { ProductLicenseInfoDescription } from '../../../productSubscription/ProductLicenseInfoDescription'
import { hasUnknownTags, ProductLicenseTags, UnknownTagWarning } from '../../../productSubscription/ProductLicenseTags'

export const siteAdminProductLicenseFragment = gql`
    fragment ProductLicenseFields on ProductLicense {
        id
        subscription {
            id
            name
            account {
                ...ProductLicenseSubscriptionAccount
            }
            activeLicense {
                id
            }
            urlForSiteAdmin
        }
        licenseKey
        info {
            ...ProductLicenseInfoFields
        }
        createdAt
    }

    fragment ProductLicenseInfoFields on ProductLicenseInfo {
        productNameWithBrand
        tags
        userCount
        expiresAt
    }

    fragment ProductLicenseSubscriptionAccount on User {
        id
        username
        displayName
    }
`

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
        <div className="d-flex align-items-center justify-content-between">
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
            <div>
                {node.info && <ProductLicenseInfoDescription licenseInfo={node.info} className="mr-3" />}
                {node.info && node.subscription.activeLicense && node.subscription.activeLicense.id === node.id ? (
                    <ProductLicenseValidity licenseInfo={node.info} primary={false} className="d-inline-block mr-3" />
                ) : (
                    <Tooltip content="A newer license was generated for this subscription. This license should no longer be used.">
                        <span className="text-warning font-weight-bold mr-3">Inactive</span>
                    </Tooltip>
                )}
                <span className="text-muted">
                    Created <Timestamp date={node.createdAt} />
                </span>
            </div>
        </div>
        {node.info && node.info.tags.length > 0 && (
            <>
                {hasUnknownTags(node.info.tags) && <UnknownTagWarning />}
                <div>
                    Tags: <ProductLicenseTags tags={node.info.tags} />
                </div>
            </>
        )}
        <CopyableText flex={true} text={node.licenseKey} className="mt-2" />
    </li>
)
