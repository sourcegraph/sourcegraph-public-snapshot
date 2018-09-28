import { gql, queryGraphQL } from '@sourcegraph/webapp/dist/backend/graphql'
import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '@sourcegraph/webapp/dist/util/errors'
import format from 'date-fns/format'
import formatDistanceStrict from 'date-fns/formatDistanceStrict'
import isAfter from 'date-fns/isAfter'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Observable, Subscription } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { ProductLicenseInfoPlanDescription } from './ProductLicenseInfoPlanDescription'

interface Props {
    className?: string
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

        return (
            <>
                <div className={`product-subscription-status card ${this.props.className || ''}`}>
                    <div className="product-subscription-status__bg" />
                    <div className="card-body d-flex">
                        <img
                            className="product-subscription-status__logo mr-1 p-2"
                            src="/.assets/img/sourcegraph-mark.svg"
                        />
                        <div className="mt-2">
                            {isErrorLike(this.state.statusOrError) ? (
                                <div className="alert alert-danger">
                                    Error querying for license information:{' '}
                                    {upperFirst(this.state.statusOrError.message)}
                                </div>
                            ) : this.state.statusOrError.license ? (
                                <>
                                    <h2 className="font-weight-normal">Sourcegraph License</h2>
                                    <h3 className="text-muted font-weight-bold">
                                        <ProductLicenseInfoPlanDescription license={this.state.statusOrError.license} />
                                    </h3>
                                    {this.state.statusOrError.license.expiresAt !== null && (
                                        <p className="text-muted">
                                            Valid until{' '}
                                            {format(this.state.statusOrError.license.expiresAt, 'MMMM dd, yyyy')}{' '}
                                            {isAfter(this.state.statusOrError.license.expiresAt, new Date()) && (
                                                <>
                                                    (
                                                    {formatDistanceStrict(
                                                        this.state.statusOrError.license.expiresAt,
                                                        new Date()
                                                    )}{' '}
                                                    remaining)
                                                </>
                                            )}
                                        </p>
                                    )}
                                </>
                            ) : (
                                <h2 className="font-weight-normal">No Sourcegraph License</h2>
                            )}
                        </div>
                    </div>
                    {!isErrorLike(this.state.statusOrError) &&
                        this.state.statusOrError.license &&
                        this.state.statusOrError.license.userCount !== null && (
                            <div className="card-footer d-flex align-items-center justify-content-between">
                                <div>
                                    <strong>User licenses:</strong> {this.state.statusOrError.actualUserCount} used /{' '}
                                    {this.state.statusOrError.license.userCount -
                                        this.state.statusOrError.actualUserCount}{' '}
                                    remaining
                                </div>
                                <a href="https://about.sourcegraph.com/pricing" className="btn btn-primary btn-sm">
                                    Upgrade license
                                </a>
                            </div>
                        )}
                </div>
            </>
        )
    }

    private queryProductLicenseInfo(): Observable<GQL.IProductSubscriptionStatus> {
        return queryGraphQL(gql`
            query ProductLicenseInfo {
                site {
                    productSubscription {
                        actualUserCount
                        license {
                            plan
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
