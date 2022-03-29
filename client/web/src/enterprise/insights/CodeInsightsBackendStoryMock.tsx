import React from 'react'

import { CodeInsightsBackendContext } from './core/backend/code-insights-backend-context'
import { CodeInsightsGqlBackend } from './core/backend/gql-backend/code-insights-gql-backend'

export const CodeInsightsBackendStoryMock: React.FunctionComponent<{
    mocks: Partial<CodeInsightsGqlBackend>
}> = ({ children, mocks }) => {
    // Pass in an empty object because mocking `watchQuery` is too difficult.
    const backend = new CodeInsightsGqlBackend({} as any)

    // Override the existing backend with whatever methods we need for this story
    const backendMock = {
        ...backend,
        ...mocks,
    }

    return <CodeInsightsBackendContext.Provider value={backendMock}>{children}</CodeInsightsBackendContext.Provider>
}
