import React, { useMemo } from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Settings } from '../../../schema/settings.schema'
import { Insight } from '../../core/types'
import { useDistinctValue } from '../../hooks/use-distinct-value'

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

    const insightIds = useDistinctValue(insights.map(insight => insight.id))
    const gridItems = useMemo(
        () =>
            insights.map(insight => (
                <SmartInsight
                    key={insight.id}
                    insight={insight}
                    telemetryService={telemetryService}
                    platformContext={platformContext}
                    settingsCascade={settingsCascade}
                    extensionsController={extensionsController}
                />
            )),
        // Ignore settings cascade change in order to avoid grid item re-render and
        // grid position items animations. In some cases (like insight filters updating
        // we want to ignore changes of insights from settings cascade).
        // But still trigger grid animation rerender if insight ordering or insight count
        // have been changed.
        // eslint-disable-next-line react-hooks/exhaustive-deps
        [telemetryService, platformContext, extensionsController, insightIds]
    )

    return (
        <ViewGrid viewIds={insights.map(insight => insight.id)} telemetryService={telemetryService}>
            {gridItems}
        </ViewGrid>
    )
}
