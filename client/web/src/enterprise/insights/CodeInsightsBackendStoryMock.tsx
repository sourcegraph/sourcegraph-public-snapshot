import React from 'react'

import { CodeInsightsBackendContext, CodeInsightsGqlBackend } from './core'

export const CodeInsightsBackendStoryMock: React.FunctionComponent<
    React.PropsWithChildren<{
        mocks: Partial<CodeInsightsGqlBackend>
    }>
> = ({ children, mocks }) => {
    // Pass in an empty object because mocking `watchQuery` is too difficult.
    const backend = new CodeInsightsGqlBackend({} as any)

    // Override the existing backend with whatever methods we need for this story
    const backendMock = {
        ...backend,
        ...mocks,
    }

    return <CodeInsightsBackendContext.Provider value={backendMock}>{children}</CodeInsightsBackendContext.Provider>
}
