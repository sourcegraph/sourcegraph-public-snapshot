import { HTMLAttributes, forwardRef, useEffect } from 'react'

import { useMergeRefs } from 'use-callback-ref'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useSearchParameters } from '@sourcegraph/wildcard'

import { Insight, isBackendInsight } from '../../../core'

import { BackendInsightView } from './backend-insight/BackendInsight'
import { BuiltInInsight } from './built-in-insight/BuiltInInsight'

export interface SmartInsightProps extends TelemetryProps, HTMLAttributes<HTMLElement> {
    insight: Insight
    resizing?: boolean
}

/**
 * Render smart insight with ф(gql or extension api) fetcher and independent mutation
 * actions.
 */
export const SmartInsight = forwardRef<HTMLElement, SmartInsightProps>((props, reference) => {
    const { insight, resizing = false, telemetryService, ...attributes } = props

    const mergedReference = useMergeRefs([reference])
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
                ref={mergedReference}
                insight={insight}
                resizing={resizing}
                telemetryService={telemetryService}
                {...attributes}
            />
        )
    }

    // Lang-stats insight is handled by built-in fetchers
    return (
        <BuiltInInsight
            innerRef={mergedReference}
            insight={insight}
            resizing={resizing}
            telemetryService={telemetryService}
            {...attributes}
        />
    )
})

SmartInsight.displayName = 'SmartInsight'
