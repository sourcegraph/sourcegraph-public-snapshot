import * as React from 'react'
import { gql } from '../../../../../packages/webapp/src/backend/graphql'
import * as GQL from '../../../../../packages/webapp/src/backend/graphqlschema'
import { LinkOrSpan } from '../../../../../packages/webapp/src/components/LinkOrSpan'
import { AccountName } from '../../../dotcom/productSubscriptions/AccountName'
import { ProductSubscriptionLabel } from '../../../dotcom/productSubscriptions/ProductSubscriptionLabel'

export const siteAdminProductSubscriptionFragment = gql`
    fragment ProductSubscriptionFields on ProductSubscription {
        id
        name
        account {
            id
            username
            displayName
        }
        invoiceItem {
            plan {
                nameWithBrand
            }
            userCount
            expiresAt
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

export const SiteAdminProductSubscriptionNodeHeader: React.SFC<{ nodes: any }> = () => (
    <thead>
        <tr>
            <th>ID</th>
            <th>Plan</th>
            <th>Customer</th>
        </tr>
    </thead>
)

export interface SiteAdminProductSubscriptionNodeProps {
    node: GQL.IProductSubscription
    onDidUpdate: () => void
}

/**
 * Displays a product subscription in a connection in the site admin area.
 */
export class SiteAdminProductSubscriptionNode extends React.PureComponent<SiteAdminProductSubscriptionNodeProps> {
    public render(): JSX.Element | null {
        return (
            <tr>
                <td className="text-nowrap">
                    <LinkOrSpan to={this.props.node.urlForSiteAdmin} className="mr-3">
                        {this.props.node.name}
                    </LinkOrSpan>
                </td>
                <td className="text-nowrap">
                    <ProductSubscriptionLabel productSubscription={this.props.node} className="mr-3" />
                </td>
                <td className="w-100">
                    <AccountName account={this.props.node.account} />
                </td>
            </tr>
        )
    }
}
