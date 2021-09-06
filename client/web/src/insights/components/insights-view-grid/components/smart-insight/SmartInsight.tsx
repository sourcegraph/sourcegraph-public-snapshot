import React from 'react'

import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Settings } from '../../../../../schema/settings.schema'
import { Insight, isSearchBasedInsight } from '../../../../core/types'
import { isSearchBackendBasedInsight } from '../../../../core/types/insight/search-insight'
import { BackendInsight } from '../backend-insight/BackendInsight'
import { BuiltInInsight } from '../built-in-insight/BuiltInInsight'

export interface SmartInsightProps<D extends keyof ViewContexts>
    extends TelemetryProps,
        SettingsCascadeProps<Settings>,
        PlatformContextProps<'updateSettings'>,
        React.HTMLAttributes<HTMLElement> {
    insight: Insight

    where: D
    context: ViewContexts[D]
}

/**
 * Render smart insight with (gql or extension api) fetcher and independent mutation
 * actions.
 */
export function SmartInsight<D extends keyof ViewContexts>(props: SmartInsightProps<D>): React.ReactElement {
    const { insight, telemetryService, settingsCascade, platformContext, where, context, ...otherProps } = props

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

    // Search based extension and lang stats insight are handled by built-in fetchers
    return (
        <BuiltInInsight
            insight={insight}
            telemetryService={telemetryService}
            settingsCascade={settingsCascade}
            platformContext={platformContext}
            where={where}
            context={context}
            {...otherProps}
        />
    )
}
