import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { gql, queryGraphQL } from '@sourcegraph/webapp/dist/backend/graphql'
import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '@sourcegraph/webapp/dist/util/errors'
import { numberWithCommas } from '@sourcegraph/webapp/dist/util/strings'
import { isEqual } from 'lodash'
import ErrorIcon from 'mdi-react/ErrorIcon'
import * as React from 'react'
import { ReactStripeElements } from 'react-stripe-elements'
import { Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, startWith, switchMap, tap } from 'rxjs/operators'
import { ProductSubscriptionInput } from '../../dotcom/productSubscriptions/helpers'
import { formatUserCount } from '../../productSubscription/helpers'
import { PaymentTokenFormControl } from './PaymentTokenFormControl'

interface Props {
    accountID: string

    /**
     * The product subscription chosen by the user, or null for an invalid choice.
     */
    productSubscription: ProductSubscriptionInput | null

    disabled?: boolean
    isLightTheme: boolean

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
            map(({ accountID, productSubscription }) => ({ accountID, productSubscription })),
            distinctUntilChanged(
                (a, b) => a.accountID === b.accountID && isEqual(a.productSubscription, b.productSubscription)
            )
        )

        this.subscriptions.add(
            argChanges
                .pipe(
                    switchMap(({ accountID, productSubscription }) => {
                        if (productSubscription === null) {
                            return of(null)
                        }
                        return queryPreviewProductSubscriptionInvoice({
                            account: accountID,
                            productSubscription: {
                                billingPlanID: productSubscription.plan.billingPlanID,
                                userCount: productSubscription.userCount,
                            },
                        }).pipe(
                            catchError(err => [asError(err)]),
                            startWith(LOADING)
                        )
                    }),
                    tap(result =>
                        this.props.onValidityChange(result !== null && result !== LOADING && !isErrorLike(result))
                    ),
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

    public render(): JSX.Element | null {
        return (
            <div className="new-product-subscription-payment-section">
                <div className="form-text mb-2">
                    Total:{' '}
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
                    ) : (
                        <>
                            ${numberWithCommas(this.state.previewInvoiceOrError.amountDue / 100)} for 1 year (
                            {formatUserCount(this.props.productSubscription.userCount)})
                        </>
                    )}
                </div>
                <PaymentTokenFormControl disabled={this.props.disabled} isLightTheme={this.props.isLightTheme} />
            </div>
        )
    }
}

function queryPreviewProductSubscriptionInvoice(
    args: GQL.IPreviewProductSubscriptionInvoiceOnDotcomQueryArguments
): Observable<GQL.IProductSubscriptionPreviewInvoice> {
    return queryGraphQL(
        gql`
            query PreviewProductSubscriptionInvoice($account: ID!, $productSubscription: ProductSubscriptionInput!) {
                dotcom {
                    previewProductSubscriptionInvoice(account: $account, productSubscription: $productSubscription) {
                        amountDue
                        prorationDate
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
