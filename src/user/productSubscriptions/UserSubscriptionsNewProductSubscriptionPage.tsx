import { gql, mutateGraphQL } from '@sourcegraph/webapp/dist/backend/graphql'
import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import { PageTitle } from '@sourcegraph/webapp/dist/components/PageTitle'
import { eventLogger } from '@sourcegraph/webapp/dist/tracking/eventLogger'
import { asError, createAggregateError, ErrorLike } from '@sourcegraph/webapp/dist/util/errors'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'
import { BackToAllSubscriptionsLink } from './BackToAllSubscriptionsLink'
import { ProductSubscriptionForm, ProductSubscriptionFormData } from './ProductSubscriptionForm'

interface Props extends RouteComponentProps<{}> {
    user: GQL.IUser
    isLightTheme: boolean
}

const LOADING: 'loading' = 'loading'

interface State {
    /**
     * The result of creating the paid product subscription: null when complete or not started yet,
     * loading, or an error.
     */
    creationOrError: null | typeof LOADING | ErrorLike
}

/**
 * Displays a form and payment flow to purchase a product subscription.
 */
export class UserSubscriptionsNewProductSubscriptionPage extends React.Component<Props, State> {
    public state: State = { creationOrError: null }

    private submits = new Subject<GQL.ICreatePaidProductSubscriptionOnDotcomMutationArguments>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSubscriptionsNewProductSubscription')
        this.subscriptions.add(
            this.submits
                .pipe(
                    switchMap(args =>
                        createPaidProductSubscription(args).pipe(
                            tap(({ productSubscription }) => {
                                // Redirect to new subscription upon success.
                                this.props.history.push(productSubscription.url)
                            }),
                            mapTo(null),
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
        return (
            <div className="user-subscriptions-new-product-subscription-page">
                <PageTitle title="New product subscription" />
                <BackToAllSubscriptionsLink user={this.props.user} />
                <h2>New subscription</h2>
                <ProductSubscriptionForm
                    accountID={this.props.user.id}
                    subscriptionID={null}
                    isLightTheme={this.props.isLightTheme}
                    onSubmit={this.onSubmit}
                    submissionState={this.state.creationOrError}
                    primaryButtonText="Buy subscription"
                    afterPrimaryButton={
                        <small className="form-text text-muted">
                            Your license key will be available immediately after payment.
                        </small>
                    }
                />
            </div>
        )
    }

    private onSubmit = (args: ProductSubscriptionFormData) => {
        this.submits.next(args)
    }
}

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
