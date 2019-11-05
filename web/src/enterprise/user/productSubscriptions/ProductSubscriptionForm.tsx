import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { isEqual, upperFirst } from 'lodash'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { ReactStripeElements } from 'react-stripe-elements'
import { from, of, Subject, Subscription, throwError } from 'rxjs'
import { catchError, map, startWith, switchMap } from 'rxjs/operators'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { Form } from '../../../components/Form'
import { StripeWrapper } from '../../dotcom/billing/StripeWrapper'
import { ProductPlanFormControl } from '../../dotcom/productPlans/ProductPlanFormControl'
import { ProductSubscriptionUserCountFormControl } from '../../dotcom/productPlans/ProductSubscriptionUserCountFormControl'
import { LicenseGenerationKeyWarning } from '../../productSubscription/LicenseGenerationKeyWarning'
import { NewProductSubscriptionPaymentSection } from './NewProductSubscriptionPaymentSection'
import { PaymentTokenFormControl } from './PaymentTokenFormControl'
import { productSubscriptionInputForLocationHash } from './UserSubscriptionsNewProductSubscriptionPage'
import { ThemeProps } from '../../../../../shared/src/theme'

/**
 * The form data that is submitted by the ProductSubscriptionForm component.
 */
export interface ProductSubscriptionFormData {
    /** The customer account (user) owning the product subscription. */
    accountID: GQL.ID
    productSubscription: GQL.IProductSubscriptionInput
    paymentToken: string
}

const LOADING: 'loading' = 'loading'

interface Props extends ThemeProps {
    /**
     * The ID of the account associated with the subscription, or null if there is none (in which case this form
     * can only be used to price out a subscription, not to buy).
     */
    accountID: GQL.ID | null

    /**
     * The existing product subscription to edit, if this form is editing an existing subscription,
     * or null if this is a new subscription.
     */
    subscriptionID: GQL.ID | null

    /** Called when the user submits the form (to buy or update the subscription). */
    onSubmit: (args: ProductSubscriptionFormData) => void

    /** The initial value of the form. */
    initialValue?: GQL.IProductSubscriptionInput

    /**
     * The state of the form submission (the operation triggered by onSubmit): null when it hasn't
     * been submitted yet, loading, or an error. The parent is expected to redirect to another page
     * when the submission is successful, so this component doesn't handle the form submission
     * success state.
     */
    submissionState: null | typeof LOADING | ErrorLike

    /** The text for the form's primary button. */
    primaryButtonText: string

    /** A fragment to render below the form's primary button. */
    afterPrimaryButton?: React.ReactFragment
}

interface State {
    /** The selected product plan. */
    billingPlanID: string | null

    /** The user count input by the user. */
    userCount: number | null

    /** Whether the payment and billing information is valid. */
    paymentValidity: boolean

    /**
     * The result of creating the billing token (which refers to the payment method chosen by the
     * user): null if successful or not yet started, loading, or an error.
     */
    paymentTokenOrError: null | typeof LOADING | ErrorLike
}

/**
 * Displays a form for a product subscription.
 */
// eslint-disable-next-line @typescript-eslint/class-name-casing
class _ProductSubscriptionForm extends React.Component<Props & ReactStripeElements.InjectedStripeProps, State> {
    constructor(props: Props) {
        super(props)

        this.state = {
            paymentValidity: false,
            paymentTokenOrError: null,
            ...this.getStateForInitialValue(props),
        }
    }

    private getStateForInitialValue({ initialValue }: Props): Pick<State, 'billingPlanID' | 'userCount'> {
        return initialValue || { billingPlanID: null, userCount: 1 }
    }

    private submits = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.submits
                .pipe(
                    switchMap(() =>
                        // TODO(sqs): store name, address, company, etc., in token
                        from(this.props.stripe!.createToken()).pipe(
                            switchMap(({ token, error }) => {
                                if (error) {
                                    return throwError(error)
                                }
                                if (!token) {
                                    return throwError(new Error('no payment token'))
                                }
                                if (!this.props.accountID) {
                                    return throwError(new Error('no account (unauthenticated user)'))
                                }
                                if (!this.state.billingPlanID) {
                                    return throwError(new Error('no product plan selected'))
                                }
                                if (this.state.userCount === null) {
                                    return throwError(new Error('invalid user count'))
                                }
                                if (!this.state.paymentValidity) {
                                    return throwError(new Error('invalid payment and billing'))
                                }
                                this.props.onSubmit({
                                    accountID: this.props.accountID,
                                    productSubscription: {
                                        billingPlanID: this.state.billingPlanID,
                                        userCount: this.state.userCount,
                                    },
                                    paymentToken: token.id,
                                })
                                return of(null)
                            }),
                            catchError(err => [asError(err)]),
                            startWith(LOADING)
                        )
                    ),
                    map(result => ({ paymentTokenOrError: result }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )
    }

    public componentDidUpdate(prevProps: Props): void {
        // When Props#initialValue changes, clobber our values. It's unlikely that this prop would
        // change without the component being unmounted, but handle this case for completeness
        // anyway.
        if (!isEqual(prevProps.initialValue, this.props.initialValue)) {
            /* eslint react/no-did-update-set-state: warn */
            this.setState(this.getStateForInitialValue(this.props))
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const disableForm = Boolean(
            this.props.submissionState === LOADING ||
                this.state.userCount === null ||
                !this.state.paymentValidity ||
                this.state.paymentTokenOrError === LOADING ||
                (this.state.paymentTokenOrError && !isErrorLike(this.state.paymentTokenOrError))
        )

        const productSubscriptionInput: GQL.IProductSubscriptionInput | null =
            this.state.billingPlanID !== null && this.state.userCount !== null
                ? {
                      billingPlanID: this.state.billingPlanID,
                      userCount: this.state.userCount,
                  }
                : null

        return (
            <div className="product-subscription-form">
                <LicenseGenerationKeyWarning />
                <Form onSubmit={this.onSubmit}>
                    <div className="row">
                        <div className="col-md-6">
                            <ProductSubscriptionUserCountFormControl
                                value={this.state.userCount}
                                onChange={this.onUserCountChange}
                            />
                            <h4 className="mt-2 mb-0">Plan</h4>
                            <ProductPlanFormControl
                                value={this.state.billingPlanID}
                                onChange={this.onBillingPlanIDChange}
                            />
                        </div>
                        <div className="col-md-6 mt-3 mt-md-0">
                            <h3 className="mt-2 mb-0">Billing</h3>
                            <NewProductSubscriptionPaymentSection
                                productSubscription={productSubscriptionInput}
                                accountID={this.props.accountID}
                                subscriptionID={this.props.subscriptionID}
                                onValidityChange={this.onPaymentValidityChange}
                            />
                            {!this.props.accountID && (
                                <div className="form-group mt-3">
                                    <Link
                                        to={`/sign-up?returnTo=${encodeURIComponent(
                                            `/subscriptions/new${productSubscriptionInputForLocationHash(
                                                productSubscriptionInput
                                            )}`
                                        )}`}
                                        className="btn btn-lg btn-primary w-100 center"
                                    >
                                        Create account or sign in to continue
                                    </Link>
                                    <small className="form-text text-muted">
                                        A user account on Sourcegraph.com is required to create a subscription so you
                                        can view the license key and invoice.
                                    </small>
                                    <hr className="my-3" />
                                    <small className="form-text text-muted">
                                        Next, you'll enter payment information and buy the subscription.
                                    </small>
                                </div>
                            )}
                            <PaymentTokenFormControl
                                disabled={disableForm || !this.props.accountID}
                                isLightTheme={this.props.isLightTheme}
                            />
                            <div className="form-group mt-3">
                                <button
                                    type="submit"
                                    disabled={disableForm || !this.props.accountID}
                                    className={`btn btn-lg btn-${
                                        disableForm || !this.props.accountID ? 'secondary' : 'success'
                                    } w-100 d-flex align-items-center justify-content-center`}
                                >
                                    {this.state.paymentTokenOrError === LOADING ||
                                    this.props.submissionState === LOADING ? (
                                        <>
                                            <LoadingSpinner className="icon-inline mr-2" /> Processing...
                                        </>
                                    ) : (
                                        this.props.primaryButtonText
                                    )}
                                </button>
                                {this.props.afterPrimaryButton}
                            </div>
                        </div>
                    </div>
                </Form>
                {isErrorLike(this.state.paymentTokenOrError) && (
                    <div className="alert alert-danger mt-3">{upperFirst(this.state.paymentTokenOrError.message)}</div>
                )}
                {isErrorLike(this.props.submissionState) && (
                    <div className="alert alert-danger mt-3">{upperFirst(this.props.submissionState.message)}</div>
                )}
            </div>
        )
    }

    private onBillingPlanIDChange = (value: string | null): void => this.setState({ billingPlanID: value })
    private onUserCountChange = (value: number | null): void => this.setState({ userCount: value })
    private onPaymentValidityChange = (value: boolean): void => this.setState({ paymentValidity: value })

    private onSubmit: React.FormEventHandler = e => {
        e.preventDefault()
        this.submits.next()
    }
}

export const ProductSubscriptionForm: React.FunctionComponent<Props> = props => (
    <StripeWrapper<Props> component={_ProductSubscriptionForm} {...props} />
)
