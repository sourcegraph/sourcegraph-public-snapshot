import React from 'react'

import * as GQL from '@sourcegraph/shared/src/schema'
import { Button, Link } from '@sourcegraph/wildcard'

import { ProductSubscriptionLabel } from '../../dotcom/productSubscriptions/ProductSubscriptionLabel'

export const ProductSubscriptionBilling: React.FunctionComponent<
    React.PropsWithChildren<{
        productSubscription: GQL.IProductSubscription
    }>
> = ({ productSubscription }) => (
    <table className="table mb-0">
        <tbody>
            <tr>
                <th className="text-nowrap align-middle">Plan</th>
                <td className="w-100 d-flex align-items-center justify-content-between">
                    <ProductSubscriptionLabel productSubscription={productSubscription} planField="name" />
                    <Button to={`${productSubscription.url}/edit`} variant="secondary" size="sm" as={Link}>
                        Change plan or add/remove users
                    </Button>
                </td>
            </tr>
        </tbody>
    </table>
)
