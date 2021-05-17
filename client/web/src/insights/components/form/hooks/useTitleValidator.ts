import { camelCase } from 'lodash'
import { useMemo } from 'react'

import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { InsightTypePrefix } from '../../../core/types'
import { composeValidators, createRequiredValidator } from '../validators'

import { Validator } from './useField'

/** Default value for final user/org settings cascade */
const DEFAULT_FINAL_SETTINGS = {}

export interface useTitleValidatorProps {
    insightType: InsightTypePrefix
    settings?: Settings | null
}

/**
 * Shared validator for title insight.
 * We can't have two or more insights with the same name, since we rely on name as on id at insights pages.
 * */
export function useTitleValidator(props: useTitleValidatorProps): Validator<string> {
    const { settings, insightType } = props

    return useMemo(() => {
        const alreadyExistsInsightNames = new Set(
            Object.keys(settings ?? DEFAULT_FINAL_SETTINGS)
                // According to our convention about insights name <insight type>.insight.<insight name>
                .filter(key => key.startsWith(`${insightType}`))
                .map(key => camelCase(key.split('.').pop()))
        )

        return composeValidators<string>(createRequiredValidator('Title is a required field.'), value =>
            alreadyExistsInsightNames.has(camelCase(value))
                ? 'An insight with this name already exists. Please set a different name for the new insight.'
                : undefined
        )
    }, [settings, insightType])
}
