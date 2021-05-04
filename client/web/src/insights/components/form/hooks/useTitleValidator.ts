import { camelCase } from 'lodash'
import { useMemo } from 'react'

import { composeValidators, createRequiredValidator } from '../validators'

import { Validator } from './useField'

export interface useTitleValidatorProps {
    insightType: 'searchInsights' | 'codeStatsInsights'
    settings: { [key: string]: unknown }
}

/**
 * Shared validator for title insight.
 * We can't have two or more insights with the same name, since we rely on name as on id at insights pages.
 * */
export function useTitleValidator(props: useTitleValidatorProps): Validator<string> {
    const { settings, insightType } = props

    return useMemo(() => {
        const alreadyExistsInsightNames = new Set(
            Object.keys(settings)
                // According to our convention about insights name <insight type>.insight.<insight name>
                .filter(key => key.startsWith(`${insightType}.insight`))
                .map(key => camelCase(key.split('.').pop()))
        )

        return composeValidators<string>(createRequiredValidator('Title is required field for code insight.'), value =>
            alreadyExistsInsightNames.has(camelCase(value))
                ? 'An insight with this name already exists. Please set a different name for the new insight.'
                : undefined
        )
    }, [settings, insightType])
}
