import React from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Settings } from '../../../schema/settings.schema'
import { Insight } from '../../core/types'

import { SmartInsight } from './components/smart-insight/SmartInsight'
import { ViewGrid } from './components/view-grid/ViewGrid'

interface SmartInsightsViewGridProps
    extends TelemetryProps,
        SettingsCascadeProps<Settings>,
        PlatformContextProps<'updateSettings'>,
        ExtensionsControllerProps {
    insights: Insight[]
}

/**
 * Renders grid of smart (stateful) insight card. These cards can independently extract and update
 * the insights settings (settings cascade subjects).
 */
export const SmartInsightsViewGrid: React.FunctionComponent<SmartInsightsViewGridProps> = props => {
    const { telemetryService, insights, platformContext, settingsCascade, extensionsController } = props

    return (
        <ViewGrid viewIds={insights.map(insight => insight.id)} telemetryService={telemetryService}>
            {insights.map(insight => (
                <SmartInsight
                    key={insight.id}
                    insight={insight}
                    telemetryService={telemetryService}
                    platformContext={platformContext}
                    settingsCascade={settingsCascade}
                    extensionsController={extensionsController}
                />
            ))}
        </ViewGrid>
    )
}
