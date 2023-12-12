import { type UseTemporarySettingsReturnType, useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'

export function useOpenCodeGraphVisibility(): UseTemporarySettingsReturnType<'openCodeGraph.annotations.visible'> {
    return useTemporarySetting('openCodeGraph.annotations.visible')
}
