import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ViewInsightProviderResult } from '../../../../core/backend/types'
import { InsightViewContent } from '../../../insight-view-content/InsightViewContent'
import { InsightErrorContent } from '../insight-card/components/insight-error-content/InsightErrorContent'
import { InsightLoadingContent } from '../insight-card/components/insight-loading-content/InsightLoadingContent'
import { getInsightViewIcon, InsightContentCard } from '../insight-card/InsightContentCard'

interface StaticView extends TelemetryProps, React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement> {
    view: ViewInsightProviderResult
}

/**
 * Component that renders insight-like extension card. Used by extension views in extension
 * consumers that have insight section (the search and the directory page).
 */
export const StaticView: React.FunctionComponent<StaticView> = props => {
    const { view, telemetryService, ...otherProps } = props

    return (
        <InsightContentCard
            data-testid={`insight-card.${view.id}`}
            telemetryService={telemetryService}
            hasContextMenu={false}
            insight={view}
            className="insight-content-card"
            {...otherProps}
        >
            {view.view === undefined ? (
                <InsightLoadingContent
                    text="Loading code insight"
                    subTitle={view.id}
                    icon={getInsightViewIcon(view.source)}
                />
            ) : isErrorLike(view.view) ? (
                <InsightErrorContent error={view.view} title={view.id} icon={getInsightViewIcon(view.source)} />
            ) : (
                <InsightViewContent
                    telemetryService={telemetryService}
                    viewContent={view.view.content}
                    viewID={view.id}
                    containerClassName="insight-content-card"
                />
            )}
        </InsightContentCard>
    )
}
