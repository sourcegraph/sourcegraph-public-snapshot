import React from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Settings } from '../../../../../schema/settings.schema'
import { Insight, isSearchBasedInsight } from '../../../../core/types'
import { isSearchBackendBasedInsight } from '../../../../core/types/insight/search-insight';
import { BackendInsight } from '../backend-insight/BackendInsight'
import { BuiltInInsight } from '../built-in-insight/BuiltInInsight';

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
    const {
        insight,
        telemetryService,
        settingsCascade,
        platformContext,
        extensionsController,
        ...otherProps
    } = props

    if (isSearchBasedInsight(insight) && isSearchBackendBasedInsight(insight)) {
        return (
            <BackendInsight
                insight={insight}
                telemetryService={telemetryService}
                settingsCascade={settingsCascade}
                platformContext={platformContext}
                {...otherProps}
            />
        )
    }

    // Search based extension and lang stats insight are handled by built-int fetchers
    return (
        <BuiltInInsight
            insight={insight}
            telemetryService={telemetryService}
            settingsCascade={settingsCascade}
            platformContext={platformContext}
            {...otherProps}
        />
    )
}
