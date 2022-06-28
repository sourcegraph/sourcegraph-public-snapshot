import React from 'react'

import { renderHook } from '@testing-library/react-hooks'

import { CodeInsightsBackend, CodeInsightsBackendContext, CodeInsightsGqlBackend } from '../core'

import { useUiFeatures } from './use-ui-features'

interface UiFeatureWrapperProps {
    mockApi: Partial<CodeInsightsBackend>
}

const UiFeatureWrapper: React.FunctionComponent<React.PropsWithChildren<UiFeatureWrapperProps>> = ({
    mockApi,
    children,
}) => {
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
        const wrapper: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({ children }) => (
            <UiFeatureWrapper mockApi={mockApi}>{children}</UiFeatureWrapper>
        )
        const { result } = renderHook(() => useUiFeatures(), { wrapper })
        expect(result.current.licensed).toBe(licensed)
    })
})
