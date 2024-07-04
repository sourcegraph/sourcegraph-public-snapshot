import { createConnectTransport } from '@connectrpc/connect-web'
import { QueryClient, QueryClientProvider as TanstackQueryClientProvider } from '@tanstack/react-query'

export const enterprisePortalLocal = createConnectTransport({
    baseUrl: `${window.context.externalURL}/.api/enterpriseportal/local`,
})

export const enterprisePortalDev = createConnectTransport({
    baseUrl: `${window.context.externalURL}/.api/enterpriseportal/dev`,
    useHttpGet: true,
})

export const enterprisePortalProd = createConnectTransport({
    baseUrl: `${window.context.externalURL}/.api/enterpriseportal/prod`,
    useHttpGet: true,
})
