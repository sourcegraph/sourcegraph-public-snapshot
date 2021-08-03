import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ViewInsightProviderResult } from '../../core/backend/types'
import { InsightViewContent } from '../insight-view-content/InsightViewContent'

import { InsightErrorContent } from './components/insight-card/components/insight-error-content/InsightErrorContent'
import { InsightLoadingContent } from './components/insight-card/components/insight-loading-content/InsightLoadingContent'
import { getInsightViewIcon, InsightContentCard } from './components/insight-card/InsightContentCard'
import { ViewGrid } from './components/view-grid/ViewGrid'

export interface StaticInsightsViewGridProps extends TelemetryProps {
    views: ViewInsightProviderResult[]
    className?: string
}

/**
 * Renders insights drag and drop grid with all type of insights
 * (backend, search based, lang stats) by views props data.
 *
 * Static means that insights within the grid are readonly views.
 * Used in all consumers that load insight by themselves like home (search) page, directory page
 */
export const StaticInsightsViewGrid: React.FunctionComponent<StaticInsightsViewGridProps> = props => {
    const { views, telemetryService, className } = props

    return (
        <ViewGrid className={className} viewIds={views.map(view => view.id)} telemetryService={telemetryService}>
            {props.views.map(view => (
                <InsightContentCard
                    key={view.id}
                    data-testid={`insight-card.${view.id}`}
                    telemetryService={telemetryService}
                    hasContextMenu={false}
                    insight={view}
                    className="insight-content-card"
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
            ))}
        </ViewGrid>
    )
}
