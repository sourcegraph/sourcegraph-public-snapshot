import { useContext } from 'react'
import { useHistory } from 'react-router-dom'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { asError } from '@sourcegraph/shared/src/util/errors'

import { Settings } from '../../../../../schema/settings.schema'
import { FORM_ERROR, SubmissionErrors } from '../../../../components/form/hooks/useForm'
import { InsightsApiContext } from '../../../../core/backend/api-provider'
import { Insight } from '../../../../core/types'

import { applyEditOperations, getUpdatedSubjectSettings } from './utils'

export interface UseHandleSubmitProps extends PlatformContextProps<'updateSettings'>, SettingsCascadeProps<Settings> {
    originalInsight: Insight | null
}

export interface useHandleSubmitOutput {
    handleEditInsightSubmit: (newInsight: Insight) => Promise<SubmissionErrors>
}

/**
 * Returns submit handler for edit submit form. Updates all related to edit-update
 * settings subjects.
 */
export function useHandleSubmit(props: UseHandleSubmitProps): useHandleSubmitOutput {
    const { originalInsight, platformContext, settingsCascade } = props
    const { getSubjectSettings, updateSubjectSettings } = useContext(InsightsApiContext)
    const history = useHistory()

    const handleEditInsightSubmit = async (newInsight: Insight): Promise<SubmissionErrors> => {
        if (!originalInsight) {
            return
        }

        try {
            const subjectsToUpdate = getUpdatedSubjectSettings({
                oldInsight: originalInsight,
                newInsight,
                settingsCascade,
            })

            const subjectUpdateRequests = Object.keys(subjectsToUpdate).map(subjectId => {
                async function updateSettings(): Promise<void> {
                    const editOperations = subjectsToUpdate[subjectId]

                    // Get jsonc subject settings file.
                    const settings = await getSubjectSettings(subjectId).toPromise()

                    // Modify this jsonc file according to this subject's operations
                    const nextSubjectSettings = applyEditOperations(settings.contents, editOperations)

                    // Call the async update mutation for the new subject's settings file
                    await updateSubjectSettings(platformContext, subjectId, nextSubjectSettings).toPromise()
                }

                return updateSettings()
            })

            await Promise.all(subjectUpdateRequests)

            // Navigate user to the dashboard page with new created dashboard
            history.push(`/insights/dashboards/${newInsight.visibility}`)
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }

    return { handleEditInsightSubmit }
}
