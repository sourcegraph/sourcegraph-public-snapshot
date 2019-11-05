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
    /** The customer to show a billing link for. */
    customer: Pick<GQL.IUser, 'id' | 'urlForSiteAdminBilling'>

    /** Called when the customer is updated. */
    onDidUpdate: () => void
}

const LOADING: 'loading' = 'loading'

interface State {
    /** The result of updating this subscription: null for done or not started, loading, or an error. */
    updateOrError: typeof LOADING | null | ErrorLike
}

/**
 * SiteAdminCustomerBillingLink shows a link to the customer on the billing system associated with a user, if any.
 * It also supports setting or unsetting the association with the billing system.
 */
export class SiteAdminCustomerBillingLink extends React.PureComponent<Props, State> {
    public state: State = {
        updateOrError: null,
    }

    private componentUpdates = new Subject<Props>()
    private updates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const customerChanges = this.componentUpdates.pipe(
            map(props => props.customer),
            distinctUntilChanged()
        )

        this.subscriptions.add(
            this.updates
                .pipe(
                    withLatestFrom(customerChanges),
                    map(([, { id, urlForSiteAdminBilling }]) => ({
                        user: id,
                        billingCustomerID: urlForSiteAdminBilling
                            ? null
                            : window.prompt('Enter new Stripe billing customer ID:', 'cus_ABCDEF12345678') || undefined,
                    })),
                    filter(
                        // Ignore if the user pressed "Cancel".
                        result => result.billingCustomerID !== undefined
                    ),
                    switchMap(({ user, billingCustomerID }) =>
                        setCustomerBilling({ user, billingCustomerID }).pipe(
                            mapTo(null),
                            tap(() => this.props.onDidUpdate()),
                            catchError(error => [asError(error)]),
                            map(c => ({ updateOrError: c })),
                            startWith({ updateOrError: LOADING })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
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
            <div className="site-admin-customer-billing-link">
                <div className="d-flex align-items-center">
                    {this.props.customer.urlForSiteAdminBilling && (
                        <a href={this.props.customer.urlForSiteAdminBilling} className="mr-2 d-flex align-items-center">
                            View customer account <ExternalLinkIcon className="icon-inline ml-1" />
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
                        className="btn btn-secondary"
                        onClick={this.setCustomerBilling}
                        disabled={this.state.updateOrError === LOADING}
                    >
                        {this.props.customer.urlForSiteAdminBilling ? 'Unlink' : 'Link billing customer'}
                    </button>
                </div>
            </div>
        )
    }

    private setCustomerBilling = (): void => this.updates.next()
}

function setCustomerBilling(args: GQL.ISetUserBillingOnDotcomMutationArguments): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation SetCustomerBilling($user: ID!, $billingCustomerID: String) {
                dotcom {
                    setUserBilling(user: $user, billingCustomerID: $billingCustomerID) {
                        alwaysNil
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.dotcom || !data.dotcom.setUserBilling || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
        })
    )
}
