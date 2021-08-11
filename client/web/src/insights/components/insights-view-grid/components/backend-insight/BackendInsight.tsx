import classnames from 'classnames'
import DatabaseIcon from 'mdi-react/DatabaseIcon'
import React, { useCallback, useContext, useRef, useState } from 'react'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useDebounce } from '@sourcegraph/wildcard'

import { Settings } from '../../../../../schema/settings.schema'
import { InsightsApiContext } from '../../../../core/backend/api-provider'
import { BackendInsightFilters } from '../../../../core/backend/types'
import { addInsightToSettings } from '../../../../core/settings-action/insights'
import { SearchBackendBasedInsight, SearchBasedBackendFilters } from '../../../../core/types/insight/search-insight'
import { useDeleteInsight } from '../../../../hooks/use-delete-insight/use-delete-insight'
import { useDistinctValue } from '../../../../hooks/use-distinct-value'
import { useParallelRequests } from '../../../../hooks/use-parallel-requests/use-parallel-request'
import { FORM_ERROR, SubmissionErrors } from '../../../form/hooks/useForm'
import { InsightViewContent } from '../../../insight-view-content/InsightViewContent'
import { InsightErrorContent } from '../insight-card/components/insight-error-content/InsightErrorContent'
import { InsightLoadingContent } from '../insight-card/components/insight-loading-content/InsightLoadingContent'
import { InsightContentCard } from '../insight-card/InsightContentCard'

import styles from './BackendInsight.module.scss'
import { DrillDownFiltersAction } from './components/drill-down-filters-action/DrillDownFiltersPanel'
import { EMPTY_DRILLDOWN_FILTERS, getBackendFilters } from './components/drill-down-filters-panel/utils'

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
    const { getBackendInsightById, getSubjectSettings, updateSubjectSettings } = useContext(InsightsApiContext)

    const insightCardReference = useRef<HTMLDivElement>(null)

    // Drill-down filters
    const [isFiltersOpen, setIsFiltersOpen] = useState(false)
    const [filters, setFilters] = useState<SearchBasedBackendFilters>(insight.filters ?? EMPTY_DRILLDOWN_FILTERS)

    const debouncedFilters = useDebounce(useDistinctValue<BackendInsightFilters>(getBackendFilters(filters)), 500)

    const handleDrillDownFiltersChange = (filters: SearchBasedBackendFilters): void => {
        setFilters(filters)
    }

    const handleFilterSave = async (filters: SearchBasedBackendFilters): Promise<SubmissionErrors> => {
        const subjectId = insight.visibility

        try {
            const settings = await getSubjectSettings(subjectId).toPromise()
            const insightWithNewFilters: SearchBackendBasedInsight = {
                ...insight,
                filters,
            }

            const editedSettings = addInsightToSettings(settings.contents, insightWithNewFilters)

            await updateSubjectSettings(platformContext, subjectId, editedSettings).toPromise()

            telemetryService.log('CodeInsightsSearchBasedFilterUpdatingClick')
            setIsFiltersOpen(false)
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }

    // Loading the insight backend data
    const { data, loading, error } = useParallelRequests(
        useCallback(() => getBackendInsightById(insight.id, debouncedFilters), [
            insight.id,
            debouncedFilters,
            getBackendInsightById,
        ])
    )

    // Handle insight delete action
    const { loading: isDeleting, delete: handleDelete } = useDeleteInsight({
        settingsCascade,
        platformContext,
    })

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
                    onFilterSave={handleFilterSave}
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
