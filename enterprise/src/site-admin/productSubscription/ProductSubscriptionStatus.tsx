import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable, Subscription } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../../../../packages/webapp/src/backend/graphql'
import * as GQL from '../../../../packages/webapp/src/backend/graphqlschema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../packages/webapp/src/util/errors'
import { numberWithCommas } from '../../../../packages/webapp/src/util/strings'
import { ExpirationDate } from '../../productSubscription/ExpirationDate'
import { formatUserCount } from '../../productSubscription/helpers'
import { ProductCertificate } from '../../productSubscription/ProductCertificate'
import { TrueUpStatusSummary } from '../../productSubscription/TrueUpStatusSummary'

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
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
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
            return (
                <div className="alert alert-danger">
                    Error checking product license: {this.state.statusOrError.message}
                </div>
            )
        }

        const { productNameWithBrand, actualUserCount, license } = this.state.statusOrError

        // No license means Sourcegraph Core. For that, show the user that they can use this for free
        // forever, and show them how to upgrade.

        return (
            <>
                <ProductCertificate
                    title={productNameWithBrand}
                    detail={
                        license ? (
                            <>
                                {formatUserCount(license.userCount, true)} license,{' '}
                                <ExpirationDate
                                    date={license.expiresAt}
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
                                        target="_blank"
                                    >
                                        Upgrade
                                    </a>
                                </>
                            ) : (
                                <>
                                    <div className="mr-2">
                                        Add a license key to activate Sourcegraph Enterprise features
                                    </div>
                                    <div className="text-nowrap flex-wrap-reverse">
                                        <Link
                                            to="/site-admin/configuration"
                                            className="mr-2"
                                            data-tooltip="Set the license key in the licenseKey site configuration property"
                                        >
                                            Add
                                        </Link>
                                        <a
                                            href="http://sourcegraph.com/user/subscriptions"
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
                />
                {license &&
                    (this.props.showTrueUpStatus ? (
                        <TrueUpStatusSummary actualUserCount={actualUserCount} license={license} />
                    ) : (
                        license.userCount - actualUserCount < 0 && (
                            <div className="alert alert-warning">
                                You have exceeded your licensed users.{' '}
                                <Link to="/site-admin/license">View your license details</Link> or{' '}
                                <a href="https://about.sourcegraph.com/pricing" target="_blank">
                                    upgrade your license
                                </a>{' '}
                                to true up and prevent a retroactive charge.
                            </div>
                        )
                    ))}
            </>
        )
    }

    private queryProductLicenseInfo(): Observable<GQL.IProductSubscriptionStatus> {
        return queryGraphQL(gql`
            query ProductLicenseInfo {
                site {
                    productSubscription {
                        productNameWithBrand
                        actualUserCount
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
