import * as React from 'react'
import { Link } from 'react-router-dom'
import { gql } from '../../../../packages/webapp/src/backend/graphql'
import * as GQL from '../../../../packages/webapp/src/backend/graphqlschema'
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

export const ProductSubscriptionNodeHeader: React.SFC<{ nodes: any }> = () => (
    <thead>
        <tr>
            <th>ID</th>
            <th>Plan</th>
        </tr>
    </thead>
)

export interface ProductSubscriptionNodeProps {
    node: GQL.IProductSubscription
    onDidUpdate: () => void
}

export class ProductSubscriptionNode extends React.PureComponent<ProductSubscriptionNodeProps> {
    public render(): JSX.Element | null {
        return (
            <tr>
                <td className="text-nowrap">
                    <Link to={this.props.node.url} className="mr-3 font-weight-bold">
                        {this.props.node.name}
                    </Link>
                </td>
                <td className="w-100">
                    <ProductSubscriptionLabel productSubscription={this.props.node} className="mr-3" />
                </td>
            </tr>
        )
    }
}
