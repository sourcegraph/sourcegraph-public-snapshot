import classnames from 'classnames'
import DatabaseIcon from 'mdi-react/DatabaseIcon'
import React, { useCallback, useContext, useRef, useState } from 'react'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useDebounce } from '@sourcegraph/wildcard'

import { Settings } from '../../../../../schema/settings.schema'
import { InsightsApiContext } from '../../../../core/backend/api-provider'
import { BackendInsightFilters } from '../../../../core/backend/types'
import { SearchBackendBasedInsight } from '../../../../core/types'
import { useDeleteInsight } from '../../../../hooks/use-delete-insight/use-delete-insight'
import { useDistinctValue } from '../../../../hooks/use-distinct-value'
import { useParallelRequests } from '../../../../hooks/use-parallel-requests/use-parallel-request'
import { InsightViewContent } from '../../../insight-view-content/InsightViewContent'
import { InsightErrorContent } from '../insight-card/components/insight-error-content/InsightErrorContent'
import { InsightLoadingContent } from '../insight-card/components/insight-loading-content/InsightLoadingContent'
import { InsightContentCard } from '../insight-card/InsightContentCard'

import styles from './BackendInsight.module.scss'
import { DrillDownFiltersAction } from './components/drill-down-filters-action/DrillDownFiltersPanel'
import { DrillDownFilters, EMPTY_DRILLDOWN_FILTERS } from './components/drill-down-filters-panel/types'

interface BackendInsightProps
    extends TelemetryProps,
        SettingsCascadeProps<Settings>,
        PlatformContextProps<'updateSettings'>,
        React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement> {
    insight: SearchBackendBasedInsight
}

/**
 * Renders BE search based insight. Fetches insight data by gql api handler.
 */
export const BackendInsight: React.FunctionComponent<BackendInsightProps> = props => {
    const { telemetryService, insight, platformContext, settingsCascade, ref, ...otherProps } = props
    const { getBackendInsightById } = useContext(InsightsApiContext)

    const insightCardReference = useRef<HTMLDivElement>(null)
    const [isFiltersOpen, setIsFiltersOpen] = useState(false)
    const [filters, setFilters] = useState<DrillDownFilters>(EMPTY_DRILLDOWN_FILTERS)

    // Currently we support only regexp filters so extract them in a separate object
    // to pass further in a gql api fetcher method
    const regexpFilters = useDistinctValue<BackendInsightFilters>({
        excludeRepoRegexp: filters.excludeRepoRegex,
        includeRepoRegexp: filters.includeRepoRegex,
    })
    const debouncedFilters = useDebounce(regexpFilters, 500)

    const { data, loading, error } = useParallelRequests(
        useCallback(
            () =>
                getBackendInsightById({
                    id: insight.id,
                    filters: debouncedFilters,
                    series: insight.series,
                }),
            [insight.id, insight.series, debouncedFilters, getBackendInsightById]
        )
    )

    const { loading: isDeleting, delete: handleDelete } = useDeleteInsight({
        settingsCascade,
        platformContext,
    })

    const handleDrillDownFiltersChange = (filters: DrillDownFilters): void => {
        setFilters(filters)
    }

    return (
        <InsightContentCard
            insight={{ id: insight.id, view: data?.view }}
            hasContextMenu={true}
            actions={
                <DrillDownFiltersAction
                    isOpen={isFiltersOpen}
                    popoverTargetRef={insightCardReference}
                    filters={filters}
                    onFilterChange={handleDrillDownFiltersChange}
                    onVisibilityChange={setIsFiltersOpen}
                />
            }
            telemetryService={telemetryService}
            onDelete={handleDelete}
            innerRef={insightCardReference}
            {...otherProps}
            className={classnames('be-insight-card', otherProps.className, {
                [styles.cardWithFilters]: isFiltersOpen,
            })}
        >
            {loading || isDeleting ? (
                <InsightLoadingContent
                    text={isDeleting ? 'Deleting code insight' : 'Loading code insight'}
                    subTitle={insight.id}
                    icon={DatabaseIcon}
                />
            ) : isErrorLike(error) ? (
                <InsightErrorContent error={error} title={insight.id} icon={DatabaseIcon} />
            ) : (
                data && (
                    <InsightViewContent
                        telemetryService={telemetryService}
                        viewContent={data.view.content}
                        viewID={insight.id}
                        containerClassName="be-insight-card"
                    />
                )
            )}
            {
                // Passing children props explicitly to render any top-level content like
                // resize-handler from the react-grid-layout library
                otherProps.children
            }
        </InsightContentCard>
    )
}
