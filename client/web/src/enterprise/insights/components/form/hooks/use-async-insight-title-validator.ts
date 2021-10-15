import { useCallback, useContext } from 'react'

import { CodeInsightsBackendContext } from '../../../core/backend/code-insights-backend-context'
import { InsightTypePrefix } from '../../../core/types'

import { AsyncValidator } from './utils/use-async-validation'

interface Props {
    initialTitle?: string
    type: InsightTypePrefix
}

export function useAsyncInsightTitleValidator(props: Props) {
    const { initialTitle, type } = props
    const { findInsightByName } = useContext(CodeInsightsBackendContext)

    return useCallback<AsyncValidator<string>>(
        async title => {
            if (!title || title.trim() === '' || title === initialTitle) {
                return
            }

            const possibleInsight = await findInsightByName({ name: title, type }).toPromise()

            if (possibleInsight) {
                return 'An insight with this name already exists. Please set a different name for the new insight.'
            }

            return
        },
        [findInsightByName, initialTitle, type]
    )
}
