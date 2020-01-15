import * as React from 'react'
import { LinkOrSpan } from '../../../../../../shared/src/components/LinkOrSpan'
import { gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { CopyableText } from '../../../../components/CopyableText'
import { Timestamp } from '../../../../components/time/Timestamp'
import { AccountName } from '../../../dotcom/productSubscriptions/AccountName'
import { ProductLicenseValidity } from '../../../dotcom/productSubscriptions/ProductLicenseValidity'
import { ProductLicenseInfoDescription } from '../../../productSubscription/ProductLicenseInfoDescription'

export const siteAdminProductLicenseFragment = gql`
    fragment ProductLicenseFields on ProductLicense {
        id
        subscription {
            id
            name
            account {
                id
                username
                displayName
            }
            activeLicense {
                id
            }
            urlForSiteAdmin
        }
        licenseKey
        info {
            productNameWithBrand
            tags
            userCount
            expiresAt
        }
        createdAt
    }
`

export interface SiteAdminProductLicenseNodeProps {
    node: GQL.IProductLicense
    showSubscription: boolean
    onDidUpdate: () => void
}

/**
 * Displays a product license in a connection in the site admin area.
 */
export class SiteAdminProductLicenseNode extends React.PureComponent<SiteAdminProductLicenseNodeProps> {
    public render(): JSX.Element | null {
        return (
            <li className="list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    {that.props.showSubscription && (
                        <div className="mr-3 text-truncate">
                            <strong>
                                License in{' '}
                                <LinkOrSpan to={that.props.node.subscription.urlForSiteAdmin} className="mr-3">
                                    {that.props.node.subscription.name}
                                </LinkOrSpan>
                            </strong>
                            <span className="mr-3">
                                <AccountName account={that.props.node.subscription.account} />
                            </span>
                        </div>
                    )}
                    <div>
                        {that.props.node.info && (
                            <ProductLicenseInfoDescription licenseInfo={that.props.node.info} className="mr-3" />
                        )}
                        {that.props.node.info &&
                        that.props.node.subscription.activeLicense &&
                        that.props.node.subscription.activeLicense.id === that.props.node.id ? (
                            <ProductLicenseValidity
                                licenseInfo={that.props.node.info}
                                primary={false}
                                className="d-inline-block mr-3"
                            />
                        ) : (
                            <span
                                className="text-warning font-weight-bold mr-3"
                                data-tooltip="A newer license was generated for this subscription. This license should no longer be used."
                            >
                                Inactive
                            </span>
                        )}
                        <span className="text-muted">
                            Created <Timestamp date={that.props.node.createdAt} />
                        </span>
                    </div>
                </div>
                {that.props.node.info && that.props.node.info.tags.length > 0 && (
                    <div>
                        Tags:{' '}
                        {that.props.node.info.tags.map(tag => (
                            <div className="mr-1 badge badge-secondary" key={tag}>
                                {tag}
                            </div>
                        ))}
                    </div>
                )}
                <CopyableText text={that.props.node.licenseKey} className="mt-2 d-block" />
            </li>
        )
    }
}
