import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'
import { gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../../../backend/graphql'
import { HeroPage } from '../../../../components/HeroPage'
import { PageTitle } from '../../../../components/PageTitle'
import { eventLogger } from '../../../../tracking/eventLogger'
import { BackToAllSubscriptionsLink } from './BackToAllSubscriptionsLink'
import { ProductSubscriptionForm, ProductSubscriptionFormData } from './ProductSubscriptionForm'
import { ThemeProps } from '../../../../../../shared/src/theme'

interface Props extends RouteComponentProps<{}>, ThemeProps {
    /**
     * The user who will own the new trial subscription when created, or null when there is no
     * authenticated user and this page is accessed at /subscriptions/new-trial.
     */
    user: GQL.IUser | null
}

const LOADING: 'loading' = 'loading'

interface State {
    /**
     * The result of creating the product trial subscription: null when complete or not started yet,
     * loading, or an error.
     */
    creationOrError: null | typeof LOADING | ErrorLike
}

/**
 * Displays a form to create a new trial subscription.
 *
 * This page is visible to both authenticated and unauthenticated users. Unauthenticated users may
 * view it at /subscriptions/new-trial and are allowed to start a trial.
 */
export class UserSubscriptionsNewProductTrialSubscriptionPage extends React.Component<Props, State> {
    public state: State = { creationOrError: null }

    private submits = new Subject<GQL.ICreatePaidProductSubscriptionOnDotcomMutationArguments>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSubscriptionsNewProductTrialSubscription')
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
        if (this.props.user && !this.props.user.viewerCanAdminister) {
            return <HeroPage icon={AlertCircleIcon} title="Not authorized" />
        }

        return (
            <div className="user-subscriptions-new-product-trial-subscription-page">
                <PageTitle title="Free trial" />
                <h2>Free trial of Sourcegraph Elite</h2>
                <h1>TRIAL FORM HERE</h1>
                <input type="text" className="form-control" />
                <br />
                Foo
                <br />
                <input type="text" className="form-control" />
                <br />
                Foo
                <br />
                <input type="text" className="form-control" />
                <br />
                Foo
                <br />
                <input type="text" className="form-control" />
                <br />
                <Link
                    className="btn btn-primary"
                    to="/site-admin/license?licenseKey=eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiJFMTdrUEVjcHRxNy9JOS8rNnJYTXhwVlZUYWZ3ZXZKemNieGwxTmRpQ1dSUmx6Wjh2Y0Q0b0pVamVKclAraXRyWUZwU0xManFSQ0R5QnpSQVlmdHVTbUp5MXRxK1QyQW0xOE9OWEdtc2ZwbWxKUkZSNm40RjhzMkhhWVpHck8wZFE4YmZHdExuc1pGcTdNTFJiZldKR1dDaEpkK3owN3hROVZXalNFSUk2OXRoKy94STlYZG0wL3FoUzZjMllJWTNkQnpxaExPYnFDd2tCYVlZbjRiWFlQTEU4UnllZDU4S2ExZHJWMTVGYVFoeENnakgzaC9WRi81Wmhoei9US2xONU5kVEhsZDdLaG8raG4rSjNyYXFUWmJJa1hPMzIrOEZVT0F6RU5jM2FZU2FNSmZ5ZUpJam05WjRxZW1iOEhXVE9qa1QwTWt5QjlkUGFNZkZPVG11TGc9PSJ9LCJpbmZvIjoiZXlKMklqb3hMQ0p1SWpwYk16RXNPVGNzTlRBc01qTXdMREkwTnl3MU9Dd3hNQ3cwTkYwc0luUWlPbHNpWkdWMklsMHNJblVpT2pFd01Dd2laU0k2SWpJd01qQXRNVEF0TWpsVU1UTTZNVFU2TURsYUluMD0ifQ"
                >
                    Start your free trial
                </Link>
            </div>
        )
    }

    private onSubmit = (args: ProductSubscriptionFormData): void => {
        this.submits.next(args)
    }
}

/**
 * Parses product subscription input from the URL hash.
 *
 * Inverse of {@link productSubscriptionInputForLocationHash}.
 */
function parseProductSubscriptionInputFromLocation(location: H.Location): GQL.IProductSubscriptionInput | null {
    if (location.hash) {
        const params = new URLSearchParams(location.hash.slice('#'.length))
        const billingPlanID = params.get('plan')
        const userCount = parseInt(params.get('userCount') || '0', 10)
        if (billingPlanID && userCount) {
            return { billingPlanID, userCount }
        }
    }
    return null
}

/**
 * Generates the URL hash value to represent the product subscription input.
 *
 * Inverse of {@link parseProductSubscriptionInputFromLocation}.
 */
export function productSubscriptionInputForLocationHash(value: GQL.IProductSubscriptionInput | null): string {
    if (value === null) {
        return ''
    }
    const params = new URLSearchParams()
    params.set('plan', value.billingPlanID)
    params.set('userCount', value.userCount.toString())
    return '#' + params.toString()
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
