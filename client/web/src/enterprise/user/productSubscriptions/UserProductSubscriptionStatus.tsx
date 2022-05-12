import React, { useState, useCallback } from 'react'

import InformationIcon from 'mdi-react/InformationIcon'
import KeyIcon from 'mdi-react/KeyIcon'

import { Button, CardFooter, Link, Icon, Typography } from '@sourcegraph/wildcard'

import { CopyableText } from '../../../components/CopyableText'
import { formatUserCount, mailtoSales } from '../../../productSubscription/helpers'
import { ExpirationDate } from '../../productSubscription/ExpirationDate'
import { LicenseGenerationKeyWarning } from '../../productSubscription/LicenseGenerationKeyWarning'
import { ProductCertificate } from '../../productSubscription/ProductCertificate'

interface Props {
    subscriptionName: string
    productNameWithBrand: string
    userCount: number
    expiresAt: Date | number
    licenseKey: string | null
}

/**
 * Displays a certificate with information about and status for a user's product subscription. It
 * supports both billing-linked and non-billing-linked subscriptions.
 */
export const UserProductSubscriptionStatus: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    subscriptionName,
    productNameWithBrand,
    userCount,
    expiresAt,
    licenseKey,
}) => {
    const [showLicenseKey, setShowLicenseKey] = useState(false)

    const toggleShowLicenseKey = useCallback((): void => setShowLicenseKey(!showLicenseKey), [showLicenseKey])

    return (
        <ProductCertificate
            title={productNameWithBrand}
            subtitle={
                <>
                    {formatUserCount(userCount, true)} license,{' '}
                    <ExpirationDate date={expiresAt} showRelative={true} showPrefix={true} lowercase={true} />
                </>
            }
            footer={
                <>
                    <CardFooter className="d-flex align-items-center justify-content-between flex-wrap">
                        <Button className="mr-4 my-1" onClick={toggleShowLicenseKey} variant="primary">
                            <Icon as={KeyIcon} /> {showLicenseKey ? 'Hide' : 'Reveal'} license key
                        </Button>
                        <div className="flex-fill" />
                        <div className="my-1" />
                    </CardFooter>
                    {showLicenseKey && (
                        <CardFooter>
                            <Typography.H3>License key</Typography.H3>
                            {licenseKey ? (
                                <>
                                    <CopyableText text={licenseKey} className="d-block" />
                                    <small className="mt-2 d-flex align-items-center">
                                        <Icon className="mr-1" as={InformationIcon} />{' '}
                                        <span>
                                            Use this license key as the{' '}
                                            <code>
                                                <strong>licenseKey</strong>
                                            </code>{' '}
                                            property value in Sourcegraph site configuration.
                                        </span>
                                    </small>
                                    <LicenseGenerationKeyWarning className="mb-0 mt-1" />
                                </>
                            ) : (
                                <div className="text-muted">
                                    No license key found.{' '}
                                    <Link
                                        to={mailtoSales({
                                            subject: `No license key for subscription ${subscriptionName}`,
                                        })}
                                    >
                                        Contact sales
                                    </Link>{' '}
                                    for help.
                                </div>
                            )}
                        </CardFooter>
                    )}
                </>
            }
        />
    )
}
