import { GraphQLResult } from '@sourcegraph/shared/src/graphql/graphql'

/**
 * Sourcegraph VS Code methods exposed to Webviews
 *
 * Note: this API object lives in the VS Code extension host runtime.
 */
export interface SourcegraphVSCodeExtensionAPI {
    ping: () => 'pong!'

    requestGraphQL: (request: string, variables: any) => Promise<GraphQLResult<any>>
}

/**
 * Webview methods exposed to the Sourcegraph VS Code extension.
 */
export interface SourcegraphVSCodeWebviewAPI {
    setRoute: (route: 'search') => void
}
