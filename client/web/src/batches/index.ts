import { Settings, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

/**
 * Utility for checking if a user has the experimental feature, batch change server-side
 * execution, enabled in their settings
 */
export const isBatchChangesExecutionEnabled = (settingsCascade: SettingsCascadeOrError<Settings>): boolean =>
    Boolean(
        settingsCascade !== null &&
            !isErrorLike(settingsCascade.final) &&
            settingsCascade.final?.experimentalFeatures?.batchChangesExecution
    )

/**
 * Common props for components needing to decide whether to show content related to Batch
 * Changes
 */
export interface BatchChangesProps {
    batchChangesExecutionEnabled: boolean
    batchChangesEnabled: boolean
}
