import * as React from 'react'
import { of, Subscription } from 'rxjs'
import { ajax } from 'rxjs/ajax'
import { catchError, delay, map } from 'rxjs/operators'
import { ExpirationDate } from '../../../../shared/src/productSubscription/ExpirationDate'
import { formatUserCount } from '../../../../shared/src/productSubscription/helpers'
import { ProductCertificate } from '../../../../shared/src/productSubscription/ProductCertificate'
import { TrueUpStatusSummary } from '../../../../shared/src/productSubscription/TrueUpStatusSummary'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { numberWithCommas } from '../../../../shared/src/util/strings'
import logo from './sourcegraph-mark.svg'

const DEBUG_LOADING_STATE_DELAY = 0 // ms

interface Props {
    className?: string

    /**
     * If true, always show the license true-up status.
     * If undefined or false, never show the full license true-up status, and instead only show an alert
     * if the user count is over the license limit.
     *
     */
    showTrueUpStatus?: boolean
}

interface State {
    licenseOrError?: LicenseKeyInfo | ErrorLike
}

interface LicenseKeyInfo {
    /**
     * The number of users on an instance.
     */
    UserCount: number
    ExpiresAt: string
    ProductNameWithBrand: string
    ActualUserCount: number
    ActualUserCountDate: string
    ExternalURL: string
}

/**
 * A component displaying information about and the status of the product subscription.
 */
export class ProductSubscriptionStatus extends React.Component<Props, State> {
    public state: State = {
        licenseOrError: undefined,
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Load the initial critical config.
        this.subscriptions.add(
            ajax('/api/license')
                .pipe(
                    delay(DEBUG_LOADING_STATE_DELAY),
                    catchError(err => of(err.xhr))
                )
                .subscribe(resp => {
                    if (resp.status !== 200) {
                        const msg = 'error fetching license: ' + resp.status
                        console.error(msg)
                        this.setState({ licenseOrError: new Error(msg) })
                        return
                    }

                    const license = resp.response as LicenseKeyInfo
                    if (license) {
                        this.setState({
                            licenseOrError: {
                                UserCount: license.UserCount,
                                ExpiresAt: license.ExpiresAt,
                                ProductNameWithBrand: license.ProductNameWithBrand,
                                ActualUserCount: license.ActualUserCount,
                                ActualUserCountDate: license.ActualUserCountDate,
                                ExternalURL: license.ExternalURL,
                            },
                        })
                    }
                })
        )
    }
    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.licenseOrError === undefined) {
            return null
        }

        if (isErrorLike(this.state.licenseOrError)) {
            return (
                <div className="alert alert-danger">
                    Error checking product license: {this.state.licenseOrError.message}
                </div>
            )
        }

        const license = this.state.licenseOrError

        if (license.ProductNameWithBrand === 'Sourcegraph OSS') {
            // Don't show license status on OSS builds.
            return null
        }

        const isNotCore = license && license.ProductNameWithBrand !== 'Sourcegraph Core'

        // No license means Sourcegraph Core. For that, show the user that they can use this for free forever, and show them how to upgrade.
        return (
            <div className="product-subscription-status my-3">
                <ProductCertificate
                    title={license.ProductNameWithBrand}
                    detail={
                        isNotCore ? (
                            <>
                                {formatUserCount(license.UserCount, true)} license,{' '}
                                <ExpirationDate
                                    date={license.ExpiresAt}
                                    showRelative={true}
                                    lowercase={true}
                                    showPrefix={true}
                                />
                            </>
                        ) : null
                    }
                    footer={
                        <div className="card-footer d-flex align-items-center justify-content-between">
                            {isNotCore ? (
                                <>
                                    <div>
                                        <strong>User licenses:</strong> {numberWithCommas(license.ActualUserCount)} used
                                        / {numberWithCommas(license.UserCount - license.ActualUserCount)} remaining
                                    </div>
                                    <a
                                        href="https://about.sourcegraph.com/pricing"
                                        className="btn btn-primary btn-sm"
                                        target="_blank"
                                    >
                                        Upgrade
                                    </a>
                                </>
                            ) : (
                                <>
                                    <div className="mr-2">
                                        Add a license key to activate Sourcegraph Enterprise features{' '}
                                        {typeof license.UserCount === 'number'
                                            ? `or to exceed ${license.UserCount} users`
                                            : ''}
                                    </div>
                                    <div className="text-nowrap flex-wrap-reverse">
                                        <a
                                            href="http://sourcegraph.com/user/subscriptions/new"
                                            className="btn btn-primary btn-sm"
                                            target="_blank"
                                            data-tooltip="Buy a Sourcegraph Enterprise subscription to get a license key"
                                        >
                                            Get license
                                        </a>
                                    </div>
                                </>
                            )}
                        </div>
                    }
                    className={this.props.className}
                    logoSrc={logo}
                />
                {this.props.showTrueUpStatus ? (
                    <TrueUpStatusSummary
                        actualUserCount={license.ActualUserCount}
                        actualUserCountDate={license.ActualUserCountDate}
                        userCount={license.UserCount}
                    />
                ) : (
                    license.UserCount - license.ActualUserCount < 0 && (
                        <div className="product-subscription-status__alert alert alert-warning">
                            You have exceeded your licensed users.{' '}
                            <a href={`${license.ExternalURL}/site-admin/license`}>View your license details</a> or{' '}
                            <a href="https://about.sourcegraph.com/pricing" target="_blank">
                                upgrade your license
                            </a>{' '}
                            to true up and prevent a retroactive charge.
                        </div>
                    )
                )}
            </div>
        )
    }
}
