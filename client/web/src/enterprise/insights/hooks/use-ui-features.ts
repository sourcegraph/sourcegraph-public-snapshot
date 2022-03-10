import { useContext, useMemo } from 'react'

import { useObservable } from '@sourcegraph/wildcard'

import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'

export interface UseUiFeatures {
    licensed: boolean
}

export function useUiFeatures(): UseUiFeatures {
    const { getUiFeatures } = useContext(CodeInsightsBackendContext)

    const uiFeatures = useObservable(useMemo(() => getUiFeatures(), [getUiFeatures]))

    return {
        licensed: !!uiFeatures?.licensed,
    }
}
