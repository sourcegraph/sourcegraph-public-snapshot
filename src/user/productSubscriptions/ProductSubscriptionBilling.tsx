import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import React from 'react'
import { ProductSubscriptionLabel } from '../../dotcom/productSubscriptions/ProductSubscriptionLabel'
import { mailtoSales } from '../../productSubscription/helpers'

export const ProductSubscriptionBilling: React.SFC<{
    productSubscription: GQL.IProductSubscription
}> = ({ productSubscription }) => (
    <table className="table mb-0">
        <tbody>
            <tr>
                <th className="text-nowrap align-middle">Plan</th>
                <td className="w-100 d-flex align-items-center justify-content-between">
                    <ProductSubscriptionLabel productSubscription={productSubscription} planField="title" />
                    <a
                        href={mailtoSales({
                            subject: `Change subscription ${productSubscription.name}`,
                        })}
                        className="btn btn-secondary btn-sm"
                    >
                        Change plan or add/remove users
                    </a>
                </td>
            </tr>
        </tbody>
    </table>
)
