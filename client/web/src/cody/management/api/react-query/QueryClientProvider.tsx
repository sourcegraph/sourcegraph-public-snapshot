import type { ReactNode } from 'react'

import { QueryClient, QueryClientProvider as ReactQueryClientProvider } from '@tanstack/react-query'

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            // If query failed, it's not likely that refetching it will succeed, so don't retry.
            retry: false,

            // Once loaded, we consider query result to be stable. No need to refetch it again.
            // See: https://tanstack.com/query/latest/docs/framework/react/guides/important-defaults
            staleTime: Infinity,
        },
    },
})

export const QueryClientProvider: React.FC<{ children?: ReactNode | undefined }> = ({ children }) => (
    <ReactQueryClientProvider client={queryClient}>{children}</ReactQueryClientProvider>
)
