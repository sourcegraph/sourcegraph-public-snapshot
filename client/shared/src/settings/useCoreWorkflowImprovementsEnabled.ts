import { useTemporarySetting, UseTemporarySettingsReturnType } from './temporary/useTemporarySetting'

export function useCoreWorkflowImprovementsEnabled(): UseTemporarySettingsReturnType<'coreWorkflowImprovements.enabled'> {
    return useTemporarySetting('coreWorkflowImprovements.enabled')
}
