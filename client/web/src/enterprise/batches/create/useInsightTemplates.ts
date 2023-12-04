import { useLocation } from 'react-router-dom'

import type { Settings, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

import { isGoCodeCheckerTemplatesEnabled } from '../../../batches'

import { getTemplateRenderer } from './go-checker-templates'

interface UseInsightTemplatesResult {
    renderTemplate?: (title: string) => string
    insightTitle?: string
}

/**
 * Custom hook for create page which checks if a user has enabled the experimental code
 * insights integration with batch changes and is creating a batch change from a Go code
 * checker insight.
 *
 * @param settingsCascade The user's current settings.
 */
export const useInsightTemplates = (settingsCascade: SettingsCascadeOrError<Settings>): UseInsightTemplatesResult => {
    const location = useLocation()
    const parameters = new URLSearchParams(location.search)
    const renderTemplate = getTemplateRenderer(parameters.get('kind'))
    const insightTitle = parameters.get('title') ?? undefined

    return isGoCodeCheckerTemplatesEnabled(settingsCascade)
        ? { renderTemplate, insightTitle }
        : { renderTemplate: undefined, insightTitle: undefined }
}
