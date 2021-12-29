import React from 'react'

import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { RouterLink } from '@sourcegraph/wildcard'

import { ProductSubscriptionLabel } from '../../dotcom/productSubscriptions/ProductSubscriptionLabel'

export const ProductSubscriptionBilling: React.FunctionComponent<{
    productSubscription: GQL.IProductSubscription
}> = ({ productSubscription }) => (
    <table className="table mb-0">
        <tbody>
            <tr>
                <th className="text-nowrap align-middle">Plan</th>
                <td className="w-100 d-flex align-items-center justify-content-between">
                    <ProductSubscriptionLabel productSubscription={productSubscription} planField="name" />
                    <RouterLink to={`${productSubscription.url}/edit`} className="btn btn-secondary btn-sm">
                        Change plan or add/remove users
                    </RouterLink>
                </td>
            </tr>
        </tbody>
    </table>
)
