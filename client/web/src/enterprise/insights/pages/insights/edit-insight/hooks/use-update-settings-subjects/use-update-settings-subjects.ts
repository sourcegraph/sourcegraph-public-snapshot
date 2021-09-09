import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'

import { Settings } from '../../../../../../../schema/settings.schema'
import { Insight } from '../../../../../core/types'
import { usePersistEditOperations } from '../../../../../hooks/use-persist-edit-operations'

import { getUpdatedSubjectSettings } from './get-updated-subject-settings'

interface UseUpdateSettingsSubjectProps extends PlatformContextProps<'updateSettings'> {}

interface UpdateSettingSubjectsInputs extends SettingsCascadeProps<Settings> {
    oldInsight: Insight
    newInsight: Insight
}

interface UseUpdateSettingsSubjectResult {
    updateSettingSubjects: (inputs: UpdateSettingSubjectsInputs) => Promise<void>
}

/**
 * Updates all settings cascade subjects according to updated and original settings
 * (config itself, visibility levels therefore dashboard settings as well)
 *
 * Extracted for testing purposes
 */
export function useUpdateSettingsSubject(props: UseUpdateSettingsSubjectProps): UseUpdateSettingsSubjectResult {
    const { platformContext } = props
    const { persist } = usePersistEditOperations({ platformContext })

    const updateSettingSubjects = (inputs: UpdateSettingSubjectsInputs): Promise<void> => {
        const editOperations = getUpdatedSubjectSettings(inputs)

        return persist(editOperations)
    }

    return { updateSettingSubjects }
}
