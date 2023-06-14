import React, { useCallback, useState } from 'react'

import { mdiChevronUp, mdiChevronDown } from '@mdi/js'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { useMutation } from '@sourcegraph/http-client'
import {
    Text,
    LinkOrSpan,
    Label,
    Alert,
    H3,
    Grid,
    Button,
    Icon,
    Collapse,
    CollapseHeader,
    CollapsePanel,
} from '@sourcegraph/wildcard'

import { CopyableText } from '../../../../components/CopyableText'
import { LoaderButton } from '../../../../components/LoaderButton'
import { ProductLicenseFields, RevokeLicenseResult, RevokeLicenseVariables } from '../../../../graphql-operations'
import { formatUserCount, isProductLicenseExpired } from '../../../../productSubscription/helpers'
import { AccountName } from '../../../dotcom/productSubscriptions/AccountName'
import { ProductLicenseValidity } from '../../../dotcom/productSubscriptions/ProductLicenseValidity'
import { hasUnknownTags, ProductLicenseTags, UnknownTagWarning } from '../../../productSubscription/ProductLicenseTags'

import { REVOKE_LICENSE } from './backend'

export interface SiteAdminProductLicenseNodeProps {
    node: ProductLicenseFields
    showSubscription: boolean
    defaultExpanded?: boolean
    onRevokeCompleted: () => void
}

/**
 * Displays a product license in a connection in the site admin area.
 */
export const SiteAdminProductLicenseNode: React.FunctionComponent<
    React.PropsWithChildren<SiteAdminProductLicenseNodeProps>
> = ({ node, showSubscription, onRevokeCompleted, defaultExpanded = false }) => {
    const [isOpen, setIsOpen] = useState<boolean>(defaultExpanded)
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
                onCompleted: () => {
                    if (onRevokeCompleted) {
                        onRevokeCompleted()
                    }
                },
            })
        }
    }, [revoke, node, onRevokeCompleted])

    return (
        <li className="list-group-item py-3" id={encodeURIComponent(node?.id)}>
            {showSubscription && (
                <div className="text-truncate d-flex mb-1">
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
            <Collapse isOpen={isOpen} onOpenChange={setIsOpen}>
                <CollapseHeader as={Button} className="w-100 m-0 p-0">
                    <Grid columnCount={3} templateColumns="5fr 1fr 1fr" className="mb-0 align-items-center">
                        {node.info && (
                            <H3 className="mb-1 text-left">
                                <ProductLicenseValidity
                                    license={node}
                                    className="mb-0 mr-1 d-inline"
                                    variant="icon-only"
                                />
                                {node.info.productNameWithBrand}
                                <Icon
                                    className="ml-1"
                                    aria-label={isOpen ? 'Less info' : 'More info'}
                                    svgPath={isOpen ? mdiChevronUp : mdiChevronDown}
                                />
                            </H3>
                        )}
                        {node.info && formatUserCount(node.info.userCount)}
                        {!node?.revokedAt && !isProductLicenseExpired(node?.info?.expiresAt ?? 0) && (
                            <LoaderButton
                                size="sm"
                                variant="danger"
                                className="ml-auto"
                                label="Revoke"
                                onClick={onRevoke}
                                loading={loading}
                            />
                        )}
                    </Grid>
                    <div className="d-flex justify-content-between align-items-center outline-none border-none">
                        <Text className="mb-0 text-muted" as="small" weight="regular">
                            v{node.version}. Created <Timestamp date={node.createdAt} />.{' '}
                            <ProductLicenseValidity license={node} className="mb-0 d-inline" variant="no-icon" />.
                        </Text>
                    </div>
                </CollapseHeader>
                <CollapsePanel>
                    {node.version > 1 && (
                        <>
                            <div className="d-flex mt-1">
                                <Label>Site ID</Label>
                                <Text className="ml-3 mb-0">
                                    {node.siteID ?? <span className="text-muted">Unused</span>}
                                </Text>
                            </div>
                            <div className="d-flex mt-1">
                                <Label>Salesforce Subscription ID</Label>
                                <Text className="ml-3 mb-0">
                                    {node.info?.salesforceSubscriptionID ?? <span className="text-muted">Unused</span>}
                                </Text>
                            </div>
                            <div className="d-flex mt-1">
                                <Label>Salesforce Opportunity ID</Label>
                                <Text className="ml-3 mb-0">
                                    {node.info?.salesforceOpportunityID ?? <span className="text-muted">Unused</span>}
                                </Text>
                            </div>
                        </>
                    )}
                    {node.info && node.info.tags.length > 0 && (
                        <div className="d-flex align-items-baseline mt-1">
                            <Label className="mr-3">Tags</Label>
                            <div className="d-flex justify-content-baseline">
                                {hasUnknownTags(node.info.tags) && <UnknownTagWarning className="mb-2" />}
                                <ProductLicenseTags tags={node.info.tags} />
                            </div>
                        </div>
                    )}
                    <Label className="d-flex align-items-center mb-0">
                        <Text className="mr-3 mb-0">License Key</Text>
                        <CopyableText flex={true} text={node.licenseKey} className="flex-grow-1" />
                    </Label>
                </CollapsePanel>
            </Collapse>
        </li>
    )
}
