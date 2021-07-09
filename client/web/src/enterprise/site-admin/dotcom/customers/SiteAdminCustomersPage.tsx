import React, { useEffect, useMemo, useCallback, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject } from 'rxjs'
import { map } from 'rxjs/operators'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { createAggregateError } from '@sourcegraph/shared/src/util/errors'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionForm,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '@sourcegraph/web/src/components/FilteredConnection/generic-ui'
import { hasNextPage } from '@sourcegraph/web/src/components/FilteredConnection/utils'

import { queryGraphQL } from '../../../../backend/graphql'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { CustomerFields, CustomersResult, CustomersVariables } from '../../../../graphql-operations'
import { eventLogger } from '../../../../tracking/eventLogger'
import { userURL } from '../../../../user'
import { usePaginatedConnection } from '../../../../user/settings/accessTokens/usePaginatedConnection'
import { AccountName } from '../../../dotcom/productSubscriptions/AccountName'

import { SiteAdminCustomerBillingLink } from './SiteAdminCustomerBillingLink'

const siteAdminCustomerFragment = gql`
    fragment CustomerFields on User {
        id
        username
        displayName
        urlForSiteAdminBilling
    }
`

interface SiteAdminCustomerNodeProps {
    node: Pick<GQL.IUser, 'id' | 'username' | 'displayName' | 'urlForSiteAdminBilling'>
    onDidUpdate: () => void
}

/**
 * Displays a customer in a connection in the site admin area.
 */
const SiteAdminCustomerNode: React.FunctionComponent<SiteAdminCustomerNodeProps> = ({ node, onDidUpdate }) => (
    <li className="list-group-item py-2">
        <div className="d-flex align-items-center justify-content-between">
            <span className="mr-3">
                <AccountName account={node} link={`${userURL(node.username)}/subscriptions`} />
            </span>
            <SiteAdminCustomerBillingLink customer={node} onDidUpdate={onDidUpdate} />
        </div>
    </li>
)

interface Props extends RouteComponentProps<{}> {}

/**
 * Displays a list of customers associated with user accounts on Sourcegraph.com.
 */
export const SiteAdminProductCustomersPage: React.FunctionComponent<Props> = props => {
    useEffect(() => eventLogger.logViewEvent('SiteAdminProductCustomers'), [])

    const updates = useMemo(() => new Subject<void>(), [])
    const onUserUpdate = useCallback(() => updates.next(), [updates])
    const nodeProps: Pick<SiteAdminCustomerNodeProps, Exclude<keyof SiteAdminCustomerNodeProps, 'node'>> = {
        onDidUpdate: onUserUpdate,
    }
    const [query, setQuery] = useState('')
    const { connection, errors, loading, fetchMore, hasNextPage } = usePaginatedConnection<
        CustomersResult,
        CustomersVariables,
        CustomerFields
    >({
        query: CUSTOMERS,
        variables: {
            first: 20,
            query,
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            return data.users
        },
        options: {
            useURLQuery: true,
        },
    })

    return (
        <div className="site-admin-customers-page">
            <PageTitle title="Customers" />
            <div className="d-flex justify-content-between align-items-center mb-1">
                <h2 className="mb-0">Customers</h2>
            </div>
            <p>User accounts may be linked to a customer on the billing system.</p>
            <ConnectionContainer className="list-group list-group-flush mt-3">
                <ConnectionForm
                    query={query}
                    onChange={event => setQuery(event.target.value)}
                    inputPlaceholder="Search customers..."
                />
                {errors.length > 0 && <ConnectionError errors={errors} />}
                <ConnectionList>
                    {connection?.nodes?.map(node => (
                        <SiteAdminCustomerNode key={node.id} {...nodeProps} node={node} />
                    ))}
                </ConnectionList>
                {loading && <ConnectionLoading />}
                {connection && (
                    <SummaryContainer>
                        <ConnectionSummary
                            noSummaryIfAllNodesVisible={true}
                            connection={connection}
                            noun="customer"
                            pluralNoun="customers"
                            totalCount={connection.totalCount ?? null}
                            hasNextPage={hasNextPage}
                        />
                        {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </div>
    )
}

export const CUSTOMERS = gql`
    query Customers($first: Int, $query: String) {
        users(first: $first, query: $query) {
            nodes {
                ...CustomerFields
            }
            totalCount
            pageInfo {
                hasNextPage
            }
        }
    }
    ${siteAdminCustomerFragment}
`
