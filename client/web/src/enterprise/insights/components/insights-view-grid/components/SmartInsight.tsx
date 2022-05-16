import React, { forwardRef } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Insight, isBackendInsight } from '../../../core'

import { BackendInsightView } from './backend-insight/BackendInsight'
import { BuiltInInsight } from './BuiltInInsight'

export interface SmartInsightProps extends TelemetryProps, React.HTMLAttributes<HTMLElement> {
    insight: Insight
    resizing?: boolean
}

/**
 * Render smart insight with (gql or extension api) fetcher and independent mutation
 * actions.
 */
export const SmartInsight = forwardRef<HTMLElement, SmartInsightProps>((props, reference) => {
    const { insight, resizing = false, telemetryService, ...otherProps } = props

    if (isBackendInsight(insight)) {
        return (
            <BackendInsightView
                insight={insight}
                resizing={resizing}
                telemetryService={telemetryService}
                {...otherProps}
                innerRef={reference}
            />
        )
    }

    // Search based extension and lang stats insight are handled by built-in fetchers
    return (
        <BuiltInInsight
            insight={insight}
            resizing={resizing}
            telemetryService={telemetryService}
            innerRef={reference}
            {...otherProps}
        />
    )
})

SmartInsight.displayName = 'SmartInsight'
