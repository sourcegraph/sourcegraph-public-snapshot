import { gql } from '@sourcegraph/http-client'

import { LLM_PROXY_ACCESS_FIELDS_FRAGMENT } from '../../site-admin/dotcom/productSubscriptions/backend'

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

    ${LLM_PROXY_ACCESS_FIELDS_FRAGMENT}
`
