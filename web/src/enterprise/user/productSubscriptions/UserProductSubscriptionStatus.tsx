import InformationIcon from 'mdi-react/InformationIcon'
import KeyIcon from 'mdi-react/KeyIcon'
import React from 'react'
import { CopyableText } from '../../../components/CopyableText'
import { ExpirationDate } from '../../productSubscription/ExpirationDate'
import { formatUserCount, mailtoSales } from '../../productSubscription/helpers'
import { LicenseGenerationKeyWarning } from '../../productSubscription/LicenseGenerationKeyWarning'
import { ProductCertificate } from '../../productSubscription/ProductCertificate'

interface Props {
    subscriptionName: string
    productNameWithBrand: string
    userCount: number
    expiresAt: Date | number
    licenseKey: string | null
}

interface State {
    showLicenseKey: boolean
}

/**
 * Displays a certificate with information about and status for a user's product subscription. It
 * supports both billing-linked and non-billing-linked subscriptions.
 */
export class UserProductSubscriptionStatus extends React.PureComponent<Props, State> {
    public state: State = { showLicenseKey: false }

    public render(): JSX.Element | null {
        return (
            <ProductCertificate
                title={this.props.productNameWithBrand}
                subtitle={
                    <>
                        {formatUserCount(this.props.userCount, true)} license,{' '}
                        <ExpirationDate
                            date={this.props.expiresAt}
                            showRelative={true}
                            showPrefix={true}
                            lowercase={true}
                        />
                    </>
                }
                footer={
                    <>
                        <div className="card-footer d-flex align-items-center justify-content-between flex-wrap">
                            <button
                                type="button"
                                className="btn btn-primary mr-4 my-1"
                                onClick={this.toggleShowLicenseKey}
                            >
                                <KeyIcon className="icon-inline" /> {this.state.showLicenseKey ? 'Hide' : 'Reveal'}{' '}
                                license key
                            </button>
                            <div className="flex-fill" />
                            <div className="my-1" />
                        </div>
                        {this.state.showLicenseKey && (
                            <div className="card-footer">
                                <h3>License key</h3>
                                {this.props.licenseKey ? (
                                    <>
                                        <CopyableText text={this.props.licenseKey} className="d-block" />
                                        <small className="mt-2 d-flex align-items-center">
                                            <InformationIcon className="icon-inline mr-1" />{' '}
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
                                        <a
                                            href={mailtoSales({
                                                subject: `No license key for subscription ${this.props.subscriptionName}`,
                                            })}
                                        >
                                            Contact sales
                                        </a>{' '}
                                        for help.
                                    </div>
                                )}
                            </div>
                        )}
                    </>
                }
            />
        )
    }

    private toggleShowLicenseKey = (): void =>
        this.setState(prevState => ({ showLicenseKey: !prevState.showLicenseKey }))
}
