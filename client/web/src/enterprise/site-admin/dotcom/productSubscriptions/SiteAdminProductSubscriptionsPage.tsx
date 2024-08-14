import React, { useEffect, useState } from 'react'

import type { PartialMessage } from '@bufbuild/protobuf'
import { mdiPlus } from '@mdi/js'
import { QueryClientProvider } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { useDebounce } from 'use-debounce'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, Container, Icon, Link, PageHeader } from '@sourcegraph/wildcard'

import {
    ConnectionContainer,
    ConnectionError,
    ConnectionForm,
    ConnectionList,
    ConnectionLoading,
    SummaryContainer,
} from '../../../../components/FilteredConnection/ui'
import { PageTitle } from '../../../../components/PageTitle'

import { queryClient, useListEnterpriseSubscriptions, type EnterprisePortalEnvironment } from './enterpriseportal'
import { EnterprisePortalEnvSelector, getDefaultEnterprisePortalEnv } from './EnterprisePortalEnvSelector'
import { EnterprisePortalEnvWarning } from './EnterprisePortalEnvWarning'
import type { ListEnterpriseSubscriptionsFilter } from './enterpriseportalgen/subscriptions_pb'
import {
    SiteAdminProductSubscriptionNode,
    SiteAdminProductSubscriptionNodeHeader,
} from './SiteAdminProductSubscriptionNode'

interface Props extends TelemetryV2Props {}

/**
 * Displays the enterprise subscriptions (formerly known as "product subscriptions") that have been
 * created on Sourcegraph.com.
 */
export const SiteAdminProductSubscriptionsPage: React.FunctionComponent<React.PropsWithChildren<Props>> = props => (
    <QueryClientProvider client={queryClient}>
        <Page {...props} />
    </QueryClientProvider>
)

const QUERY_PARAM_KEY = 'query'
const QUERY_PARAM_ENV = 'env'
const QUERY_PARAM_FILTER = 'filter'

type FilterType = 'display_name' | 'sf_sub_id'

const MAX_RESULTS = 50

const Page: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ telemetryRecorder }) => {
    const [searchParams, setSearchParams] = useSearchParams()
    const [env, setEnv] = useState<EnterprisePortalEnvironment>(
        (searchParams.get(QUERY_PARAM_ENV) as EnterprisePortalEnvironment) || getDefaultEnterprisePortalEnv()
    )
    const [query, setQuery] = useState<string>(searchParams.get(QUERY_PARAM_KEY) ?? '')
    const [filters, setFilters] = useState<{ filter: FilterType }>({
        filter: (searchParams.get(QUERY_PARAM_FILTER) as FilterType) ?? 'display_name',
    })

    useEffect(() => {
        const currentEnv = searchParams.get(QUERY_PARAM_ENV) as EnterprisePortalEnvironment

        searchParams.set(QUERY_PARAM_KEY, query?.trim() ?? '')
        searchParams.set(QUERY_PARAM_FILTER, filters.filter)
        searchParams.set(QUERY_PARAM_ENV, env)
        setSearchParams(searchParams)

        // HACK: env state doesn't propagate to hooks correctly, so conditionally
        // reload the page.
        // Required until we fix https://linear.app/sourcegraph/issue/CORE-245
        if (env !== currentEnv) {
            window.location.reload()
            return
        }

        telemetryRecorder.recordEvent('admin.productSubscriptions', 'view', {
            version: 2,
            privateMetadata: { env, filters },
        })
    }, [telemetryRecorder, query, searchParams, setSearchParams, filters, env])

    const [debouncedQuery] = useDebounce(query, 200)

    let listFilters: PartialMessage<ListEnterpriseSubscriptionsFilter>[] = []
    // no filter without a query
    if (debouncedQuery) {
        switch (filters.filter) {
            case 'display_name': {
                listFilters = [
                    {
                        filter: {
                            case: 'displayName',
                            value: debouncedQuery,
                        },
                    },
                ]
                break
            }
            case 'sf_sub_id': {
                listFilters = [
                    {
                        filter: {
                            case: 'salesforce',
                            value: { subscriptionId: debouncedQuery },
                        },
                    },
                ]
                break
            }
        }
    }
    const { error, isFetching, data } = useListEnterpriseSubscriptions(env, listFilters, {
        limit: MAX_RESULTS,
    })

    return (
        <div className="site-admin-product-subscriptions-page">
            <PageTitle title="Enterprise subscriptions" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Enterprise subscriptions' }]}
                description="Manage subscriptions for Sourcegraph Enterprise instances."
                actions={
                    <div className="align-items-end d-flex">
                        <EnterprisePortalEnvSelector env={env} setEnv={setEnv} />
                        <div>
                            <Button to={`./new?env=${env}`} variant="primary" as={Link} display="block">
                                <Icon aria-hidden={true} svgPath={mdiPlus} />
                                Create subscription
                            </Button>
                        </div>
                    </div>
                }
                className="mb-3"
            />

            <EnterprisePortalEnvWarning env={env} actionText="managing subscriptions" />

            <ConnectionContainer>
                <Container className="mb-3">
                    <ConnectionForm
                        inputValue={query}
                        filterValues={filters}
                        inputClassName="ml-2"
                        filters={[
                            {
                                id: 'filter',
                                type: 'select',
                                label: 'Filter',
                                options: [
                                    {
                                        args: {},
                                        label: 'Display name',
                                        value: 'display_name',
                                        tooltip: 'Partial, case-insensitive match on the subscription display name',
                                    },
                                    {
                                        args: {},
                                        label: 'Salesforce subscription ID',
                                        value: 'sf_sub_id',
                                        tooltip:
                                            'Exact match on the Salesforce subscription ID associated with the subscription',
                                    },
                                ],
                            },
                        ]}
                        onInputChange={event => {
                            setQuery(event.target.value)
                        }}
                        onFilterSelect={(filter, value) => {
                            if (value) {
                                setFilters({ ...filters, [filter.id]: value as FilterType })
                            }
                        }}
                        inputPlaceholder="Enter a query to find subscriptions"
                    />
                </Container>

                <Container className="mb-3">
                    {error && <ConnectionError errors={[error.message]} />}

                    {isFetching && <ConnectionLoading />}
                    {data && (
                        <ConnectionList as="table" aria-label="Enterprise subscriptions">
                            <SiteAdminProductSubscriptionNodeHeader />
                            <tbody>
                                {data?.subscriptions?.map(node => (
                                    <SiteAdminProductSubscriptionNode key={node.id} node={node} env={env} />
                                ))}
                            </tbody>
                        </ConnectionList>
                    )}
                    <SummaryContainer className="mt-4" centered={true}>
                        {data && data.subscriptions.length > 0 && (
                            <span className="text-muted">Showing {data.subscriptions.length} subscriptions.</span>
                        )}
                        {data && data.subscriptions.length === 0 && (
                            <span className="text-muted">No subscriptions found.</span>
                        )}
                        {data && data.subscriptions.length >= MAX_RESULTS && (
                            <ConnectionError
                                className="mt-2"
                                errors={[
                                    `Only ${MAX_RESULTS} results are shown at a time - narrow your search for more accurate results.`,
                                ]}
                            />
                        )}
                    </SummaryContainer>
                </Container>
            </ConnectionContainer>
        </div>
    )
}
