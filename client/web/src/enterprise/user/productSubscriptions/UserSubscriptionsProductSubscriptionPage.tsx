import React, { useEffect, useMemo } from 'react'

import { parseISO } from 'date-fns'
import * as H from 'history'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { catchError, map, startWith } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, createAggregateError, isErrorLike } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'
import {
    LoadingSpinner,
    useObservable,
    Link,
    CardHeader,
    CardBody,
    Card,
    CardFooter,
    Typography,
} from '@sourcegraph/wildcard'

import { queryGraphQL } from '../../../backend/graphql'
import { PageTitle } from '../../../components/PageTitle'
import { mailtoSales } from '../../../productSubscription/helpers'
import { SiteAdminAlert } from '../../../site-admin/SiteAdminAlert'
import { eventLogger } from '../../../tracking/eventLogger'

import { BackToAllSubscriptionsLink } from './BackToAllSubscriptionsLink'
import { ProductSubscriptionBilling } from './ProductSubscriptionBilling'
import { ProductSubscriptionHistory } from './ProductSubscriptionHistory'
import { UserProductSubscriptionStatus } from './UserProductSubscriptionStatus'

interface Props extends Pick<RouteComponentProps<{ subscriptionUUID: string }>, 'match'> {
    user: Pick<GQL.IUser, 'settingsURL'>

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
            ) : isErrorLike(productSubscription) ? (
                <ErrorAlert className="my-2" error={productSubscription} />
            ) : (
                <>
                    <Typography.H2>Subscription {productSubscription.name}</Typography.H2>
                    {(productSubscription.invoiceItem || productSubscription.activeLicense?.info) && (
                        <UserProductSubscriptionStatus
                            subscriptionName={productSubscription.name}
                            productNameWithBrand={
                                productSubscription.activeLicense?.info
                                    ? productSubscription.activeLicense.info.productNameWithBrand
                                    : productSubscription.invoiceItem!.plan.nameWithBrand
                            }
                            userCount={
                                productSubscription.activeLicense?.info
                                    ? productSubscription.activeLicense.info.userCount
                                    : productSubscription.invoiceItem!.userCount
                            }
                            expiresAt={
                                productSubscription.activeLicense?.info
                                    ? parseISO(productSubscription.activeLicense.info.expiresAt)
                                    : parseISO(productSubscription.invoiceItem!.expiresAt)
                            }
                            licenseKey={productSubscription.activeLicense?.licenseKey ?? null}
                        />
                    )}
                    <Card className="mt-3">
                        <CardHeader>Billing</CardHeader>
                        {productSubscription.invoiceItem ? (
                            <>
                                <ProductSubscriptionBilling productSubscription={productSubscription} />
                                <CardFooter>
                                    <Link
                                        to={mailtoSales({
                                            subject: `Change payment method for subscription ${productSubscription.name}`,
                                        })}
                                    >
                                        Contact sales
                                    </Link>{' '}
                                    to change your payment method.
                                </CardFooter>
                            </>
                        ) : (
                            <CardBody>
                                <span className="text-muted ">
                                    No billing information is associated with this subscription.{' '}
                                    <Link
                                        to={mailtoSales({
                                            subject: `Billing for subscription ${productSubscription.name}`,
                                        })}
                                    >
                                        Contact sales
                                    </Link>{' '}
                                    for help.
                                </span>
                            </CardBody>
                        )}
                    </Card>
                    <Card className="mt-3">
                        <CardHeader>History</CardHeader>
                        <ProductSubscriptionHistory productSubscription={productSubscription} />
                    </Card>
                </>
            )}
        </div>
    )
}

function queryProductSubscription(uuid: string): Observable<GQL.IProductSubscription> {
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
                invoiceItem {
                    plan {
                        billingPlanID
                        name
                        nameWithBrand
                        pricePerUserPerYear
                    }
                    userCount
                    expiresAt
                }
                events {
                    id
                    date
                    title
                    description
                    url
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
