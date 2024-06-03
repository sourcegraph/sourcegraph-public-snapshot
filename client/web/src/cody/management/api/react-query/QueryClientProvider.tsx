import { QueryClient, QueryClientProvider as ReactQueryClientProvider } from '@tanstack/react-query'

// Tweak the default queries and mutations behavior.
// See defaults here: https://tanstack.com/query/latest/docs/framework/react/guides/important-defaults
const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            // If query failed, it's not likely that refetching it will succeed, so don't retry.
            retry: false,
        },
        mutations: {
            // If query failed, it's not likely that refetching it will succeed, so don't retry.
            retry: false,
        },
    },
})

export const QueryClientProvider: React.FC<React.PropsWithChildren<{}>> = ({ children }) => (
    <ReactQueryClientProvider client={queryClient}>{children}</ReactQueryClientProvider>
)
