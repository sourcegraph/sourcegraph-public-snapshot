import { isErrorLike } from '@sourcegraph/common'
import type { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import type { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

/**
 * Utility for checking if a user has the experimental feature, batch change server-side
 * execution, enabled in their settings.
 */
export const isBatchChangesExecutionEnabled = (settingsCascade: SettingsCascadeOrError<Settings>): boolean =>
    Boolean(
        settingsCascade.final !== null &&
            !isErrorLike(settingsCascade.final) &&
            settingsCascade.final.experimentalFeatures?.batchChangesExecution !== false
    )

/**
 * Utility for checking if a user has the experimental feature flag for the code insights
 * integration with batch changes enabled in their settings.
 */
export const isGoCodeCheckerTemplatesEnabled = (settingsCascade: SettingsCascadeOrError<Settings>): boolean =>
    Boolean(
        settingsCascade.final !== null &&
            !isErrorLike(settingsCascade.final) &&
            settingsCascade.final.experimentalFeatures?.goCodeCheckerTemplates
    )

/**
 * Common props for components needing to decide whether to show content related to Batch
 * Changes.
 */
export interface BatchChangesProps {
    batchChangesExecutionEnabled: boolean
    batchChangesEnabled: boolean
    batchChangesWebhookLogsEnabled: boolean
}
