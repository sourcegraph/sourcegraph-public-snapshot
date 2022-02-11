import React, { forwardRef, ReactElement, Ref } from 'react'

import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Insight, isBackendInsight } from '../../../../core/types'
import { BackendInsightView } from '../backend-insight/BackendInsight'
import { BuiltInInsight } from '../built-in-insight/BuiltInInsight'

export interface SmartInsightProps<D extends keyof ViewContexts>
    extends TelemetryProps,
        React.HTMLAttributes<HTMLElement> {
    insight: Insight

    where: D
    context: ViewContexts[D]
    resizing?: boolean
}

/**
 * Render smart insight with (gql or extension api) fetcher and independent mutation
 * actions.
 */
export const SmartInsight = forwardRef<HTMLElement, SmartInsightProps<keyof ViewContexts>>((props, reference) => {
    const { insight, resizing = false, telemetryService, where, context, ...otherProps } = props

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
            where={where}
            context={context}
            innerRef={reference}
            {...otherProps}
        />
    )
    // Cast here is needed since forwardRef doesn't support generics properly
    // cause of static nature
}) as <D extends keyof ViewContexts>(p: SmartInsightProps<D> & { ref?: Ref<HTMLElement> }) => ReactElement
