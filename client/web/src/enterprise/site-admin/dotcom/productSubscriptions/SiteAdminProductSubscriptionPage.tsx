import React, { useState, useMemo, useEffect, useCallback } from 'react'

import * as H from 'history'
import AddIcon from 'mdi-react/AddIcon'
import ArrowLeftIcon from 'mdi-react/ArrowLeftIcon'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject, NEVER } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap, tap, filter } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, createAggregateError, isErrorLike } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'
import {
    Button,
    LoadingSpinner,
    useObservable,
    useEventObservable,
    Link,
    CardHeader,
    CardBody,
    Card,
    Icon,
    Typography,
} from '@sourcegraph/wildcard'

import { queryGraphQL, requestGraphQL } from '../../../../backend/graphql'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { Timestamp } from '../../../../components/time/Timestamp'
import { ArchiveProductSubscriptionResult, ArchiveProductSubscriptionVariables } from '../../../../graphql-operations'
import { eventLogger } from '../../../../tracking/eventLogger'
import { AccountEmailAddresses } from '../../../dotcom/productSubscriptions/AccountEmailAddresses'
import { AccountName } from '../../../dotcom/productSubscriptions/AccountName'
import { ProductSubscriptionLabel } from '../../../dotcom/productSubscriptions/ProductSubscriptionLabel'
import { LicenseGenerationKeyWarning } from '../../../productSubscription/LicenseGenerationKeyWarning'
import { ProductSubscriptionHistory } from '../../../user/productSubscriptions/ProductSubscriptionHistory'

import { SiteAdminGenerateProductLicenseForSubscriptionForm } from './SiteAdminGenerateProductLicenseForSubscriptionForm'
import {
    siteAdminProductLicenseFragment,
    SiteAdminProductLicenseNode,
    SiteAdminProductLicenseNodeProps,
} from './SiteAdminProductLicenseNode'
import { SiteAdminProductSubscriptionBillingLink } from './SiteAdminProductSubscriptionBillingLink'

interface Props extends RouteComponentProps<{ subscriptionUUID: string }> {
    /** For mocking in tests only. */
    _queryProductSubscription?: typeof queryProductSubscription

    /** For mocking in tests only. */
    _queryProductLicenses?: typeof queryProductLicenses
    history: H.History
}

class FilteredSiteAdminProductLicenseConnection extends FilteredConnection<
    GQL.IProductLicense,
    Pick<SiteAdminProductLicenseNodeProps, 'showSubscription'>
> {}

const LOADING = 'loading' as const

/**
 * Displays a product subscription in the site admin area.
 */
export const SiteAdminProductSubscriptionPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    history,
    location,
    match: {
        params: { subscriptionUUID },
    },
    _queryProductSubscription = queryProductSubscription,
    _queryProductLicenses = queryProductLicenses,
}) => {
    useEffect(() => eventLogger.logViewEvent('SiteAdminProductSubscription'), [])

    const [showGenerate, setShowGenerate] = useState<boolean>(false)

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

    /** The result of archiving this subscription: undefined for done or not started, loading, or an error. */
    const [nextArchival, archival] = useEventObservable(
        useCallback(
            (archivals: Observable<React.MouseEvent>) => {
                if (productSubscription === LOADING || isErrorLike(productSubscription)) {
                    return NEVER
                }
                return archivals.pipe(
                    filter(() =>
                        window.confirm(
                            'Really archive this product subscription? This will hide it from site admins and users.\n\nHowever, it does NOT:\n\n- invalidate the license key\n- refund payment or cancel billing\n\nYou must manually do those things.'
                        )
                    ),
                    switchMap(() =>
                        archiveProductSubscription({ id: productSubscription.id }).pipe(
                            mapTo(undefined),
                            tap(() => history.push('/site-admin/dotcom/product/subscriptions')),
                            catchError(error => [asError(error)]),
                            startWith(LOADING)
                        )
                    )
                )
            },
            [history, productSubscription]
        )
    )

    const queryProductLicensesForSubscription = useCallback(
        (args: { first?: number }) => _queryProductLicenses(subscriptionUUID, args),
        [_queryProductLicenses, subscriptionUUID]
    )

    const toggleShowGenerate = useCallback((): void => setShowGenerate(previousValue => !previousValue), [])

    /** Updates to the subscription. */
    const updates = useMemo(() => new Subject<void>(), [])
    const onUpdate = useCallback(() => updates.next(), [updates])

    /** Updates to the subscription's licenses. */
    const licenseUpdates = useMemo(() => new Subject<void>(), [])
    const onLicenseUpdate = useCallback(() => {
        licenseUpdates.next()
        toggleShowGenerate()
    }, [licenseUpdates, toggleShowGenerate])

    const nodeProps: Pick<SiteAdminProductLicenseNodeProps, 'showSubscription'> = {
        showSubscription: false,
    }

    return (
        <div className="site-admin-product-subscription-page">
            <PageTitle title="Product subscription" />
            <div className="mb-2">
                <Button to="/site-admin/dotcom/product/subscriptions" variant="link" size="sm" as={Link}>
                    <Icon as={ArrowLeftIcon} /> All subscriptions
                </Button>
            </div>
            {productSubscription === LOADING ? (
                <LoadingSpinner />
            ) : isErrorLike(productSubscription) ? (
                <ErrorAlert className="my-2" error={productSubscription} />
            ) : (
                <>
                    <Typography.H2>Product subscription {productSubscription.name}</Typography.H2>
                    <div className="mb-3">
                        <Button onClick={nextArchival} disabled={archival === LOADING} variant="danger">
                            Archive
                        </Button>
                        {isErrorLike(archival) && <ErrorAlert className="mt-2" error={archival} />}
                    </div>
                    <Card className="mt-3">
                        <CardHeader>Details</CardHeader>
                        <table className="table mb-0">
                            <tbody>
                                <tr>
                                    <th className="text-nowrap">ID</th>
                                    <td className="w-100">{productSubscription.name}</td>
                                </tr>
                                <tr>
                                    <th className="text-nowrap">Plan</th>
                                    <td className="w-100">
                                        <ProductSubscriptionLabel productSubscription={productSubscription} />
                                    </td>
                                </tr>
                                <tr>
                                    <th className="text-nowrap">Account</th>
                                    <td className="w-100">
                                        <AccountName account={productSubscription.account} /> &mdash;{' '}
                                        <Link to={productSubscription.url}>View as user</Link>
                                    </td>
                                </tr>
                                <tr>
                                    <th className="text-nowrap">Account emails</th>
                                    <td className="w-100">
                                        {productSubscription.account && (
                                            <AccountEmailAddresses emails={productSubscription.account.emails} />
                                        )}
                                    </td>
                                </tr>
                                <tr>
                                    <th className="text-nowrap">Billing</th>
                                    <td className="w-100">
                                        <SiteAdminProductSubscriptionBillingLink
                                            productSubscription={productSubscription}
                                            onDidUpdate={onUpdate}
                                        />
                                    </td>
                                </tr>
                                <tr>
                                    <th className="text-nowrap">Created at</th>
                                    <td className="w-100">
                                        <Timestamp date={productSubscription.createdAt} />
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </Card>
                    <LicenseGenerationKeyWarning className="mt-3" />
                    <Card className="mt-1">
                        <CardHeader className="d-flex align-items-center justify-content-between">
                            Licenses
                            {showGenerate ? (
                                <Button onClick={toggleShowGenerate} variant="secondary">
                                    Dismiss new license form
                                </Button>
                            ) : (
                                <Button onClick={toggleShowGenerate} variant="primary" size="sm">
                                    <Icon as={AddIcon} /> Generate new license manually
                                </Button>
                            )}
                        </CardHeader>
                        {showGenerate && (
                            <CardBody>
                                <SiteAdminGenerateProductLicenseForSubscriptionForm
                                    subscriptionID={productSubscription.id}
                                    onGenerate={onLicenseUpdate}
                                />
                            </CardBody>
                        )}
                        <FilteredSiteAdminProductLicenseConnection
                            className="list-group list-group-flush"
                            noun="product license"
                            pluralNoun="product licenses"
                            queryConnection={queryProductLicensesForSubscription}
                            nodeComponent={SiteAdminProductLicenseNode}
                            nodeComponentProps={nodeProps}
                            compact={true}
                            hideSearch={true}
                            noSummaryIfAllNodesVisible={true}
                            updates={licenseUpdates}
                            history={history}
                            location={location}
                        />
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
            query DotComProductSubscription($uuid: String!) {
                dotcom {
                    productSubscription(uuid: $uuid) {
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
                        productLicenses {
                            nodes {
                                id
                                info {
                                    tags
                                    userCount
                                    expiresAt
                                }
                                licenseKey
                                createdAt
                            }
                            totalCount
                            pageInfo {
                                hasNextPage
                            }
                        }
                        createdAt
                        isArchived
                        url
                        urlForSiteAdminBilling
                    }
                }
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

function queryProductLicenses(
    subscriptionUUID: string,
    args: { first?: number }
): Observable<GQL.IProductLicenseConnection> {
    return queryGraphQL(
        gql`
            query ProductLicenses($first: Int, $subscriptionUUID: String!) {
                dotcom {
                    productSubscription(uuid: $subscriptionUUID) {
                        productLicenses(first: $first) {
                            nodes {
                                ...ProductLicenseFields
                            }
                            totalCount
                            pageInfo {
                                hasNextPage
                            }
                        }
                    }
                }
            }
            ${siteAdminProductLicenseFragment}
        `,
        {
            first: args.first,
            subscriptionUUID,
        }
    ).pipe(
        map(({ data, errors }) => {
            if (
                !data ||
                !data.dotcom ||
                !data.dotcom.productSubscription ||
                !data.dotcom.productSubscription.productLicenses ||
                (errors && errors.length > 0)
            ) {
                throw createAggregateError(errors)
            }
            return data.dotcom.productSubscription.productLicenses
        })
    )
}

function archiveProductSubscription(args: ArchiveProductSubscriptionVariables): Observable<void> {
    return requestGraphQL<ArchiveProductSubscriptionResult, ArchiveProductSubscriptionVariables>(
        gql`
            mutation ArchiveProductSubscription($id: ID!) {
                dotcom {
                    archiveProductSubscription(id: $id) {
                        alwaysNil
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.dotcom || !data.dotcom.archiveProductSubscription || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
        })
    )
}
