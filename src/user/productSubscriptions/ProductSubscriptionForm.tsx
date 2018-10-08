import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import { Form } from '@sourcegraph/webapp/dist/components/Form'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/webapp/dist/util/errors'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { ReactStripeElements } from 'react-stripe-elements'
import { from, of, Subject, Subscription, throwError } from 'rxjs'
import { catchError, map, startWith, switchMap } from 'rxjs/operators'
import { StripeWrapper } from '../../dotcom/billing/StripeWrapper'
import { ProductPlanFormControl } from '../../dotcom/productPlans/ProductPlanFormControl'
import { ProductSubscriptionUserCountFormControl } from '../../dotcom/productPlans/ProductSubscriptionUserCountFormControl'
import { LicenseGenerationKeyWarning } from '../../productSubscription/LicenseGenerationKeyWarning'
import { NewProductSubscriptionPaymentSection } from './NewProductSubscriptionPaymentSection'

const LOADING: 'loading' = 'loading'

interface Props {
    /** The ID of the account associated with the subscription. */
    accountID: GQL.ID

    isLightTheme: boolean

    /** Called when the user submits the form (to buy or update the subscription). */
    onSubmit: (args: GQL.ICreatePaidProductSubscriptionOnDotcomMutationArguments) => void

    /**
     * The state of the form submission (the operation triggered by onSubmit): null when it hasn't
     * been submitted yet, loading, or an error. The parent is expected to redirect to another page
     * when the submission is successful, so this component doesn't handle the form submission
     * success state.
     */
    submissionState: null | typeof LOADING | ErrorLike
}

interface State {
    /** The selected product plan. */
    plan: GQL.IProductPlan | null

    /** The user count input by the user. */
    userCount: number | null

    /** Whether the payment and billing information is valid. */
    paymentValidity: boolean

    /**
     * The result of creating the billing token (which refers to the payment method chosen by the
     * user): null if successful or not yet started, loading, or an error.
     */
    billingTokenOrError: null | typeof LOADING | ErrorLike
}

/**
 * Displays a form for a product subscription.
 */
// tslint:disable-next-line:class-name
class _ProductSubscriptionForm extends React.Component<Props & ReactStripeElements.InjectedStripeProps, State> {
    public state: State = {
        plan: null,
        userCount: 1,
        paymentValidity: false,
        billingTokenOrError: null,
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
                                if (!this.state.plan) {
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
                                        billingPlanID: this.state.plan.billingPlanID,
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
                    map(result => ({ billingTokenOrError: result }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const disableForm = Boolean(
            this.props.submissionState === LOADING ||
                isErrorLike(this.props.submissionState) ||
                this.state.userCount === null ||
                !this.state.paymentValidity ||
                this.state.billingTokenOrError === LOADING ||
                (this.state.billingTokenOrError && !isErrorLike(this.state.billingTokenOrError))
        )

        return (
            <div className="product-subscription-form">
                <div className="alert alert-warning">
                    Subscriptions and license keys will be introduced in Sourcegraph 2.12 (coming soon). Only use this
                    payment form before then if you've been directed here by Sourcegraph.
                </div>
                <LicenseGenerationKeyWarning />
                <Form onSubmit={this.onSubmit}>
                    <div className="row">
                        <div className="col-md-6">
                            <ProductSubscriptionUserCountFormControl
                                plan={this.state.plan}
                                value={this.state.userCount}
                                onChange={this.onUserCountChange}
                            />
                            <h4 className="mt-2 mb-0">Plan</h4>
                            <ProductPlanFormControl value={this.state.plan} onChange={this.onPlanChange} />
                        </div>
                        <div className="col-md-6 mt-3 mt-md-0">
                            <h3 className="mt-2 mb-0">Billing</h3>
                            <NewProductSubscriptionPaymentSection
                                productSubscription={
                                    this.state.plan && this.state.userCount !== null
                                        ? {
                                              plan: this.state.plan,
                                              userCount: this.state.userCount,
                                          }
                                        : null
                                }
                                disabled={disableForm}
                                isLightTheme={this.props.isLightTheme}
                                accountID={this.props.accountID}
                                onValidityChange={this.onPaymentValidityChange}
                            />
                            <div className="form-group mt-3">
                                <button
                                    type="submit"
                                    disabled={disableForm}
                                    className={`btn btn-lg btn-${
                                        disableForm ? 'secondary' : 'primary'
                                    } w-100 d-flex align-items-center justify-content-center`}
                                >
                                    {this.state.billingTokenOrError === LOADING ||
                                    this.props.submissionState === LOADING ? (
                                        <>
                                            <LoadingSpinner className="icon-inline mr-2" /> Processing...
                                        </>
                                    ) : (
                                        'Buy subscription'
                                    )}
                                </button>
                                <small className="form-text text-muted">
                                    Your license key will be available immediately after payment.
                                </small>
                            </div>
                        </div>
                    </div>
                </Form>
                {isErrorLike(this.state.billingTokenOrError) && (
                    <div className="alert alert-danger mt-3">{upperFirst(this.state.billingTokenOrError.message)}</div>
                )}
            </div>
        )
    }

    private onPlanChange = (value: GQL.IProductPlan | null): void => this.setState({ plan: value })
    private onUserCountChange = (value: number | null): void => this.setState({ userCount: value })
    private onPaymentValidityChange = (value: boolean) => this.setState({ paymentValidity: value })

    private onSubmit: React.FormEventHandler = e => {
        e.preventDefault()
        this.submits.next()
    }
}

export const ProductSubscriptionForm: React.SFC<Props> = props => (
    <StripeWrapper<Props> component={_ProductSubscriptionForm} {...props} />
)
