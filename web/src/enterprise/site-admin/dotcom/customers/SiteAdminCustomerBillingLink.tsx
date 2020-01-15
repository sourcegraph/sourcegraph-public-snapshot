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
    /** The result of updating that subscription: null for done or not started, loading, or an error. */
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
        const customerChanges = that.componentUpdates.pipe(
            map(props => props.customer),
            distinctUntilChanged()
        )

        that.subscriptions.add(
            that.updates
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
                            tap(() => that.props.onDidUpdate()),
                            catchError(error => [asError(error)]),
                            map(c => ({ updateOrError: c })),
                            startWith({ updateOrError: LOADING })
                        )
                    )
                )
                .subscribe(
                    stateUpdate => that.setState(stateUpdate),
                    error => console.error(error)
                )
        )

        that.componentUpdates.next(that.props)
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-customer-billing-link">
                <div className="d-flex align-items-center">
                    {that.props.customer.urlForSiteAdminBilling && (
                        <a href={that.props.customer.urlForSiteAdminBilling} className="mr-2 d-flex align-items-center">
                            View customer account <ExternalLinkIcon className="icon-inline ml-1" />
                        </a>
                    )}
                    {isErrorLike(that.state.updateOrError) && (
                        <ErrorIcon
                            className="icon-inline text-danger mr-2"
                            data-tooltip={that.state.updateOrError.message}
                        />
                    )}
                    <button
                        type="button"
                        className="btn btn-secondary"
                        onClick={that.setCustomerBilling}
                        disabled={that.state.updateOrError === LOADING}
                    >
                        {that.props.customer.urlForSiteAdminBilling ? 'Unlink' : 'Link billing customer'}
                    </button>
                </div>
            </div>
        )
    }

    private setCustomerBilling = (): void => that.updates.next()
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
