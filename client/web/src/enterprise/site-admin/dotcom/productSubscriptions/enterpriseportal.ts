import type { ConnectError, Transport } from '@connectrpc/connect'
import { createQueryOptions, defaultOptions } from '@connectrpc/connect-query'
import { createConnectTransport } from '@connectrpc/connect-web'
import { QueryClient, type UseQueryResult, useQuery } from '@tanstack/react-query'

import { getCodyGatewayUsage } from './enterpriseportalgen/codyaccess-CodyAccessService_connectquery'
import type { GetCodyGatewayUsageResponse } from './enterpriseportalgen/codyaccess_pb'

/**
 * Use a shared QueryClient defined here and explicitly provided to react-query
 * for now to avoid bleading the QueryClientProvider to the site admin parent.
 *
 * Another problem is that @robert was unable to get QueryClientProvider working
 * even when placing it at various points the the tree.
 *
 * Overall this is only meant to be an interim integration, until Enterprise
 * Portal gets its own dedicated UI:
 * https://linear.app/sourcegraph/project/kr-p-enterprise-portal-user-interface-dadd5ff28bd8
 */
const queryClient = new QueryClient({ defaultOptions })

const enterprisePortalLocal = createConnectTransport({
    baseUrl: '/.api/enterpriseportal/local',
})

const enterprisePortalDev = createConnectTransport({
    baseUrl: '/.api/enterpriseportal/dev',
})

const enterprisePortalProd = createConnectTransport({
    baseUrl: '/.api/enterpriseportal/prod',
})

/**
 * Environment describes the Enteprise Portal instance to target.
 *
 * 'local' is only valid in local dev.
 */
export type Environment = 'prod' | 'dev' | 'local'

const environments = new Map<Environment, Transport>([
    ['prod', enterprisePortalProd],
    ['dev', enterprisePortalDev],
    ['local', enterprisePortalLocal],
])

function mustGetEnvironment(env: Environment): Transport {
    if (env === 'local' && window.context.deployType !== 'dev') {
        throw new Error(`Environment ${env} not allowed outside local dev`)
    }
    const transport = environments.get(env)
    if (transport) {
        return transport
    }
    throw new Error(`Environment ${env} not configured`)
}

/**
 * Retrieves Cody Gateway usage for a given subscription.
 * @param env Enterprise Portal environment to target.
 * @param subscriptionUUID Subscription UUID
 */
export function useGetCodyGatewayUsage(
    env: Environment,
    subscriptionUUID: string
): UseQueryResult<GetCodyGatewayUsageResponse, ConnectError> {
    return useQuery(
        createQueryOptions(
            getCodyGatewayUsage,
            {
                query: { value: subscriptionUUID, case: 'subscriptionId' },
            },
            { transport: mustGetEnvironment(env) }
        ),
        queryClient
    )
}
