import { gql } from '@sourcegraph/http-client'

export const USER_PRODUCT_SUBSCRIPTION = gql`
    query UserProductSubscription($uuid: String!) {
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
        currentSourcegraphAccessToken
        llmProxyAccess {
            ...LLMProxyAccessFields
        }
    }

    fragment LLMProxyAccessFields on LLMProxyAccess {
        enabled
        rateLimit {
            ...LLMProxyRateLimitFields
        }
        usage {
            date
            count
        }
    }

    fragment LLMProxyRateLimitFields on LLMProxyRateLimit {
        source
        limit
        intervalSeconds
    }
`
