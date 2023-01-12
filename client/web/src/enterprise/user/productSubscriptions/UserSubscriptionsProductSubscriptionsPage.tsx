import React, { useEffect, useCallback } from 'react'

import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { Container, PageHeader, Link, Text } from '@sourcegraph/wildcard'

import { queryGraphQL } from '../../../backend/graphql'
import { FilteredConnection } from '../../../components/FilteredConnection'
import { PageTitle } from '../../../components/PageTitle'
import {
    ProductSubscriptionFields,
    ProductSubscriptionsResult,
    ProductSubscriptionsVariables,
    UserAreaUserFields,
} from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import {
    productSubscriptionFragment,
    ProductSubscriptionNode,
    ProductSubscriptionNodeHeader,
    ProductSubscriptionNodeProps,
} from '../../dotcom/productSubscriptions/ProductSubscriptionNode'

interface Props extends RouteComponentProps<{}> {
    user: UserAreaUserFields
}

class FilteredProductSubscriptionConnection extends FilteredConnection<
    ProductSubscriptionFields,
    ProductSubscriptionNodeProps
> {}

/**
 * Displays the product subscriptions associated with this account.
 */
export const UserSubscriptionsProductSubscriptionsPage: React.FunctionComponent<
    React.PropsWithChildren<Props>
> = props => {
    useEffect(() => {
        eventLogger.logViewEvent('UserSubscriptionsProductSubscriptions')
    }, [])

    const queryLicenses = useCallback(
        (args: { first?: number }): Observable<ProductSubscriptionsResult['dotcom']['productSubscriptions']> => {
            const variables: ProductSubscriptionsVariables = {
                first: args.first ?? null,
                account: props.user.id,
            }
            return queryGraphQL<ProductSubscriptionsResult>(
                gql`
                    query ProductSubscriptions($first: Int, $account: ID) {
                        dotcom {
                            productSubscriptions(first: $first, account: $account) {
                                nodes {
                                    ...ProductSubscriptionFields
                                }
                                totalCount
                                pageInfo {
                                    hasNextPage
                                }
                            }
                        }
                    }
                    ${productSubscriptionFragment}
                `,
                variables
            ).pipe(
                map(({ data, errors }) => {
                    if (!data || !data.dotcom || !data.dotcom.productSubscriptions || (errors && errors.length > 0)) {
                        throw createAggregateError(errors)
                    }
                    return data.dotcom.productSubscriptions
                })
            )
        },
        [props.user.id]
    )

    return (
        <div className="user-subscriptions-product-subscriptions-page">
            <PageTitle title="Subscriptions" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Subscriptions' }]}
                description={
                    <>
                        <>
                            Search your private code with{' '}
                            <Link
                                to="https://signup.sourcegraph.com/"
                                onClick={() => eventLogger.log('ClickedOnCloudCTA', { cloudCtaType: 'Subscriptions' })}
                            >
                                Sourcegraph Cloud
                            </Link>{' '}
                            or contact us to purchase a subscription for a self-hosted Sourcegraph instance. See{' '}
                            <Link to="https://about.sourcegraph.com/pricing">pricing</Link> for more information.
                        </>
                    </>
                }
                className="mb-3"
            />
            <Container className="mb-3">
                <FilteredProductSubscriptionConnection
                    listComponent="table"
                    listClassName="table mb-0"
                    noun="subscription"
                    pluralNoun="subscriptions"
                    queryConnection={queryLicenses}
                    headComponent={ProductSubscriptionNodeHeader}
                    nodeComponent={ProductSubscriptionNode}
                    hideSearch={true}
                    noSummaryIfAllNodesVisible={true}
                    history={props.history}
                    location={props.location}
                    emptyElement={
                        <Text alignment="center" className="w-100 mb-0 text-muted">
                            You have no subscriptions.
                        </Text>
                    }
                />
            </Container>
        </div>
    )
}
