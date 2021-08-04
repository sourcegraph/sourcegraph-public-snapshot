import React from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Settings } from '../../../../../schema/settings.schema'
import { Insight, InsightType, isSearchBasedInsight } from '../../../../core/types'
import { BackendInsight } from '../backend-insight/BackendInsight'
import { ExtensionInsight } from '../extension-insight/ExtensionInsight'

export interface InsightProps
    extends TelemetryProps,
        SettingsCascadeProps<Settings>,
        PlatformContextProps<'updateSettings'>,
        ExtensionsControllerProps,
        React.HTMLAttributes<HTMLElement> {
    insight: Insight
}

/**
 * Render smart insight with (gql or extension api) fetcher and independent mutation
 * actions.
 */
export const SmartInsight: React.FunctionComponent<InsightProps> = props => {
    const { insight, telemetryService, settingsCascade, platformContext, extensionsController, ...otherProps } = props

    if (isSearchBasedInsight(insight)) {
        return insight.type === InsightType.Backend ? (
            <BackendInsight
                insight={insight}
                telemetryService={telemetryService}
                settingsCascade={settingsCascade}
                platformContext={platformContext}
                {...otherProps}
            />
        ) : (
            <ExtensionInsight
                viewId={insight.id}
                telemetryService={telemetryService}
                settingsCascade={settingsCascade}
                platformContext={platformContext}
                extensionsController={extensionsController}
                {...otherProps}
            />
        )
    }

    // Code-stats insight is always extension-based
    return (
        <ExtensionInsight
            viewId={insight.id}
            telemetryService={telemetryService}
            settingsCascade={settingsCascade}
            platformContext={platformContext}
            extensionsController={extensionsController}
            {...otherProps}
        />
    )
}
