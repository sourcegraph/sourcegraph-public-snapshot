import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, of, Subject, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { eventLogger } from '../../../../tracking/eventLogger'
import {
    siteAdminProductLicenseFragment,
    SiteAdminProductLicenseNode,
    SiteAdminProductLicenseNodeProps,
} from './SiteAdminProductLicenseNode'

interface Props extends RouteComponentProps<{}> {}

class FilteredProductLicenseConnection extends FilteredConnection<
    GQL.IProductLicense,
    Pick<SiteAdminProductLicenseNodeProps, 'onDidUpdate' | 'showSubscription'>
> {}

/**
 * Displays the product licenses that have been created on Sourcegraph.com.
 */
export class SiteAdminProductLicensesPage extends React.Component<Props> {
    private subscriptions = new Subscription()
    private updates = new Subject<void>()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminProductLicenses')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<SiteAdminProductLicenseNodeProps, 'onDidUpdate' | 'showSubscription'> = {
            onDidUpdate: this.onDidUpdateProductLicense,
            showSubscription: true,
        }

        return (
            <div className="site-admin-product-subscriptions-page">
                <PageTitle title="Product subscriptions" />
                <div className="d-flex justify-content-between align-items-center mt-3 mb-1">
                    <h2 className="mb-0">License key lookup</h2>
                </div>
                <p>Find matching licenses and their associated product subscriptions.</p>
                <FilteredProductLicenseConnection
                    className="list-group list-group-flush mt-3"
                    noun="product license"
                    pluralNoun="product licenses"
                    queryConnection={this.queryLicenses}
                    nodeComponent={SiteAdminProductLicenseNode}
                    nodeComponentProps={nodeProps}
                    emptyElement={<span className="text-muted mt-2">Enter a partial license key to find matches.</span>}
                    noSummaryIfAllNodesVisible={true}
                    autoFocus={true}
                    updates={this.updates}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryLicenses = (args: { first?: number; query?: string }): Observable<GQL.IProductLicenseConnection> => {
        const vars: GQL.IProductLicensesOnDotcomQueryArguments = {
            first: args.first,
            licenseKeySubstring: args.query,
        }
        return args.query
            ? queryGraphQL(
                  gql`
                      query ProductLicenses($first: Int, $licenseKeySubstring: String) {
                          dotcom {
                              productLicenses(first: $first, licenseKeySubstring: $licenseKeySubstring) {
                                  nodes {
                                      ...ProductLicenseFields
                                  }
                                  totalCount
                                  pageInfo {
                                      hasNextPage
                                  }
                              }
                          }
                      }
                      ${siteAdminProductLicenseFragment}
                  `,
                  vars
              ).pipe(
                  map(({ data, errors }) => {
                      if (!data || !data.dotcom || !data.dotcom.productLicenses || (errors && errors.length > 0)) {
                          throw createAggregateError(errors)
                      }
                      return data.dotcom.productLicenses
                  })
              )
            : of({
                  __typename: 'ProductLicenseConnection' as const,
                  nodes: [],
                  totalCount: 0,
                  pageInfo: { __typename: 'PageInfo' as const, hasNextPage: false, endCursor: null },
              })
    }

    private onDidUpdateProductLicense = (): void => this.updates.next()
}
