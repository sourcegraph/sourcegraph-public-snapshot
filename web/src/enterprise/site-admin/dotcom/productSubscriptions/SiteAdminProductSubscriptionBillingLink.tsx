import ErrorIcon from 'mdi-react/ErrorIcon'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import React from 'react'
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
import { gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../../../backend/graphql'

interface Props {
    /** The product subscription to show a billing link for. */
    productSubscription: Pick<GQL.IProductSubscription, 'id' | 'urlForSiteAdminBilling'>

    /** Called when the product subscription is updated. */
    onDidUpdate: () => void
}

const LOADING: 'loading' = 'loading'

interface State {
    /** The result of updating this subscription: null for done or not started, loading, or an error. */
    updateOrError: typeof LOADING | null | ErrorLike
}

/**
 * SiteAdminProductSubscriptionBillingLink shows a link to the product subscription on the billing system, if there
 * is an associated billing record. It also supports setting or unsetting the association with the billing system.
 */
export class SiteAdminProductSubscriptionBillingLink extends React.PureComponent<Props, State> {
    public state: State = {
        updateOrError: null,
    }

    private componentUpdates = new Subject<Props>()
    private updates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const productSubscriptionChanges = this.componentUpdates.pipe(
            map(props => props.productSubscription),
            distinctUntilChanged()
        )

        this.subscriptions.add(
            this.updates
                .pipe(
                    withLatestFrom(productSubscriptionChanges),
                    map(([, { id, urlForSiteAdminBilling }]) => ({
                        id,
                        billingSubscriptionID: urlForSiteAdminBilling
                            ? null
                            : window.prompt('Enter new Stripe billing subscription ID:', 'sub_ABCDEF12345678') ||
                              undefined,
                    })),
                    filter(
                        // Ignore if the user pressed "Cancel".
                        result => result.billingSubscriptionID !== undefined
                    ),
                    switchMap(({ id, billingSubscriptionID }) =>
                        setProductSubscriptionBilling({ id, billingSubscriptionID }).pipe(
                            mapTo(null),
                            tap(() => this.props.onDidUpdate()),
                            catchError(error => [asError(error)]),
                            map(c => ({ updateOrError: c })),
                            startWith({ updateOrError: LOADING })
                        )
                    )
                )
                .subscribe(
                    stateUpdate => this.setState(stateUpdate),
                    error => console.error(error)
                )
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
            <div className="site-admin-product-subscription-billing-link">
                <div className="d-flex align-items-center">
                    {this.props.productSubscription.urlForSiteAdminBilling && (
                        <a
                            href={this.props.productSubscription.urlForSiteAdminBilling}
                            className="mr-2 d-flex align-items-center"
                        >
                            View billing subscription <ExternalLinkIcon className="icon-inline ml-1" />
                        </a>
                    )}
                    {isErrorLike(this.state.updateOrError) && (
                        <ErrorIcon
                            className="icon-inline text-danger mr-2"
                            data-tooltip={this.state.updateOrError.message}
                        />
                    )}
                    <button
                        type="button"
                        className="btn btn-secondary btn-sm"
                        onClick={this.setProductSubscriptionBilling}
                        disabled={this.state.updateOrError === LOADING}
                    >
                        {this.props.productSubscription.urlForSiteAdminBilling ? 'Unlink' : 'Link billing subscription'}
                    </button>
                </div>
            </div>
        )
    }

    private setProductSubscriptionBilling = (): void => this.updates.next()
}

function setProductSubscriptionBilling(
    args: GQL.ISetProductSubscriptionBillingOnDotcomMutationArguments
): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation SetProductSubscriptionBilling($id: ID!, $billingSubscriptionID: String) {
                dotcom {
                    setProductSubscriptionBilling(id: $id, billingSubscriptionID: $billingSubscriptionID) {
                        alwaysNil
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.dotcom || !data.dotcom.setProductSubscriptionBilling || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
        })
    )
}
