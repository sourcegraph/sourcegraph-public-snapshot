import { useCallback, useContext } from 'react'

import { CodeInsightsBackendContext } from '../../../core/backend/code-insights-backend-context'

import { AsyncValidator } from './utils/use-async-validation'

interface Props {
    initialTitle?: string
    mode: 'creation' | 'edit'
}

export function useAsyncInsightTitleValidator(props: Props): AsyncValidator<string> {
    const { initialTitle, mode } = props
    const { findInsightByName } = useContext(CodeInsightsBackendContext)

    return useCallback<AsyncValidator<string>>(
        async value => {
            const insightTitle = value?.trim() ?? ''

            if (insightTitle === '') {
                return
            }

            // If a user edits existing insight it's ok if they save insight
            // with the same name
            if (mode === 'edit' && insightTitle === initialTitle) {
                return
            }

            const possibleInsight = await findInsightByName({ name: insightTitle }).toPromise()

            if (possibleInsight) {
                return 'An insight with this name already exists. Please set a different name for the new insight.'
            }

            return
        },
        [mode, initialTitle, findInsightByName]
    )
}
