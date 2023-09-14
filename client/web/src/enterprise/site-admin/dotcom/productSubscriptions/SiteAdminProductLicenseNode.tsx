import React, { useCallback, useMemo, useState } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { useMutation } from '@sourcegraph/http-client'
import {
    Alert,
    Button,
    Collapse,
    CollapseHeader,
    CollapsePanel,
    Grid,
    H3,
    Icon,
    Label,
    LinkOrSpan,
    Text,
} from '@sourcegraph/wildcard'

import { CopyableText } from '../../../../components/CopyableText'
import { LoaderButton } from '../../../../components/LoaderButton'
import type { ProductLicenseFields, RevokeLicenseResult, RevokeLicenseVariables } from '../../../../graphql-operations'
import { isProductLicenseExpired } from '../../../../productSubscription/helpers'
import { AccountName } from '../../../dotcom/productSubscriptions/AccountName'
import { ProductLicenseValidity } from '../../../dotcom/productSubscriptions/ProductLicenseValidity'
import { ProductLicenseInfoDescription } from '../../../productSubscription/ProductLicenseInfoDescription'
import { ProductLicenseTags, UnknownTagWarning, hasUnknownTags } from '../../../productSubscription/ProductLicenseTags'

import { REVOKE_LICENSE } from './backend'

const getLicenseUUID = (id: string): string => atob(id).slice('ProductLicense:"'.length, -1)

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

    const [open, setOpen] = useState(defaultExpanded)
    const toggleOpen = useCallback(() => {
        setOpen(!open)
    }, [open, setOpen])

    const uuid = useMemo(() => getLicenseUUID(node.id), [node])

    return (
        <li className="list-group-item p-3 mb-3 border">
            <Collapse isOpen={open} onOpenChange={setOpen}>
                <Grid columnCount={2} templateColumns="auto 1fr" spacing={0}>
                    <Button variant="icon" onClick={toggleOpen} className="pr-3">
                        <Icon
                            aria-label={`collapse ${open ? 'opened' : 'closed'}`}
                            svgPath={open ? mdiChevronUp : mdiChevronDown}
                        />
                    </Button>
                    <CollapseHeader as="div" className="d-flex align-items-start">
                        <div>
                            {showSubscription && (
                                <div className="text-truncate d-flex">
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
                            {!loading && error && (
                                <Alert variant="danger">Error revoking license: {error.message}</Alert>
                            )}
                            <div className="d-flex align-items-baseline mb-2">
                                {node.info && (
                                    <ProductLicenseInfoDescription licenseInfo={node.info} className="mb-0" />
                                )}
                                <Text className="ml-3 mb-0">
                                    <small className="text-muted">
                                        Created <Timestamp date={node.createdAt} />
                                    </small>
                                </Text>
                                <Text className="ml-3 mb-0 text-muted">
                                    <small>v{node.version}</small>
                                </Text>
                            </div>
                            <ProductLicenseValidity license={node} />
                        </div>
                        {!node?.revokedAt && !isProductLicenseExpired(node?.info?.expiresAt ?? 0) && (
                            <LoaderButton
                                className="ml-auto"
                                variant="danger"
                                label="Revoke"
                                onClick={onRevoke}
                                loading={loading}
                            />
                        )}
                    </CollapseHeader>
                    <div />
                    <CollapsePanel className="mt-4">
                        <div className="d-flex">
                            <Label>License Key ID</Label>
                            <Text className="ml-3">{uuid}</Text>
                        </div>
                        {node.version > 1 && (
                            <>
                                <div className="d-flex">
                                    <Label>Site ID</Label>
                                    <Text className="ml-3">
                                        {node.siteID ?? <span className="text-muted">Unused</span>}
                                    </Text>
                                </div>
                                <div className="d-flex">
                                    <Label>Salesforce Subscription ID</Label>
                                    <Text className="ml-3">
                                        {node.info?.salesforceSubscriptionID ?? (
                                            <span className="text-muted">Unused</span>
                                        )}
                                    </Text>
                                </div>
                                <div className="d-flex">
                                    <Label>Salesforce Opportunity ID</Label>
                                    <Text className="ml-3">
                                        {node.info?.salesforceOpportunityID ?? (
                                            <span className="text-muted">Unused</span>
                                        )}
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
                    </CollapsePanel>
                </Grid>
            </Collapse>
        </li>
    )
}
