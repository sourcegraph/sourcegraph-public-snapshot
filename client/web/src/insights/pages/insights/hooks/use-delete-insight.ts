import { useCallback, useContext } from 'react'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { InsightsApiContext } from '../../../core/backend/api-provider'
import { modify } from '../../../core/jsonc-operation'
import { InsightTypePrefix } from '../../../core/types'

export interface UseDeleteInsightProps extends SettingsCascadeProps, PlatformContextProps<'updateSettings'> {}

export interface UseDeleteInsightAPI {
    handleDelete: (insightID: string) => Promise<void>
}

export function useDeleteInsight(props: UseDeleteInsightProps): UseDeleteInsightAPI {
    const { settingsCascade, platformContext } = props

    const { getSubjectSettings, updateSubjectSettings } = useContext(InsightsApiContext)

    const handleDelete = useCallback(
        async (id: string) => {
            // According to our naming convention of insight
            // <type>.<name>.<render view = insight page | directory | home page>
            const insightID = id.split('.').slice(0, -1).join('.')
            const subjects = settingsCascade.subjects

            // For backward compatibility with old code stats insight api we have to delete
            // this insight in a special way. See link below for more information.
            // https://github.com/sourcegraph/sourcegraph-code-stats-insights/blob/master/src/code-stats-insights.ts#L33
            const isOldCodeStatsInsight = insightID === `${InsightTypePrefix.langStats}.language`

            const keyForSearchInSettings = isOldCodeStatsInsight
                ? // Hardcoded value of id from old version of stats insight extension API
                  'codeStatsInsights.query'
                : insightID

            const subjectID = subjects?.find(
                ({ settings }) => settings && !isErrorLike(settings) && !!settings[keyForSearchInSettings]
            )?.subject?.id

            if (!subjectID) {
                console.error("Couldn't find the subject when trying to delete insight. Parameters", {
                    insightID,
                    subjects,
                })
                return
            }

            try {
                // Fetch the settings of particular subject which the insight belongs to
                const settings = await getSubjectSettings(subjectID).toPromise()

                const editedSettings = getEditedSettings({
                    originSettings: settings.contents,
                    insightID,
                    isOldCodeStatsInsight,
                })

                // Update local settings of application with new settings without insight
                await updateSubjectSettings(platformContext, subjectID, editedSettings).toPromise()
            } catch (error) {
                // TODO [VK] Improve error UI for deleting
                console.error(error)
            }
        },
        [platformContext, settingsCascade, getSubjectSettings, updateSubjectSettings]
    )

    return { handleDelete }
}

interface GetEditedSettingsProps {
    originSettings: string
    isOldCodeStatsInsight: boolean
    insightID: string
}

/**
 * Return edited settings without deleted insight settings section
 */
function getEditedSettings(props: GetEditedSettingsProps): string {
    const { originSettings, isOldCodeStatsInsight, insightID } = props

    if (isOldCodeStatsInsight) {
        const editedSettings = modify(
            originSettings,
            // According to our naming convention <type>.insight.<name>
            ['codeStatsInsights.query'],
            undefined
        )

        return modify(
            editedSettings,
            // According to our naming convention <type>.insight.<name>
            ['codeStatsInsights.otherThreshold'],
            undefined
        )
    }

    // Remove insight settings from subject (user/org settings)
    return modify(
        originSettings,
        // According to our naming convention <type>.insight.<name>
        [insightID],
        undefined
    )
}
