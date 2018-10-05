import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { gql, mutateGraphQL } from '@sourcegraph/webapp/dist/backend/graphql'
import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import { Form } from '@sourcegraph/webapp/dist/components/Form'
import { PageTitle } from '@sourcegraph/webapp/dist/components/PageTitle'
import { eventLogger } from '@sourcegraph/webapp/dist/tracking/eventLogger'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '@sourcegraph/webapp/dist/util/errors'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { ReactStripeElements } from 'react-stripe-elements'
import { from, Observable, Subject, Subscription, throwError } from 'rxjs'
import { catchError, map, startWith, switchMap, tap } from 'rxjs/operators'
import { StripeWrapper } from '../../dotcom/billing/StripeWrapper'
import { ProductPlanFormControl } from '../../dotcom/productPlans/ProductPlanFormControl'
import { ProductSubscriptionUserCountFormControl } from '../../dotcom/productPlans/ProductSubscriptionUserCountFormControl'
import { BackToAllSubscriptionsLink } from './BackToAllSubscriptionsLink'
import { NewProductSubscriptionPaymentSection } from './NewProductSubscriptionPaymentSection'

interface Props extends RouteComponentProps<{}> {
    user: GQL.IUser
    isLightTheme: boolean
}

const LOADING: 'loading' = 'loading'

interface State {
    /** The selected product plan. */
    plan: GQL.IProductPlan | null

    /** The user count input by the user. */
    userCount: number | null

    /**
     * The result of creating the paid product subscription, or null when not pending or complete, or loading, or
     * an error.
     */
    creationOrError: null | GQL.ICreatePaidProductSubscriptionResult | typeof LOADING | ErrorLike
}

/**
 * Displays a form and payment flow to purchase a product subscription.
 */
// tslint:disable-next-line:class-name
class _UserSubscriptionsNewProductSubscriptionPage extends React.Component<
    Props & ReactStripeElements.InjectedStripeProps,
    State
> {
    private get emptyState(): Pick<State, 'plan' | 'userCount' | 'creationOrError'> {
        return {
            plan: null,
            userCount: 1,
            creationOrError: null,
        }
    }

    public state: State = { ...this.emptyState }

    private submits = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSubscriptionsNewProductSubscription')
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
                                return createPaidProductSubscription({
                                    accountID: this.props.user.id,
                                    productSubscription: {
                                        plan: this.state.plan.billingID,
                                        userCount: this.state.userCount,
                                        totalPriceNonAuthoritative:
                                            this.state.plan.pricePerUserPerYear * this.state.userCount,
                                    },
                                    paymentToken: token.id,
                                })
                            }),
                            tap(({ productSubscription }) => {
                                // Redirect to new subscription upon success.
                                this.props.history.push(productSubscription.url)
                            }),
                            catchError(err => [asError(err)]),
                            startWith(LOADING),
                            map(c => ({ creationOrError: c }))
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const disableForm = Boolean(
            this.state.userCount === null ||
                this.state.creationOrError === LOADING ||
                (this.state.creationOrError && !isErrorLike(this.state.creationOrError))
        )

        return (
            <div className="user-subscriptions-new-product-subscription-page">
                <PageTitle title="New product subscription" />
                <BackToAllSubscriptionsLink user={this.props.user} />
                <h2>New subscription</h2>
                <div className="alert alert-warning">
                    Subscriptions and license keys will be introduced in Sourcegraph 2.12 (to be released on Monday,
                    October 8, 2018). Only use this payment form before then if you've been directed here by
                    Sourcegraph.
                </div>
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
                                              totalPriceNonAuthoritative:
                                                  this.state.plan.pricePerUserPerYear * this.state.userCount,
                                          }
                                        : null
                                }
                                disabled={disableForm}
                                isLightTheme={this.props.isLightTheme}
                                user={this.props.user}
                            />
                            <div className="form-group mt-3">
                                <button
                                    type="submit"
                                    disabled={disableForm}
                                    className={`btn btn-lg btn-${
                                        disableForm ? 'secondary' : 'primary'
                                    } w-100 d-flex align-items-center justify-content-center`}
                                >
                                    {this.state.creationOrError === LOADING ? (
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
                {isErrorLike(this.state.creationOrError) && (
                    <div className="alert alert-danger mt-3">{upperFirst(this.state.creationOrError.message)}</div>
                )}
            </div>
        )
    }

    private onPlanChange = (value: GQL.IProductPlan | null): void => this.setState({ plan: value })
    private onUserCountChange = (value: number | null): void => this.setState({ userCount: value })

    private onSubmit: React.FormEventHandler = e => {
        e.preventDefault()
        this.submits.next()
    }
}

export const UserSubscriptionsNewProductSubscriptionPage: React.SFC<Props> = props => (
    <StripeWrapper<Props> component={_UserSubscriptionsNewProductSubscriptionPage} {...props} />
)

function createPaidProductSubscription(
    args: GQL.ICreatePaidProductSubscriptionOnDotcomMutationArguments
): Observable<GQL.ICreatePaidProductSubscriptionResult> {
    return mutateGraphQL(
        gql`
            mutation CreatePaidProductSubscription(
                $accountID: ID!
                $productSubscription: ProductSubscriptionInput!
                $paymentToken: String!
            ) {
                dotcom {
                    createPaidProductSubscription(
                        accountID: $accountID
                        productSubscription: $productSubscription
                        paymentToken: $paymentToken
                    ) {
                        productSubscription {
                            id
                            name
                            url
                        }
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.dotcom || !data.dotcom.createPaidProductSubscription || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return data.dotcom.createPaidProductSubscription
        })
    )
}
