import { parseISO } from 'date-fns'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable, Subscription } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { numberWithCommas } from '../../../../../shared/src/util/strings'
import { queryGraphQL } from '../../../backend/graphql'
import { ExpirationDate } from '../../productSubscription/ExpirationDate'
import { formatUserCount } from '../../productSubscription/helpers'
import { ProductCertificate } from '../../productSubscription/ProductCertificate'
import { TrueUpStatusSummary } from '../../productSubscription/TrueUpStatusSummary'
import { ErrorAlert } from '../../../components/alerts'

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
    /** The product subscription status, or an error, or undefined while loading. */
    statusOrError?: GQL.IProductSubscriptionStatus | ErrorLike
}

/**
 * A component displaying information about and the status of the product subscription.
 */
export class ProductSubscriptionStatus extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.queryProductLicenseInfo()
                .pipe(
                    catchError(err => [asError(err)]),
                    map(v => ({ statusOrError: v }))
                )
                .subscribe(
                    stateUpdate => this.setState(stateUpdate),
                    err => console.error(err)
                )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.statusOrError === undefined) {
            return null
        }
        if (isErrorLike(this.state.statusOrError)) {
            return <ErrorAlert error={this.state.statusOrError} prefix="Error checking product license" />
        }

        const {
            productNameWithBrand,
            actualUserCount,
            actualUserCountDate,
            noLicenseWarningUserCount,
            license,
        } = this.state.statusOrError

        // No license means Sourcegraph Core. For that, show the user that they can use this for free
        // forever, and show them how to upgrade.

        return (
            <div className="mt-3">
                <ProductCertificate
                    title={productNameWithBrand}
                    detail={
                        license ? (
                            <>
                                {formatUserCount(license.userCount, true)} license,{' '}
                                <ExpirationDate
                                    date={parseISO(license.expiresAt)}
                                    showRelative={true}
                                    lowercase={true}
                                    showPrefix={true}
                                />
                            </>
                        ) : null
                    }
                    footer={
                        <div className="card-footer d-flex align-items-center justify-content-between">
                            {license ? (
                                <>
                                    <div>
                                        <strong>User licenses:</strong> {numberWithCommas(actualUserCount)} used /{' '}
                                        {numberWithCommas(license.userCount - actualUserCount)} remaining
                                    </div>
                                    <a
                                        href="https://about.sourcegraph.com/pricing"
                                        className="btn btn-primary btn-sm"
                                        // eslint-disable-next-line react/jsx-no-target-blank
                                        target="_blank"
                                    >
                                        Upgrade
                                    </a>
                                </>
                            ) : (
                                <>
                                    <div className="mr-2">
                                        Add a license key to activate Sourcegraph Enterprise features{' '}
                                        {typeof noLicenseWarningUserCount === 'number'
                                            ? `or to exceed ${noLicenseWarningUserCount} users`
                                            : ''}
                                    </div>
                                    <div className="text-nowrap flex-wrap-reverse">
                                        <a
                                            href="http://sourcegraph.com/subscriptions/new"
                                            className="btn btn-primary btn-sm"
                                            // eslint-disable-next-line react/jsx-no-target-blank
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
                />
                {license &&
                    (this.props.showTrueUpStatus ? (
                        <TrueUpStatusSummary
                            actualUserCount={actualUserCount}
                            actualUserCountDate={actualUserCountDate}
                            license={license}
                        />
                    ) : (
                        license.userCount - actualUserCount < 0 && (
                            <div className="alert alert-warning">
                                You have exceeded your licensed users.{' '}
                                <Link to="/site-admin/license">View your license details</Link> or{' '}
                                {/* eslint-disable-next-line react/jsx-no-target-blank */}
                                <a href="https://about.sourcegraph.com/pricing" target="_blank">
                                    upgrade your license
                                </a>{' '}
                                to true up and prevent a retroactive charge.
                            </div>
                        )
                    ))}
            </div>
        )
    }

    private queryProductLicenseInfo(): Observable<GQL.IProductSubscriptionStatus> {
        return queryGraphQL(gql`
            query ProductLicenseInfo {
                site {
                    productSubscription {
                        productNameWithBrand
                        actualUserCount
                        actualUserCountDate
                        noLicenseWarningUserCount
                        license {
                            tags
                            userCount
                            expiresAt
                        }
                    }
                }
            }
        `).pipe(
            map(({ data, errors }) => {
                if (!data || !data.site || !data.site.productSubscription) {
                    throw createAggregateError(errors)
                }
                return data.site.productSubscription
            })
        )
    }
}
