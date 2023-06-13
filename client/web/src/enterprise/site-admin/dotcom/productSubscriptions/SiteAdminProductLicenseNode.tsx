import React, { useCallback } from 'react'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { useMutation } from '@sourcegraph/http-client'
import { Text, LinkOrSpan, Label, Alert, H3 } from '@sourcegraph/wildcard'

import { CopyableText } from '../../../../components/CopyableText'
import { LoaderButton } from '../../../../components/LoaderButton'
import { ProductLicenseFields, RevokeLicenseResult, RevokeLicenseVariables } from '../../../../graphql-operations'
import { isProductLicenseExpired } from '../../../../productSubscription/helpers'
import { AccountName } from '../../../dotcom/productSubscriptions/AccountName'
import { ProductLicenseValidity } from '../../../dotcom/productSubscriptions/ProductLicenseValidity'
import { ProductLicenseInfoDescription } from '../../../productSubscription/ProductLicenseInfoDescription'
import { hasUnknownTags, ProductLicenseTags, UnknownTagWarning } from '../../../productSubscription/ProductLicenseTags'

import { REVOKE_LICENSE } from './backend'

export interface SiteAdminProductLicenseNodeProps {
    node: ProductLicenseFields
    showSubscription: boolean
    onRevokeCompleted: () => void
}

/**
 * Displays a product license in a connection in the site admin area.
 */
export const SiteAdminProductLicenseNode: React.FunctionComponent<
    React.PropsWithChildren<SiteAdminProductLicenseNodeProps>
> = ({ node, showSubscription, onRevokeCompleted }) => {
    const [revoke, { loading, error }] = useMutation<RevokeLicenseResult, RevokeLicenseVariables>(REVOKE_LICENSE)

    const onRevoke = useCallback(() => {
        const reason = window.prompt('Reason for revoking the license key:')
        if (reason) {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            revoke({
                variables: {
                    id: node.id,
                    reason,
                },
                onCompleted: data => {
                    if (onRevokeCompleted) {
                        onRevokeCompleted()
                    }
                },
            })
        }
    }, [revoke, node, onRevokeCompleted])

    return (
        <li className="list-group-item py-3">
            {showSubscription && (
                <div className="text-truncate d-flex mb-2">
                    <H3>
                        License in{' '}
                        <LinkOrSpan to={node.subscription.urlForSiteAdmin} className="mr-3">
                            {node.subscription.name}
                        </LinkOrSpan>
                    </H3>
                    <span className="mr-3">
                        <AccountName account={node.subscription.account} />
                    </span>
                </div>
            )}
            {!loading && error && <Alert variant="danger">Error revoking license: {error.message}</Alert>}
            <div className="d-flex justify-content-baseline mb-2">
                {node.info && <ProductLicenseInfoDescription licenseInfo={node.info} className="mb-0" />}
                <Text className="ml-2 mb-0">
                    <small className="text-muted">
                        Created <Timestamp date={node.createdAt} />
                    </small>
                </Text>
                <Text className="ml-3 mb-0 text-muted">
                    <small>Version {node.version}</small>
                </Text>
                {!node?.revokedAt && !isProductLicenseExpired(node?.info?.expiresAt ?? 0) && (
                    <LoaderButton
                        variant="danger"
                        className="ml-auto"
                        label="Revoke"
                        onClick={onRevoke}
                        loading={loading}
                    />
                )}
            </div>
            <ProductLicenseValidity license={node} className="mb-2" />
            {node.version > 1 && (
                <>
                    <div className="d-flex">
                        <Label>Site ID</Label>
                        <Text className="ml-3">{node.siteID ?? <span className="text-muted">Unused</span>}</Text>
                    </div>
                    <div className="d-flex">
                        <Label>Salesforce Subscription ID</Label>
                        <Text className="ml-3">
                            {node.info?.salesforceSubscriptionID ?? <span className="text-muted">Unused</span>}
                        </Text>
                    </div>
                    <div className="d-flex">
                        <Label>Salesforce Opportunity ID</Label>
                        <Text className="ml-3">
                            {node.info?.salesforceOpportunityID ?? <span className="text-muted">Unused</span>}
                        </Text>
                    </div>
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
}
