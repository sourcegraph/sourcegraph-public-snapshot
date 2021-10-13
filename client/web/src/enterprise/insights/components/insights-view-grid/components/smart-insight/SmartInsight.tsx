import React from 'react'

import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Insight, isSearchBasedInsight } from '../../../../core/types'
import { isSearchBackendBasedInsight } from '../../../../core/types/insight/search-insight'
import { BackendInsight } from '../backend-insight/BackendInsight'
import { BuiltInInsight } from '../built-in-insight/BuiltInInsight'

export interface SmartInsightProps<D extends keyof ViewContexts>
    extends TelemetryProps,
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
    const { insight, telemetryService, where, context, ...otherProps } = props

    if (isSearchBasedInsight(insight) && isSearchBackendBasedInsight(insight)) {
        return <BackendInsight insight={insight} telemetryService={telemetryService} {...otherProps} />
    }

    // Search based extension and lang stats insight are handled by built-in fetchers
    return (
        <BuiltInInsight
            insight={insight}
            telemetryService={telemetryService}
            where={where}
            context={context}
            {...otherProps}
        />
    )
}
