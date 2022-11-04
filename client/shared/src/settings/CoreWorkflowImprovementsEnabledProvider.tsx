import React, { createContext } from 'react'

import { noop } from 'lodash'

import { useTemporarySetting, UseTemporarySettingsReturnType } from './temporary/useTemporarySetting'

// The `coreWorkflowImprovements.enabled` temporary setting guards a lot of UI changes. To prevent flashes of the old UI
// we have to share the evaluated temporary setting value between components through the context.
// Otherwise, each component would evaluate `useTemporarySetting` on their own leading to UI jitter while the
// temporary setting was "loading".
export const CoreWorkflowImprovementsEnabledContext = createContext<
    UseTemporarySettingsReturnType<'coreWorkflowImprovements.enabled_deprecated'>
>([true, noop, 'initial'])
CoreWorkflowImprovementsEnabledContext.displayName = 'CoreWorkflowImprovementsContext'

export const CoreWorkflowImprovementsEnabledProvider: React.FunctionComponent<React.PropsWithChildren<{}>> = ({
    children,
}) => {
    const coreWorkflowImprovementsEnabled = useTemporarySetting('coreWorkflowImprovements.enabled_deprecated', true)

    return (
        <CoreWorkflowImprovementsEnabledContext.Provider value={coreWorkflowImprovementsEnabled}>
            {children}
        </CoreWorkflowImprovementsEnabledContext.Provider>
    )
}
