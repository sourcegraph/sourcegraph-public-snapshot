import React, { useCallback } from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import { Observable } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'

import { asError, createAggregateError, isErrorLike } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'
import { Button, useEventObservable, Link, Icon } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../../../backend/graphql'
import { Scalars, SetCustomerBillingResult, SetCustomerBillingVariables } from '../../../../graphql-operations'

interface Props {
    /** The customer to show a billing link for. */
    customer: Pick<GQL.IUser, 'id' | 'urlForSiteAdminBilling'>

    /** Called when the customer is updated. */
    onDidUpdate: () => void
}

const LOADING = 'loading' as const

/**
 * SiteAdminCustomerBillingLink shows a link to the customer on the billing system associated with a user, if any.
 * It also supports setting or unsetting the association with the billing system.
 */
export const SiteAdminCustomerBillingLink: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    customer,
    onDidUpdate,
}) => {
    /** The result of updating this customer: undefined for done or not started, loading, or an error. */
    const [nextUpdate, update] = useEventObservable(
        useCallback(
            (updates: Observable<{ user: Scalars['ID']; billingCustomerID: string | null }>) =>
                updates.pipe(
                    switchMap(({ user, billingCustomerID }) =>
                        setCustomerBilling({ user, billingCustomerID }).pipe(
                            mapTo(undefined),
                            tap(() => onDidUpdate()),
                            catchError(error => [asError(error)]),
                            startWith(LOADING)
                        )
                    )
                ),
            [onDidUpdate]
        )
    )
    const onLinkBillingClick = useCallback(() => {
        const billingCustomerID = window.prompt('Enter new Stripe billing customer ID:', 'cus_ABCDEF12345678')

        // Ignore if the user pressed "Cancel" or did not enter any value.
        if (!billingCustomerID) {
            return
        }

        nextUpdate({ user: customer.id, billingCustomerID })
    }, [customer.id, nextUpdate])
    const onUnlinkBillingClick = useCallback(() => nextUpdate({ user: customer.id, billingCustomerID: null }), [
        customer.id,
        nextUpdate,
    ])

    const customerHasLinkedBilling = customer.urlForSiteAdminBilling !== null
    return (
        <div className="site-admin-customer-billing-link">
            <div className="d-flex align-items-center">
                {customer.urlForSiteAdminBilling && (
                    <Link to={customer.urlForSiteAdminBilling} className="mr-2 d-flex align-items-center">
                        View customer account <Icon className="ml-1" as={ExternalLinkIcon} />
                    </Link>
                )}
                {isErrorLike(update) && (
                    <Icon className="text-danger mr-2" data-tooltip={update.message} as={AlertCircleIcon} />
                )}
                <Button
                    onClick={customerHasLinkedBilling ? onUnlinkBillingClick : onLinkBillingClick}
                    disabled={update === LOADING}
                    variant="secondary"
                >
                    {customerHasLinkedBilling ? 'Unlink' : 'Link billing customer'}
                </Button>
            </div>
        </div>
    )
}

function setCustomerBilling(args: SetCustomerBillingVariables): Observable<void> {
    return requestGraphQL<SetCustomerBillingResult, SetCustomerBillingVariables>(
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
