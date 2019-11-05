import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import ArrowLeftIcon from 'mdi-react/ArrowLeftIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import {
    catchError,
    distinctUntilChanged,
    filter,
    map,
    mapTo,
    startWith,
    switchMap,
    tap,
    withLatestFrom,
} from 'rxjs/operators'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { mutateGraphQL, queryGraphQL } from '../../../backend/graphql'
import { PageTitle } from '../../../components/PageTitle'
import { eventLogger } from '../../../tracking/eventLogger'
import { ProductSubscriptionForm, ProductSubscriptionFormData } from './ProductSubscriptionForm'
import { ThemeProps } from '../../../../../shared/src/theme'

interface Props extends RouteComponentProps<{ subscriptionUUID: string }>, ThemeProps {
    user: GQL.IUser
}

const LOADING: 'loading' = 'loading'

interface State {
    /**
     * The product subscription, or loading, or an error.
     */
    productSubscriptionOrError: typeof LOADING | GQL.IProductSubscription | ErrorLike

    /**
     * The result of updating the paid product subscription: null when complete or not started yet,
     * loading, or an error.
     */
    updateOrError: null | typeof LOADING | ErrorLike
}

/**
 * Displays a page for editing a product subscription in the user subscriptions area.
 */
export class UserSubscriptionsEditProductSubscriptionPage extends React.Component<Props, State> {
    public state: State = {
        productSubscriptionOrError: LOADING,
        updateOrError: null,
    }

    private componentUpdates = new Subject<Props>()
    private submits = new Subject<ProductSubscriptionFormData>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSubscriptionsEditProductSubscription')

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

        this.subscriptions.add(
            this.submits
                .pipe(
                    withLatestFrom(
                        productSubscriptionChanges.pipe(
                            filter((v): v is GQL.IProductSubscription => v !== LOADING && !isErrorLike(v))
                        )
                    ),
                    switchMap(([args, productSubscription]) =>
                        updatePaidProductSubscription({
                            update: args.productSubscription,
                            subscriptionID: productSubscription.id,
                            paymentToken: args.paymentToken,
                        }).pipe(
                            tap(({ productSubscription }) => {
                                // Redirect back to subscription upon success.
                                this.props.history.push(productSubscription.url)
                            }),
                            mapTo(null),
                            startWith(LOADING)
                        )
                    ),
                    catchError(err => [asError(err)]),
                    map(c => ({ updateOrError: c }))
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
            <div className="user-subscriptions-edit-product-subscription-page">
                <PageTitle title="Edit subscription" />
                {this.state.productSubscriptionOrError === LOADING ? (
                    <LoadingSpinner className="icon-inline" />
                ) : isErrorLike(this.state.productSubscriptionOrError) ? (
                    <div className="alert alert-danger my-2">
                        Error: {this.state.productSubscriptionOrError.message}
                    </div>
                ) : (
                    <>
                        <Link to={this.state.productSubscriptionOrError.url} className="btn btn-link btn-sm mb-3">
                            <ArrowLeftIcon className="icon-inline" /> Subscription
                        </Link>
                        <h2>Upgrade or change subscription {this.state.productSubscriptionOrError.name}</h2>
                        <ProductSubscriptionForm
                            accountID={this.props.user.id}
                            subscriptionID={this.state.productSubscriptionOrError.id}
                            isLightTheme={this.props.isLightTheme}
                            onSubmit={this.onSubmit}
                            submissionState={this.state.updateOrError}
                            initialValue={
                                this.state.productSubscriptionOrError.invoiceItem
                                    ? {
                                          billingPlanID: this.state.productSubscriptionOrError.invoiceItem.plan
                                              .billingPlanID,
                                          userCount: this.state.productSubscriptionOrError.invoiceItem.userCount,
                                      }
                                    : undefined
                            }
                            primaryButtonText="Upgrade subscription"
                            afterPrimaryButton={
                                <small className="form-text text-muted">
                                    An upgraded license key will be available immediately after payment.
                                </small>
                            }
                        />
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
                    invoiceItem {
                        plan {
                            billingPlanID
                        }
                        userCount
                        expiresAt
                    }
                    url
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

    private onSubmit = (args: ProductSubscriptionFormData): void => {
        this.submits.next(args)
    }
}

function updatePaidProductSubscription(
    args: GQL.IUpdatePaidProductSubscriptionOnDotcomMutationArguments
): Observable<GQL.IUpdatePaidProductSubscriptionResult> {
    return mutateGraphQL(
        gql`
            mutation UpdatePaidProductSubscription(
                $subscriptionID: ID!
                $update: ProductSubscriptionInput!
                $paymentToken: String!
            ) {
                dotcom {
                    updatePaidProductSubscription(
                        subscriptionID: $subscriptionID
                        update: $update
                        paymentToken: $paymentToken
                    ) {
                        productSubscription {
                            url
                        }
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.dotcom || !data.dotcom.updatePaidProductSubscription || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return data.dotcom.updatePaidProductSubscription
        })
    )
}
