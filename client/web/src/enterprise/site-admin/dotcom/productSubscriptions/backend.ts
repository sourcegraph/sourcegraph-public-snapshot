import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { queryGraphQL } from '../../../../backend/graphql'
import {
    type UseShowMorePaginationResult,
    useShowMorePagination,
} from '../../../../components/FilteredConnection/hooks/useShowMorePagination'
import type {
    ProductLicensesResult,
    ProductLicenseFields,
    ProductLicensesVariables,
    ProductSubscriptionsDotComResult,
    ProductSubscriptionsDotComVariables,
    DotComProductLicensesResult,
    DotComProductLicensesVariables,
} from '../../../../graphql-operations'

const siteAdminProductSubscriptionFragment = gql`
    fragment SiteAdminProductSubscriptionFields on ProductSubscription {
        id
        name
        uuid
        account {
            id
            username
            displayName
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
        createdAt
        isArchived
        urlForSiteAdmin
    }
`

export const CODY_GATEWAY_ACCESS_FIELDS_FRAGMENT = gql`
    fragment CodyGatewayAccessFields on CodyGatewayAccess {
        enabled
        codeCompletionsRateLimit {
            ...CodyGatewayRateLimitFields
        }
        chatCompletionsRateLimit {
            ...CodyGatewayRateLimitFields
        }
        embeddingsRateLimit {
            ...CodyGatewayRateLimitFields
        }
    }

    fragment CodyGatewayRateLimitFields on CodyGatewayRateLimit {
        allowedModels
        source
        limit
        intervalSeconds
    }
`

const CODY_GATEWAY_RATE_LIMIT_USAGE_FIELDS = gql`
    fragment CodyGatewayRateLimitUsageFields on CodyGatewayRateLimit {
        usage {
            ...CodyGatewayRateLimitUsageDatapoint
        }
    }

    fragment CodyGatewayRateLimitUsageDatapoint on CodyGatewayUsageDatapoint {
        date
        count
        model
    }
`

export const CODY_GATEWAY_ACCESS_COMPLETIONS_USAGE_FIELDS_FRAGMENT = gql`
    fragment CodyGatewayAccessCompletionsUsageFields on CodyGatewayAccess {
        codeCompletionsRateLimit {
            ...CodyGatewayRateLimitUsageFields
        }
        chatCompletionsRateLimit {
            ...CodyGatewayRateLimitUsageFields
        }
    }

    ${CODY_GATEWAY_RATE_LIMIT_USAGE_FIELDS}
`

export const DOTCOM_PRODUCT_SUBSCRIPTION_CODY_GATEWAY_COMPLETIONS_USAGE = gql`
    query DotComProductSubscriptionCodyGatewayCompletionsUsage($uuid: String!) {
        dotcom {
            productSubscription(uuid: $uuid) {
                id
                codyGatewayAccess {
                    ...CodyGatewayAccessCompletionsUsageFields
                }
            }
        }
    }

    ${CODY_GATEWAY_ACCESS_COMPLETIONS_USAGE_FIELDS_FRAGMENT}
`

export const CODY_GATEWAY_ACCESS_EMBEDDINGS_USAGE_FIELDS_FRAGMENT = gql`
    fragment CodyGatewayAccessEmbeddingsUsageFields on CodyGatewayAccess {
        embeddingsRateLimit {
            ...CodyGatewayRateLimitUsageFields
        }
    }

    ${CODY_GATEWAY_RATE_LIMIT_USAGE_FIELDS}
`

export const DOTCOM_PRODUCT_SUBSCRIPTION_CODY_GATEWAY_EMBEDDINGS_USAGE = gql`
    query DotComProductSubscriptionCodyGatewayEmbeddingsUsage($uuid: String!) {
        dotcom {
            productSubscription(uuid: $uuid) {
                id
                codyGatewayAccess {
                    ...CodyGatewayAccessEmbeddingsUsageFields
                }
            }
        }
    }

    ${CODY_GATEWAY_ACCESS_EMBEDDINGS_USAGE_FIELDS_FRAGMENT}
`

export const DOTCOM_PRODUCT_SUBSCRIPTION = gql`
    query DotComProductSubscription($uuid: String!) {
        dotcom {
            productSubscription(uuid: $uuid) {
                id
                name
                account {
                    id
                    username
                    displayName
                }
                productLicenses {
                    nodes {
                        id
                        info {
                            tags
                            userCount
                            expiresAt
                            salesforceSubscriptionID
                            salesforceOpportunityID
                        }
                        licenseKey
                        createdAt
                        revokedAt
                        revokeReason
                        siteID
                        version
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
                        salesforceSubscriptionID
                        salesforceOpportunityID
                    }
                    licenseKey
                    createdAt
                }
                currentSourcegraphAccessToken
                codyGatewayAccess {
                    ...CodyGatewayAccessFields
                }
                createdAt
                isArchived
                url
            }
        }
    }

    ${CODY_GATEWAY_ACCESS_FIELDS_FRAGMENT}
`

export const ARCHIVE_PRODUCT_SUBSCRIPTION = gql`
    mutation ArchiveProductSubscription($id: ID!) {
        dotcom {
            archiveProductSubscription(id: $id) {
                alwaysNil
            }
        }
    }
`

export const UPDATE_CODY_GATEWAY_CONFIG = gql`
    mutation UpdateCodyGatewayConfig($productSubscriptionID: ID!, $access: UpdateCodyGatewayAccessInput!) {
        dotcom {
            updateProductSubscription(id: $productSubscriptionID, update: { codyGatewayAccess: $access }) {
                alwaysNil
            }
        }
    }
`

export const GENERATE_PRODUCT_LICENSE = gql`
    mutation GenerateProductLicenseForSubscription($productSubscriptionID: ID!, $license: ProductLicenseInput!) {
        dotcom {
            generateProductLicenseForSubscription(productSubscriptionID: $productSubscriptionID, license: $license) {
                id
            }
        }
    }
`

const siteAdminProductLicenseFragment = gql`
    fragment ProductLicenseFields on ProductLicense {
        id
        subscription {
            id
            uuid
            name
            account {
                ...ProductLicenseSubscriptionAccount
            }
            activeLicense {
                id
            }
            urlForSiteAdmin
        }
        licenseKey
        siteID
        info {
            ...ProductLicenseInfoFields
        }
        createdAt
        revokedAt
        revokeReason
        version
    }

    fragment ProductLicenseInfoFields on ProductLicenseInfo {
        productNameWithBrand
        tags
        userCount
        expiresAt
        salesforceSubscriptionID
        salesforceOpportunityID
    }

    fragment ProductLicenseSubscriptionAccount on User {
        id
        username
        displayName
    }
`

export const PRODUCT_LICENSES = gql`
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
`

export const useProductSubscriptionLicensesConnection = (
    subscriptionUUID: string,
    first: number
): UseShowMorePaginationResult<ProductLicensesResult, ProductLicenseFields> =>
    useShowMorePagination<ProductLicensesResult, ProductLicensesVariables, ProductLicenseFields>({
        query: PRODUCT_LICENSES,
        variables: {
            first: first ?? 20,
            after: null,
            subscriptionUUID,
        },
        getConnection: result => {
            const { dotcom } = dataOrThrowErrors(result)
            return dotcom.productSubscription.productLicenses
        },
        options: {
            fetchPolicy: 'cache-and-network',
        },
    })

export function queryProductSubscriptions(args: {
    first?: number
    query?: string
}): Observable<ProductSubscriptionsDotComResult['dotcom']['productSubscriptions']> {
    return queryGraphQL<ProductSubscriptionsDotComResult>(
        gql`
            query ProductSubscriptionsDotCom($first: Int, $account: ID, $query: String) {
                dotcom {
                    productSubscriptions(first: $first, account: $account, query: $query) {
                        nodes {
                            ...SiteAdminProductSubscriptionFields
                        }
                        totalCount
                        pageInfo {
                            hasNextPage
                        }
                    }
                }
            }
            ${siteAdminProductSubscriptionFragment}
        `,
        {
            first: args.first,
            query: args.query,
        } as Partial<ProductSubscriptionsDotComVariables>
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.dotcom.productSubscriptions)
    )
}

const QUERY_PRODUCT_LICENSES = gql`
    query DotComProductLicenses($first: Int, $licenseKeySubstring: String) {
        dotcom {
            productLicenses(first: $first, licenseKeySubstring: $licenseKeySubstring) {
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
    ${siteAdminProductLicenseFragment}
`

export const useQueryProductLicensesConnection = (
    licenseKeySubstring: string,
    first: number
): UseShowMorePaginationResult<DotComProductLicensesResult, ProductLicenseFields> =>
    useShowMorePagination<DotComProductLicensesResult, DotComProductLicensesVariables, ProductLicenseFields>({
        query: QUERY_PRODUCT_LICENSES,
        variables: {
            first: first ?? 20,
            after: null,
            licenseKeySubstring,
        },
        getConnection: result => {
            const { dotcom } = dataOrThrowErrors(result)
            return dotcom.productLicenses
        },
        options: {
            fetchPolicy: 'cache-and-network',
            skip: !licenseKeySubstring,
        },
    })

export const REVOKE_LICENSE = gql`
    mutation RevokeLicense($id: ID!, $reason: String!) {
        dotcom {
            revokeLicense(id: $id, reason: $reason) {
                alwaysNil
            }
        }
    }
`
