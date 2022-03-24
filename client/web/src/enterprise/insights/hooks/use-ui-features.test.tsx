import React from 'react'

import { renderHook } from '@testing-library/react-hooks'

import { CodeInsightsBackend } from '../core/backend/code-insights-backend'
import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'
import { CodeInsightsGqlBackend } from '../core/backend/gql-backend/code-insights-gql-backend'

import { useUiFeatures } from './use-ui-features'

interface UiFeatureWrapperProps {
    mockApi: Partial<CodeInsightsBackend>
}

const UiFeatureWrapper: React.FunctionComponent<UiFeatureWrapperProps> = ({ mockApi, children }) => {
    const api: CodeInsightsBackend = {
        ...new CodeInsightsGqlBackend({} as any),
        ...mockApi,
    }
    return <CodeInsightsBackendContext.Provider value={api}>{children}</CodeInsightsBackendContext.Provider>
}

describe('useUiFeatures', () => {
    test.each([true, false])('should return licensed: %s', licensed => {
        const mockApi: Partial<CodeInsightsBackend> = {
            UIFeatures: { licensed, insightsLimit: 2 },
        }
        const wrapper: React.FunctionComponent = ({ children }) => (
            <UiFeatureWrapper mockApi={mockApi}>{children}</UiFeatureWrapper>
        )
        const { result } = renderHook(() => useUiFeatures(), { wrapper })
        expect(result.current.licensed).toBe(licensed)
    })
})
