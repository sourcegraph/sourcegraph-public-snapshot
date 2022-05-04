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
import {
    Scalars,
    SetProductSubscriptionBillingResult,
    SetProductSubscriptionBillingVariables,
} from '../../../../graphql-operations'

interface Props {
    /** The product subscription to show a billing link for. */
    productSubscription: Pick<GQL.IProductSubscription, 'id' | 'urlForSiteAdminBilling'>

    /** Called when the product subscription is updated. */
    onDidUpdate: () => void
}

const LOADING = 'loading' as const

/**
 * SiteAdminProductSubscriptionBillingLink shows a link to the product subscription on the billing system, if there
 * is an associated billing record. It also supports setting or unsetting the association with the billing system.
 */
export const SiteAdminProductSubscriptionBillingLink: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    productSubscription,
    onDidUpdate,
}) => {
    /** The result of updating this subscription: undefined for done or not started, loading, or an error. */
    const [nextUpdate, update] = useEventObservable(
        useCallback(
            (updates: Observable<{ id: Scalars['ID']; billingSubscriptionID: string | null }>) =>
                updates.pipe(
                    switchMap(({ id, billingSubscriptionID }) =>
                        setProductSubscriptionBilling({ id, billingSubscriptionID }).pipe(
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
        const billingSubscriptionID = window.prompt('Enter new Stripe billing subscription ID:', 'sub_ABCDEF12345678')

        // Ignore if the user pressed "Cancel" or did not enter any value.
        if (!billingSubscriptionID) {
            return
        }

        nextUpdate({ id: productSubscription.id, billingSubscriptionID })
    }, [nextUpdate, productSubscription.id])
    const onUnlinkBillingClick = useCallback(
        () => nextUpdate({ id: productSubscription.id, billingSubscriptionID: null }),
        [nextUpdate, productSubscription.id]
    )

    const productSubscriptionHasLinkedBilling = productSubscription.urlForSiteAdminBilling !== null
    return (
        <div className="site-admin-product-subscription-billing-link">
            <div className="d-flex align-items-center">
                {productSubscription.urlForSiteAdminBilling && (
                    <Link to={productSubscription.urlForSiteAdminBilling} className="mr-2 d-flex align-items-center">
                        View billing subscription <Icon className="ml-1" as={ExternalLinkIcon} />
                    </Link>
                )}
                {isErrorLike(update) && (
                    <Icon className="text-danger mr-2" data-tooltip={update.message} as={AlertCircleIcon} />
                )}
                <Button
                    onClick={productSubscriptionHasLinkedBilling ? onUnlinkBillingClick : onLinkBillingClick}
                    disabled={update === LOADING}
                    variant="secondary"
                    size="sm"
                >
                    {productSubscriptionHasLinkedBilling ? 'Unlink' : 'Link billing subscription'}
                </Button>
            </div>
        </div>
    )
}

function setProductSubscriptionBilling(args: SetProductSubscriptionBillingVariables): Observable<void> {
    return requestGraphQL<SetProductSubscriptionBillingResult, SetProductSubscriptionBillingVariables>(
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
