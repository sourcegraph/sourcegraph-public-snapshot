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

/**
 * QueryClientProvider wraps its children with the react-query ClientProvider.
 * It is used to access the Cody Pro API and is only utilized on dotcom.
 * In enterprise mode, it simply passes through the children.
 */
export const QueryClientProvider: React.FC<React.PropsWithChildren<{ isSourcegraphDotCom: boolean }>> = ({
    isSourcegraphDotCom,
    children,
}) => {
    if (!isSourcegraphDotCom) {
        return <>{children}</>
    }

    return <ReactQueryClientProvider client={queryClient}>{children}</ReactQueryClientProvider>
}
