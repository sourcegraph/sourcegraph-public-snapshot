import { useContext } from 'react'

import { CoreWorkflowImprovementsEnabledContext } from './CoreWorkflowImprovementsEnabledProvider'
import { UseTemporarySettingsReturnType } from './temporary/useTemporarySetting'

export function useCoreWorkflowImprovementsEnabled(): UseTemporarySettingsReturnType<'coreWorkflowImprovements.enabled_deprecated'> {
    return useContext(CoreWorkflowImprovementsEnabledContext)
}
