import * as React from 'react'
import { LinkOrSpan } from '../../../../../../shared/src/components/LinkOrSpan'
import { gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { Timestamp } from '../../../../components/time/Timestamp'
import { AccountName } from '../../../dotcom/productSubscriptions/AccountName'
import { ProductSubscriptionLabel } from '../../../dotcom/productSubscriptions/ProductSubscriptionLabel'
import { ProductLicenseTags } from '../../../productSubscription/ProductLicenseTags'

export const siteAdminProductSubscriptionFragment = gql`
    fragment ProductSubscriptionFields on ProductSubscription {
        id
        name
        account {
            id
            username
            displayName
            emails {
                email
                isPrimary
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

export const SiteAdminProductSubscriptionNodeHeader: React.FunctionComponent<{ nodes: any }> = () => (
    <thead>
        <tr>
            <th>ID</th>
            <th>Customer</th>
            <th>Plan</th>
            <th>Expiration</th>
            <th>Tags</th>
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
                <td className="w-100">
                    <AccountName account={this.props.node.account} />
                    {this.props.node.account && (
                        <div>
                            <small>
                                {this.props.node.account.emails
                                    .filter(email => email.isPrimary)
                                    .map(({ email }) => email)
                                    .join(', ')}
                            </small>
                        </div>
                    )}
                </td>
                <td className="text-nowrap">
                    <ProductSubscriptionLabel productSubscription={this.props.node} className="mr-3" />
                </td>
                <td className="text-nowrap">
                    {this.props.node.activeLicense?.info ? (
                        <Timestamp date={this.props.node.activeLicense.info.expiresAt} />
                    ) : (
                        <span className="text-muted font-italic">None</span>
                    )}
                </td>
                <td className="w-100">
                    {this.props.node.activeLicense &&
                    this.props.node.activeLicense.info &&
                    this.props.node.activeLicense.info.tags.length > 0 ? (
                        <ProductLicenseTags tags={this.props.node.activeLicense.info.tags} />
                    ) : (
                        <span className="text-muted font-italic">None</span>
                    )}
                </td>
            </tr>
        )
    }
}
