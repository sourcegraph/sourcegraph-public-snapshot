import React, { forwardRef, useEffect, useRef } from 'react'

import { useMergeRefs } from 'use-callback-ref'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useSearchParameters } from '@sourcegraph/wildcard'

import { Insight, isBackendInsight } from '../../../core'

import { BackendInsightView } from './backend-insight/BackendInsight'
import { BuiltInInsight } from './built-in-insight/BuiltInInsight'

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
    const localReference = useRef<HTMLElement>(null)
    const mergedReference = useMergeRefs([reference, localReference])
    const search = useSearchParameters()

    useEffect(() => {
        const insightIdToBeFocused = search.get('focused')
        const element = mergedReference.current

        if (element && insightIdToBeFocused === insight.id) {
            element.focus()
        }
    }, [insight.id, mergedReference, search])

    if (isBackendInsight(insight)) {
        return (
            <BackendInsightView
                insight={insight}
                resizing={resizing}
                telemetryService={telemetryService}
                {...otherProps}
                innerRef={mergedReference}
            />
        )
    }

    // Search based extension and lang stats insight are handled by built-in fetchers
    return (
        <BuiltInInsight
            insight={insight}
            resizing={resizing}
            telemetryService={telemetryService}
            innerRef={mergedReference}
            {...otherProps}
        />
    )
})

SmartInsight.displayName = 'SmartInsight'
