import * as React from 'react'

import { gql } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'
import { Link } from '@sourcegraph/wildcard'

import { ProductSubscriptionLabel } from './ProductSubscriptionLabel'

export const productSubscriptionFragment = gql`
    fragment ProductSubscriptionFields on ProductSubscription {
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
                nameWithBrand
            }
            userCount
            expiresAt
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
    }
`

export const ProductSubscriptionNodeHeader: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <thead>
        <tr>
            <th>ID</th>
            <th>Plan</th>
        </tr>
    </thead>
)

export interface ProductSubscriptionNodeProps {
    node: GQL.IProductSubscription
}

export const ProductSubscriptionNode: React.FunctionComponent<
    React.PropsWithChildren<ProductSubscriptionNodeProps>
> = ({ node }) => (
    <tr>
        <td className="text-nowrap">
            <Link to={node.url} className="mr-3 font-weight-bold">
                {node.name}
            </Link>
        </td>
        <td className="w-100">
            <ProductSubscriptionLabel productSubscription={node} className="mr-3" />
        </td>
    </tr>
)
