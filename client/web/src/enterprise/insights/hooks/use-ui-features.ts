import { useContext, useMemo } from 'react'

import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'

export interface UseUiFeatures {
    licensed: boolean
}

export function useUiFeatures(): UseUiFeatures {
    const { getUiFeatures } = useContext(CodeInsightsBackendContext)

    const uiFeatures = useMemo(() => getUiFeatures(), [getUiFeatures])

    return {
        licensed: !!uiFeatures?.licensed,
    }
}
