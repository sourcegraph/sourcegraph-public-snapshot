import { renderHook } from '@testing-library/react-hooks'
import React from 'react'

import { CodeInsightsBackend } from '../core/backend/code-insights-backend'
import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'
import { CodeInsightsGqlBackend } from '../core/backend/gql-api/code-insights-gql-backend'

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
            getUiFeatures: () => ({ licensed }),
        }
        const wrapper: React.FunctionComponent = ({ children }) => (
            <UiFeatureWrapper mockApi={mockApi}>{children}</UiFeatureWrapper>
        )
        const { result } = renderHook(() => useUiFeatures(), { wrapper })
        expect(result.current.licensed).toBe(licensed)
    })
})
