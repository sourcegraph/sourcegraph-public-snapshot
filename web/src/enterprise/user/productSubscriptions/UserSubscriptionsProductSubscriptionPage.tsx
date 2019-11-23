import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { parseISO } from 'date-fns'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, startWith, switchMap } from 'rxjs/operators'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'
import { PageTitle } from '../../../components/PageTitle'
import { SiteAdminAlert } from '../../../site-admin/SiteAdminAlert'
import { eventLogger } from '../../../tracking/eventLogger'
import { mailtoSales } from '../../productSubscription/helpers'
import { BackToAllSubscriptionsLink } from './BackToAllSubscriptionsLink'
import { ProductSubscriptionBilling } from './ProductSubscriptionBilling'
import { ProductSubscriptionHistory } from './ProductSubscriptionHistory'
import { UserProductSubscriptionStatus } from './UserProductSubscriptionStatus'
import { ErrorAlert } from '../../../components/alerts'

interface Props extends RouteComponentProps<{ subscriptionUUID: string }> {
    user: GQL.IUser
}

const LOADING: 'loading' = 'loading'

interface State {
    /**
     * The product subscription, or loading, or an error.
     */
    productSubscriptionOrError: typeof LOADING | GQL.IProductSubscription | ErrorLike
}

/**
 * Displays a product subscription in the user subscriptions area.
 */
export class UserSubscriptionsProductSubscriptionPage extends React.Component<Props, State> {
    public state: State = { productSubscriptionOrError: LOADING }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSubscriptionsProductSubscription')

        const subscriptionUUIDChanges = this.componentUpdates.pipe(
            map(props => props.match.params.subscriptionUUID),
            distinctUntilChanged()
        )

        const productSubscriptionChanges = subscriptionUUIDChanges.pipe(
            switchMap(subscriptionUUID =>
                this.queryProductSubscription(subscriptionUUID).pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
            )
        )

        this.subscriptions.add(
            productSubscriptionChanges
                .pipe(map(result => ({ productSubscriptionOrError: result })))
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )

        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="user-subscriptions-product-subscription-page">
                <PageTitle title="Subscription" />
                <div className="d-flex align-items-center justify-content-between">
                    <BackToAllSubscriptionsLink user={this.props.user} />
                    {this.state.productSubscriptionOrError !== LOADING &&
                        !isErrorLike(this.state.productSubscriptionOrError) &&
                        this.state.productSubscriptionOrError.urlForSiteAdmin && (
                            <SiteAdminAlert className="small m-0 p-1">
                                <Link
                                    to={this.state.productSubscriptionOrError.urlForSiteAdmin}
                                    className="mt-2 d-block"
                                >
                                    View subscription
                                </Link>
                            </SiteAdminAlert>
                        )}
                </div>
                {this.state.productSubscriptionOrError === LOADING ? (
                    <LoadingSpinner className="icon-inline" />
                ) : isErrorLike(this.state.productSubscriptionOrError) ? (
                    <ErrorAlert className="my-2" error={this.state.productSubscriptionOrError} />
                ) : (
                    <>
                        <h2>Subscription {this.state.productSubscriptionOrError.name}</h2>
                        {(this.state.productSubscriptionOrError.invoiceItem ||
                            (this.state.productSubscriptionOrError.activeLicense &&
                                this.state.productSubscriptionOrError.activeLicense.info)) && (
                            <UserProductSubscriptionStatus
                                subscriptionName={this.state.productSubscriptionOrError.name}
                                productNameWithBrand={
                                    this.state.productSubscriptionOrError.activeLicense &&
                                    this.state.productSubscriptionOrError.activeLicense.info
                                        ? this.state.productSubscriptionOrError.activeLicense.info.productNameWithBrand
                                        : this.state.productSubscriptionOrError.invoiceItem!.plan.nameWithBrand
                                }
                                userCount={
                                    this.state.productSubscriptionOrError.activeLicense &&
                                    this.state.productSubscriptionOrError.activeLicense.info
                                        ? this.state.productSubscriptionOrError.activeLicense.info.userCount
                                        : this.state.productSubscriptionOrError.invoiceItem!.userCount
                                }
                                expiresAt={
                                    this.state.productSubscriptionOrError.activeLicense &&
                                    this.state.productSubscriptionOrError.activeLicense.info
                                        ? parseISO(this.state.productSubscriptionOrError.activeLicense.info.expiresAt)
                                        : parseISO(this.state.productSubscriptionOrError.invoiceItem!.expiresAt)
                                }
                                licenseKey={
                                    this.state.productSubscriptionOrError.activeLicense &&
                                    this.state.productSubscriptionOrError.activeLicense.licenseKey
                                }
                            />
                        )}
                        <div className="card mt-3">
                            <div className="card-header">Billing</div>
                            {this.state.productSubscriptionOrError.invoiceItem ? (
                                <>
                                    <ProductSubscriptionBilling
                                        productSubscription={this.state.productSubscriptionOrError}
                                    />
                                    <div className="card-footer">
                                        <a
                                            href={mailtoSales({
                                                subject: `Change payment method for subscription ${this.state.productSubscriptionOrError.name}`,
                                            })}
                                        >
                                            Contact sales
                                        </a>{' '}
                                        to change your payment method.
                                    </div>
                                </>
                            ) : (
                                <div className="card-body">
                                    <span className="text-muted ">
                                        No billing information is associated with this subscription.{' '}
                                        <a
                                            href={mailtoSales({
                                                subject: `Billing for subscription ${this.state.productSubscriptionOrError.name}`,
                                            })}
                                        >
                                            Contact sales
                                        </a>{' '}
                                        for help.
                                    </span>
                                </div>
                            )}
                        </div>
                        <div className="card mt-3">
                            <div className="card-header">History</div>
                            <ProductSubscriptionHistory productSubscription={this.state.productSubscriptionOrError} />
                        </div>
                    </>
                )}
            </div>
        )
    }

    private queryProductSubscription = (uuid: string): Observable<GQL.IProductSubscription> =>
        queryGraphQL(
            gql`
                query ProductSubscription($uuid: String!) {
                    dotcom {
                        productSubscription(uuid: $uuid) {
                            ...ProductSubscriptionFields
                        }
                    }
                }

                fragment ProductSubscriptionFields on ProductSubscription {
                    id
                    name
                    account {
                        id
                        username
                        displayName
                        emails {
                            email
                            verified
                        }
                    }
                    invoiceItem {
                        plan {
                            billingPlanID
                            name
                            nameWithBrand
                            pricePerUserPerYear
                        }
                        userCount
                        expiresAt
                    }
                    events {
                        id
                        date
                        title
                        description
                        url
                    }
                    activeLicense {
                        licenseKey
                        info {
                            productNameWithBrand
                            tags
                            userCount
                            expiresAt
                        }
                    }
                    createdAt
                    isArchived
                    url
                    urlForSiteAdmin
                }
            `,
            { uuid }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.dotcom || !data.dotcom.productSubscription || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
                return data.dotcom.productSubscription
            })
        )
}
