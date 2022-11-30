import React, { useEffect, useMemo } from 'react'

import { parseISO } from 'date-fns'
import * as H from 'history'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { catchError, map, startWith } from 'rxjs/operators'
import { validate as validateUUID } from 'uuid'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, createAggregateError, isErrorLike } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { LoadingSpinner, useObservable, Link, H2 } from '@sourcegraph/wildcard'

import { queryGraphQL } from '../../../backend/graphql'
import { PageTitle } from '../../../components/PageTitle'
import { ProductSubscriptionFieldsOnSubscriptionPage, UserAreaUserFields } from '../../../graphql-operations'
import { SiteAdminAlert } from '../../../site-admin/SiteAdminAlert'
import { eventLogger } from '../../../tracking/eventLogger'

import { BackToAllSubscriptionsLink } from './BackToAllSubscriptionsLink'
import { UserProductSubscriptionStatus } from './UserProductSubscriptionStatus'

interface Props extends Pick<RouteComponentProps<{ subscriptionUUID: string }>, 'match'> {
    user: Pick<UserAreaUserFields, 'settingsURL'>

    /** For mocking in tests only. */
    _queryProductSubscription?: typeof queryProductSubscription
    history: H.History
}

const LOADING = 'loading' as const

/**
 * Displays a product subscription in the user subscriptions area.
 */
export const UserSubscriptionsProductSubscriptionPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    user,
    match: {
        params: { subscriptionUUID },
    },
    _queryProductSubscription = queryProductSubscription,
}) => {
    useEffect(() => eventLogger.logViewEvent('UserSubscriptionsProductSubscription'), [])

    const isValidUUID = validateUUID(subscriptionUUID)
    const validationError = !isValidUUID && new Error('Subscription ID is not a valid UUID')

    /**
     * The product subscription, or loading, or an error.
     */
    const productSubscription =
        useObservable(
            useMemo(
                () =>
                    _queryProductSubscription(subscriptionUUID).pipe(
                        catchError(error => [asError(error)]),
                        startWith(LOADING)
                    ),
                [_queryProductSubscription, subscriptionUUID]
            )
        ) || LOADING

    return (
        <div className="user-subscriptions-product-subscription-page">
            <PageTitle title="Subscription" />
            <div className="d-flex align-items-center justify-content-between">
                <BackToAllSubscriptionsLink user={user} />
                {productSubscription !== LOADING &&
                    !isErrorLike(productSubscription) &&
                    productSubscription.urlForSiteAdmin && (
                        <SiteAdminAlert className="small m-0">
                            <Link to={productSubscription.urlForSiteAdmin} className="mt-2 d-block">
                                View subscription
                            </Link>
                        </SiteAdminAlert>
                    )}
            </div>
            {productSubscription === LOADING ? (
                <LoadingSpinner />
            ) : !isValidUUID || isErrorLike(productSubscription) ? (
                <ErrorAlert className="my-2" error={validationError || productSubscription} />
            ) : (
                <>
                    <H2>Subscription {productSubscription.name}</H2>
                    {productSubscription.activeLicense?.info && (
                        <UserProductSubscriptionStatus
                            subscriptionName={productSubscription.name}
                            productNameWithBrand={productSubscription.activeLicense?.info.productNameWithBrand}
                            userCount={productSubscription.activeLicense?.info.userCount}
                            expiresAt={parseISO(productSubscription.activeLicense.info.expiresAt)}
                            licenseKey={productSubscription.activeLicense?.licenseKey ?? null}
                        />
                    )}
                </>
            )}
        </div>
    )
}

function queryProductSubscription(uuid: string): Observable<ProductSubscriptionFieldsOnSubscriptionPage> {
    return queryGraphQL(
        gql`
            query ProductSubscription($uuid: String!) {
                dotcom {
                    productSubscription(uuid: $uuid) {
                        ...ProductSubscriptionFieldsOnSubscriptionPage
                    }
                }
            }

            fragment ProductSubscriptionFieldsOnSubscriptionPage on ProductSubscription {
                id
                name
                account {
                    id
                    username
                    displayName
                    emails {
                        email
                        verified
                    }
                }
                activeLicense {
                    licenseKey
                    info {
                        productNameWithBrand
                        tags
                        userCount
                        expiresAt
                    }
                }
                createdAt
                isArchived
                url
                urlForSiteAdmin
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
}
