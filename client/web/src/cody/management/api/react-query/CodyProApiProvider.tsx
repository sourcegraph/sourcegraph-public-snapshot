import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

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
 * CodyProApiProvider wraps its children with the react-query QueryClientProvider.
 * It is used to access the Cody Pro API and is only utilized on dotcom.
 * In enterprise mode, it simply passes through the children.
 */
export const CodyProApiProvider: React.FC<React.PropsWithChildren<{ isSourcegraphDotCom: boolean }>> = ({
    isSourcegraphDotCom,
    children,
}) => {
    if (!isSourcegraphDotCom) {
        return <>{children}</>
    }

    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
}
