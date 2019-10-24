import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { parseISO } from 'date-fns'
import formatDistanceStrict from 'date-fns/formatDistanceStrict'
import { isEqual } from 'lodash'
import ErrorIcon from 'mdi-react/ErrorIcon'
import * as React from 'react'
import { ReactStripeElements } from 'react-stripe-elements'
import { Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, startWith, switchMap, tap } from 'rxjs/operators'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { numberWithCommas } from '../../../../../shared/src/util/strings'
import { queryGraphQL } from '../../../backend/graphql'
import { formatUserCount, mailtoSales } from '../../productSubscription/helpers'
import { ProductSubscriptionBeforeAfterInvoiceItem } from './ProductSubscriptionBeforeAfterInvoiceItem'

interface Props {
    /**
     * The ID of the account associated with the subscription, or null if there is none (in which case the
     * subscription price can be quoted but the subscription can't be bought).
     */
    accountID: GQL.ID | null

    /** The existing product subscription to edit, or null if this is a new subscription. */
    subscriptionID: GQL.ID | null

    /**
     * The product subscription chosen by the user, or null for an invalid choice.
     */
    productSubscription: GQL.IProductSubscriptionInput | null

    /**
     * Called when the validity state of the payment and billing information changes. Initially it
     * is always false.
     */
    onValidityChange: (value: boolean) => void
}

const LOADING: 'loading' = 'loading'

interface State {
    /**
     * The preview invoice for the subscription, null if the input is invalid to generate an
     * invoice, loading, or an error.
     */
    previewInvoiceOrError: GQL.IProductSubscriptionPreviewInvoice | null | typeof LOADING | ErrorLike
}

/**
 * Displays the payment section of the new product subscription form.
 */
export class NewProductSubscriptionPaymentSection extends React.PureComponent<
    Props & ReactStripeElements.InjectedStripeProps,
    State
> {
    public state: State = {
        previewInvoiceOrError: LOADING,
    }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const argChanges = this.componentUpdates.pipe(
            map(({ accountID, subscriptionID, productSubscription }) => ({
                accountID,
                subscriptionID,
                productSubscription,
            })),
            distinctUntilChanged(
                (a, b) =>
                    a.accountID === b.accountID &&
                    a.subscriptionID === b.subscriptionID &&
                    isEqual(a.productSubscription, b.productSubscription)
            )
        )

        this.subscriptions.add(
            argChanges
                .pipe(
                    switchMap(({ accountID, subscriptionID, productSubscription }) => {
                        if (productSubscription === null) {
                            return of(null)
                        }
                        return queryPreviewProductSubscriptionInvoice({
                            account: accountID,
                            subscriptionToUpdate: subscriptionID,
                            productSubscription,
                        }).pipe(
                            catchError(err => [asError(err)]),
                            startWith(LOADING)
                        )
                    }),
                    tap(result => this.props.onValidityChange(!this.isPreviewInvoiceInvalid(result))),
                    map(result => ({ previewInvoiceOrError: result }))
                )
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

    private isPreviewInvoiceInvalid(previewInvoiceOrError: State['previewInvoiceOrError']): boolean {
        return Boolean(
            previewInvoiceOrError === null ||
                previewInvoiceOrError === LOADING ||
                isErrorLike(previewInvoiceOrError) ||
                isEqual(previewInvoiceOrError.beforeInvoiceItem, previewInvoiceOrError.afterInvoiceItem) ||
                previewInvoiceOrError.isDowngradeRequiringManualIntervention
        )
    }

    public render(): JSX.Element | null {
        return (
            <div className="new-product-subscription-payment-section">
                <div className="form-text mb-2">
                    {this.state.previewInvoiceOrError === LOADING ? (
                        <LoadingSpinner className="icon-inline" />
                    ) : !this.props.productSubscription || this.state.previewInvoiceOrError === null ? (
                        <>&mdash;</>
                    ) : isErrorLike(this.state.previewInvoiceOrError) ? (
                        <span className="text-danger">
                            <ErrorIcon
                                className="icon-inline"
                                data-tooltip={this.state.previewInvoiceOrError.message}
                            />{' '}
                            Error
                        </span>
                    ) : this.state.previewInvoiceOrError.beforeInvoiceItem ? (
                        <>
                            <ProductSubscriptionBeforeAfterInvoiceItem
                                beforeInvoiceItem={this.state.previewInvoiceOrError.beforeInvoiceItem}
                                afterInvoiceItem={this.state.previewInvoiceOrError.afterInvoiceItem}
                                className="mb-2"
                            />
                            {this.state.previewInvoiceOrError.isDowngradeRequiringManualIntervention ? (
                                <div className="alert alert-danger mb-2">
                                    Self-service downgrades are not yet supported.{' '}
                                    <a
                                        href={mailtoSales({
                                            subject: `Downgrade subscription ${this.props.subscriptionID}`,
                                        })}
                                    >
                                        Contact sales
                                    </a>{' '}
                                    for help.
                                </div>
                            ) : (
                                !isEqual(
                                    this.state.previewInvoiceOrError.beforeInvoiceItem,
                                    this.state.previewInvoiceOrError.afterInvoiceItem
                                ) && (
                                    <div className="mb-2">
                                        Amount due: ${numberWithCommas(this.state.previewInvoiceOrError.price / 100)}
                                    </div>
                                )
                            )}
                        </>
                    ) : (
                        <>
                            Total: ${numberWithCommas(this.state.previewInvoiceOrError.price / 100)} for{' '}
                            {formatDistanceStrict(
                                parseISO(this.state.previewInvoiceOrError.afterInvoiceItem.expiresAt),
                                Date.now()
                            )}{' '}
                            ({formatUserCount(this.props.productSubscription.userCount)})
                            {/* Include invisible LoadingSpinner to ensure that the height remains constant between loading and total. */}
                            <LoadingSpinner className="icon-inline invisible" />
                        </>
                    )}
                </div>
            </div>
        )
    }
}

function queryPreviewProductSubscriptionInvoice(
    args: GQL.IPreviewProductSubscriptionInvoiceOnDotcomQueryArguments
): Observable<GQL.IProductSubscriptionPreviewInvoice> {
    return queryGraphQL(
        gql`
            query PreviewProductSubscriptionInvoice(
                $account: ID!
                $subscriptionToUpdate: ID
                $productSubscription: ProductSubscriptionInput!
            ) {
                dotcom {
                    previewProductSubscriptionInvoice(
                        account: $account
                        subscriptionToUpdate: $subscriptionToUpdate
                        productSubscription: $productSubscription
                    ) {
                        price
                        prorationDate
                        isDowngradeRequiringManualIntervention
                        beforeInvoiceItem {
                            plan {
                                billingPlanID
                                name
                                pricePerUserPerYear
                            }
                            userCount
                            expiresAt
                        }
                        afterInvoiceItem {
                            plan {
                                billingPlanID
                                name
                                pricePerUserPerYear
                            }
                            userCount
                            expiresAt
                        }
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (
                !data ||
                !data.dotcom ||
                !data.dotcom.previewProductSubscriptionInvoice ||
                (errors && errors.length > 0)
            ) {
                throw createAggregateError(errors)
            }
            return data.dotcom.previewProductSubscriptionInvoice
        })
    )
}
