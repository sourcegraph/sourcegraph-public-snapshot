import { camelCase } from 'lodash'
import { useMemo } from 'react'

import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { INSIGHTS_ALL_REPOS_SETTINGS_KEY, InsightTypePrefix } from '../../../core/types'
import { composeValidators, createRequiredValidator } from '../validators'

import { Validator } from './useField'

/**
 * Default value for final user/org settings cascade
 */
const DEFAULT_FINAL_SETTINGS = {}

export interface useTitleValidatorProps {
    insightType: InsightTypePrefix
    settings?: Settings | null
}

export function useInsightTitleDuplicationCheck(props: useTitleValidatorProps): Validator<string> {
    const { settings, insightType } = props

    return useMemo(() => {
        const normalizedSettingsKeys = Object.keys(settings ?? DEFAULT_FINAL_SETTINGS)
        const normalizedInsightAllReposKeys = Object.keys(
            settings?.[INSIGHTS_ALL_REPOS_SETTINGS_KEY] ?? DEFAULT_FINAL_SETTINGS
        )

        const existingInsightNames = new Set(
            [...normalizedSettingsKeys, ...normalizedInsightAllReposKeys]
                // According to our convention about insights name <insight type>.insight.<insight name>
                .filter(key => key.startsWith(`${insightType}`))
                .map(key => camelCase(key.split(`${insightType}.`).pop()))
        )

        return value =>
            existingInsightNames.has(camelCase(value))
                ? 'An insight with this name already exists. Please set a different name for the new insight.'
                : undefined
    }, [settings, insightType])
}

/**
 * Shared validator for title insight.
 * We can't have two or more insights with the same name, since we rely on the insight name
 * as an id (which is camel cased title) in settings cascade.
 */
export function useInsightTitleValidator(props: useTitleValidatorProps): Validator<string> {
    const hasInsightTitleDuplication = useInsightTitleDuplicationCheck(props)

    return useMemo(
        () =>
            composeValidators<string>(
                createRequiredValidator('Title is a required field.'),
                hasInsightTitleDuplication
            ),
        [hasInsightTitleDuplication]
    )
}
