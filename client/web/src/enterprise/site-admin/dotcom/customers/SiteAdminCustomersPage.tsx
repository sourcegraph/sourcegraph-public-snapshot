import React, { useEffect, useMemo, useCallback, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject } from 'rxjs'
import { map } from 'rxjs/operators'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { createAggregateError } from '@sourcegraph/shared/src/util/errors'
import {
    ConnectionContainer,
    ConnectionList,
    ConnectionSummary,
    ShowMoreButton,
} from '@sourcegraph/web/src/components/FilteredConnection/generic-ui'

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
    console.log('query is', query)
    const { connection, fetchMore } = usePaginatedConnection<CustomersResult, CustomersVariables, CustomerFields>({
        query: CUSTOMERS,
        variables: {
            first: 20,
            query,
        },
        getConnection: data => {
            if (!data || !data.users) {
                throw new Error('bleh')
                // throw createAggregateError(errors)
            }
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
                <Form
                    className="w-100 d-inline-flex justify-content-between flex-row filtered-connection__form"
                    onSubmit={event => event.preventDefault()}
                >
                    <input
                        className="form-control"
                        type="search"
                        placeholder="Search customerz..."
                        name="query"
                        value={query}
                        onChange={event => setQuery(event.target.value)}
                        autoFocus={true}
                        autoComplete="off"
                        autoCorrect="off"
                        autoCapitalize="off"
                        spellCheck={false}
                    />
                </Form>
                <ConnectionList>
                    {connection?.nodes?.map(node => (
                        <SiteAdminCustomerNode key={node.id} {...nodeProps} node={node} />
                    ))}
                </ConnectionList>
                {connection && (
                    <div className="filtered-connection__summary-container">
                        {!connection.pageInfo?.hasNextPage && (
                            // TODO does this hide summary if all nodes visible?
                            <ConnectionSummary
                                connection={connection}
                                noun="customer"
                                pluralNoun="customers"
                                totalCount={connection.totalCount}
                            />
                        )}
                        {connection?.pageInfo?.hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                    </div>
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
