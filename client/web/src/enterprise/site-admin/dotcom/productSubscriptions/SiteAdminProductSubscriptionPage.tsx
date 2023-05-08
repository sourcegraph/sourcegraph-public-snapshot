import React, { useState, useMemo, useEffect, useCallback } from 'react'

import { gql as apolloGQL, useMutation } from '@apollo/client'
import { mdiArrowLeft, mdiPlus } from '@mdi/js'
import { useNavigate, useParams } from 'react-router-dom'
import { Observable, Subject, NEVER } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap, tap, filter } from 'rxjs/operators'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { asError, createAggregateError, isErrorLike } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
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
    H2,
    ErrorAlert,
    Text,
    Checkbox,
    H3,
} from '@sourcegraph/wildcard'

import { queryGraphQL, requestGraphQL } from '../../../../backend/graphql'
import { CopyableText } from '../../../../components/CopyableText'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { useFeatureFlag } from '../../../../featureFlags/useFeatureFlag'
import {
    ArchiveProductSubscriptionResult,
    ArchiveProductSubscriptionVariables,
    DotComProductSubscriptionResult,
    ProductLicensesResult,
    ProductLicenseFields,
    GenerateProductSubscriptionAccessTokenResult,
    GenerateProductSubscriptionAccessTokenVariables,
} from '../../../../graphql-operations'
import { eventLogger } from '../../../../tracking/eventLogger'
import { AccountEmailAddresses } from '../../../dotcom/productSubscriptions/AccountEmailAddresses'
import { AccountName } from '../../../dotcom/productSubscriptions/AccountName'
import { ProductSubscriptionLabel } from '../../../dotcom/productSubscriptions/ProductSubscriptionLabel'
import { LicenseGenerationKeyWarning } from '../../../productSubscription/LicenseGenerationKeyWarning'

import { SiteAdminGenerateProductLicenseForSubscriptionForm } from './SiteAdminGenerateProductLicenseForSubscriptionForm'
import {
    siteAdminProductLicenseFragment,
    SiteAdminProductLicenseNode,
    SiteAdminProductLicenseNodeProps,
} from './SiteAdminProductLicenseNode'

interface Props {
    /** For mocking in tests only. */
    _queryProductSubscription?: typeof queryProductSubscription

    /** For mocking in tests only. */
    _queryProductLicenses?: typeof queryProductLicenses
}

const LOADING = 'loading' as const

/**
 * Displays a product subscription in the site admin area.
 */
export const SiteAdminProductSubscriptionPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    _queryProductSubscription = queryProductSubscription,
    _queryProductLicenses = queryProductLicenses,
}) => {
    const navigate = useNavigate()
    const { subscriptionUUID = '' } = useParams<{ subscriptionUUID: string }>()
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
                            tap(() => navigate('/site-admin/dotcom/product/subscriptions')),
                            catchError(error => [asError(error)]),
                            startWith(LOADING)
                        )
                    )
                )
            },
            [navigate, productSubscription]
        )
    )

    const queryProductLicensesForSubscription = useCallback(
        (args: { first?: number }) => _queryProductLicenses(subscriptionUUID, args),
        [_queryProductLicenses, subscriptionUUID]
    )

    const toggleShowGenerate = useCallback((): void => setShowGenerate(previousValue => !previousValue), [])

    /** Updates to the subscription's licenses. */
    const licenseUpdates = useMemo(() => new Subject<void>(), [])
    const onLicenseUpdate = useCallback(() => {
        licenseUpdates.next()
        toggleShowGenerate()
    }, [licenseUpdates, toggleShowGenerate])

    const [
        generateAccessTokenMutation,
        { loading: tokenLoading, called: generateTokenCalled, data: tokenData, error: tokenError },
    ] = useMutation<GenerateProductSubscriptionAccessTokenResult, GenerateProductSubscriptionAccessTokenVariables>(
        GENERATE_ACCESS_TOKEN_GQL
    )

    // Feature flag only used as this is under development - will be enabled by default
    const [llmProxyManagementUI] = useFeatureFlag('llm-proxy-management-ui')

    const nodeProps: Pick<SiteAdminProductLicenseNodeProps, 'showSubscription'> = {
        showSubscription: false,
    }

    return (
        <div className="site-admin-product-subscription-page">
            <PageTitle title="Product subscription" />
            <div className="mb-2">
                <Button to="/site-admin/dotcom/product/subscriptions" variant="link" size="sm" as={Link}>
                    <Icon aria-hidden={true} svgPath={mdiArrowLeft} /> All subscriptions
                </Button>
            </div>
            {productSubscription === LOADING ? (
                <LoadingSpinner />
            ) : isErrorLike(productSubscription) ? (
                <ErrorAlert className="my-2" error={productSubscription} />
            ) : (
                <>
                    <H2>Product subscription {productSubscription.name}</H2>
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
                                    <th className="text-nowrap">Created at</th>
                                    <td className="w-100">
                                        <Timestamp date={productSubscription.createdAt} />
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </Card>
                    <Card className="mt-3" hidden={!llmProxyManagementUI}>
                        <CardHeader className="d-flex align-items-center justify-content-between">
                            Access token
                            <Button
                                onClick={() =>
                                    generateAccessTokenMutation({
                                        variables: { productSubscriptionID: productSubscription.id },
                                    })
                                }
                                variant="primary"
                                size="sm"
                                disabled={tokenLoading}
                            >
                                <Icon aria-hidden={true} svgPath={mdiPlus} /> Generate access token
                            </Button>
                        </CardHeader>
                        <CardBody>
                            <Text>Access tokens can be used for LLM-proxy access - coming soon!</Text>
                            {tokenLoading && <LoadingSpinner />}
                            {tokenError && <ErrorAlert className="mt-2" error={tokenError.message} />}
                            {generateTokenCalled && !tokenLoading && tokenData && (
                                <CopyableText
                                    label="Access token"
                                    secret={true}
                                    flex={true}
                                    text={tokenData.dotcom.generateAccessTokenForSubscription.accessToken}
                                    className="mt-2"
                                />
                            )}
                        </CardBody>
                    </Card>
                    <Card className="mt-3" hidden={!llmProxyManagementUI}>
                        <CardHeader>Cody services</CardHeader>
                        <CardBody hidden={!productSubscription.llmProxyAccess.enabled}>
                            <H3>Completions</H3>
                            <Checkbox
                                id="llm-proxy-enabled"
                                checked={productSubscription.llmProxyAccess.enabled}
                                disabled={true}
                                label="Enable access to hosted completions (LLM-proxy)"
                                className="mb-2"
                            />
                            <Text>Rate limits: {JSON.stringify(productSubscription.llmProxyAccess.rateLimit)}</Text>
                        </CardBody>
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
                                    <Icon aria-hidden={true} svgPath={mdiPlus} /> Generate new license manually
                                </Button>
                            )}
                        </CardHeader>
                        {showGenerate && (
                            <CardBody>
                                <SiteAdminGenerateProductLicenseForSubscriptionForm
                                    subscriptionID={productSubscription.id}
                                    subscriptionAccount={productSubscription.account?.username || ''}
                                    onGenerate={onLicenseUpdate}
                                />
                            </CardBody>
                        )}
                        <FilteredConnection<
                            ProductLicenseFields,
                            Pick<SiteAdminProductLicenseNodeProps, 'showSubscription'>
                        >
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
                        />
                    </Card>
                </>
            )}
        </div>
    )
}

function queryProductSubscription(
    uuid: string
): Observable<DotComProductSubscriptionResult['dotcom']['productSubscription']> {
    return queryGraphQL<DotComProductSubscriptionResult>(
        gql`
            query DotComProductSubscription($uuid: String!) {
                dotcom {
                    productSubscription(uuid: $uuid) {
                        id
                        name
                        account {
                            ...DotComProductSubscriptionEmailFields
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
                        activeLicense {
                            id
                            info {
                                productNameWithBrand
                                tags
                                userCount
                                expiresAt
                            }
                            licenseKey
                            createdAt
                        }
                        llmProxyAccess {
                            enabled
                            rateLimit {
                                limit
                                intervalSeconds
                            }
                        }
                        createdAt
                        isArchived
                        url
                    }
                }
            }

            fragment DotComProductSubscriptionEmailFields on User {
                id
                username
                displayName
                emails {
                    email
                    verified
                }
            }
        `,
        { uuid }
    ).pipe(
        map(({ data, errors }) => {
            if (!data?.dotcom?.productSubscription || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return data.dotcom.productSubscription
        })
    )
}

function queryProductLicenses(
    subscriptionUUID: string,
    args: { first?: number }
): Observable<ProductLicensesResult['dotcom']['productSubscription']['productLicenses']> {
    return queryGraphQL<ProductLicensesResult>(
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
            if (!data?.dotcom?.productSubscription?.productLicenses || (errors && errors.length > 0)) {
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
            if (!data?.dotcom?.archiveProductSubscription || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
        })
    )
}

const GENERATE_ACCESS_TOKEN_GQL = apolloGQL`
    mutation GenerateProductSubscriptionAccessToken($productSubscriptionID: ID!) {
        dotcom {
            generateAccessTokenForSubscription(
                productSubscriptionID: $productSubscriptionID
            ) {
                accessToken
            }
        }
    }
`
