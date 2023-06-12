import React, { useEffect, useState } from 'react'

import { useSearchParams } from 'react-router-dom'

import { Container, H2, Text } from '@sourcegraph/wildcard'

import {
    ConnectionContainer,
    ConnectionError,
    ConnectionLoading,
    ConnectionList,
    SummaryContainer,
    ConnectionSummary,
    ShowMoreButton,
    ConnectionForm,
} from '../../../../components/FilteredConnection/ui'
import { PageTitle } from '../../../../components/PageTitle'
import { eventLogger } from '../../../../tracking/eventLogger'

import { useQueryProductLicensesConnection } from './backend'
import { SiteAdminProductLicenseNode } from './SiteAdminProductLicenseNode'

interface Props {}

const SEARCH_PARAM_KEY = 'query'

/**
 * Displays the product licenses that have been created on Sourcegraph.com.
 */
export const SiteAdminLicenseKeyLookupPage: React.FunctionComponent<React.PropsWithChildren<Props>> = () => {
    useEffect(() => eventLogger.logPageView('SiteAdminLicenseKeyLookup'), [])

    const [searchParams, setSearchParams] = useSearchParams()

    const [search, setSearch] = useState<string>(searchParams.get(SEARCH_PARAM_KEY) ?? '')

    const { loading, hasNextPage, fetchMore, refetchAll, connection, error } = useQueryProductLicensesConnection(
        search,
        20
    )

    useEffect(() => {
        const query = search?.trim() ?? ''
        searchParams.set(SEARCH_PARAM_KEY, query)
        setSearchParams(searchParams)
    }, [search, searchParams, setSearchParams])

    return (
        <div className="site-admin-product-subscriptions-page">
            <PageTitle title="Product subscriptions" />
            <H2>License key lookup</H2>
            <Text>Find matching licenses and their associated product subscriptions.</Text>
            <ConnectionContainer>
                {error && <ConnectionError errors={[error.message]} />}
                {loading && !connection && <ConnectionLoading />}
                <ConnectionForm
                    inputValue={search}
                    onInputChange={event => {
                        const search = event.target.value
                        setSearch(search)
                    }}
                    inputPlaceholder="Search product licenses..."
                />
                <div className="text-muted mb-2">
                    <small>Enter a partial license key to find matches.</small>
                </div>
                {search && (
                    <Container>
                        <ConnectionList
                            as="ul"
                            className="list-group list-group-flush mb-0"
                            aria-label="Subscription licenses"
                        >
                            {connection?.nodes?.length === 0 && (
                                <div className="text-center">No matching license key found.</div>
                            )}

                            {connection?.nodes?.map(node => (
                                <SiteAdminProductLicenseNode
                                    key={node.id}
                                    node={node}
                                    showSubscription={true}
                                    onRevokeCompleted={refetchAll}
                                />
                            ))}
                        </ConnectionList>
                    </Container>
                )}
                {connection && (
                    <SummaryContainer className="mt-2">
                        <ConnectionSummary
                            first={15}
                            centered={true}
                            connection={connection}
                            noun="product license"
                            pluralNoun="product licenses"
                            hasNextPage={hasNextPage}
                            noSummaryIfAllNodesVisible={true}
                        />
                        {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </div>
    )
}
