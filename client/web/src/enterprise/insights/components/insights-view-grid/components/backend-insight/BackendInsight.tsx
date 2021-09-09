import classnames from 'classnames'
import DatabaseIcon from 'mdi-react/DatabaseIcon'
import React, { useCallback, useContext, useRef, useState } from 'react'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useDebounce } from '@sourcegraph/wildcard'

import { Settings } from '../../../../../../schema/settings.schema'
import {
    ViewCard,
    ViewContent,
    ViewErrorContent,
    ViewLoadingContent,
    LineChartSettingsContext,
} from '../../../../../../views'
import { InsightsApiContext } from '../../../../core/backend/api-provider'
import { InsightStillProcessingError } from '../../../../core/backend/api/get-backend-insight'
import { addInsightToSettings } from '../../../../core/settings-action/insights'
import { SearchBackendBasedInsight, SearchBasedBackendFilters } from '../../../../core/types/insight/search-insight'
import { useDeleteInsight } from '../../../../hooks/use-delete-insight/use-delete-insight'
import { useDistinctValue } from '../../../../hooks/use-distinct-value'
import { useParallelRequests } from '../../../../hooks/use-parallel-requests/use-parallel-request'
import { DashboardInsightsContext } from '../../../../pages/dashboards/dashboard-page/components/dashboards-content/components/dashboard-inisghts/DashboardInsightsContext'
import { FORM_ERROR, SubmissionErrors } from '../../../form/hooks/useForm'
import { InsightContextMenu } from '../insight-context-menu/InsightContextMenu'

import { BackendAlertOverlay } from './BackendAlertOverlay'
import styles from './BackendInsight.module.scss'
import { DrillDownFiltersAction } from './components/drill-down-filters-action/DrillDownFiltersPanel'
import { DrillDownInsightCreationFormValues } from './components/drill-down-filters-panel/components/drill-down-insight-creation-form/DrillDownInsightCreationForm'
import { EMPTY_DRILLDOWN_FILTERS } from './components/drill-down-filters-panel/utils'
import { useInsightFilterCreation } from './hooks/use-insight-filter-creation'

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

    const { dashboard } = useContext(DashboardInsightsContext)
    const { getBackendInsight, getSubjectSettings, updateSubjectSettings } = useContext(InsightsApiContext)

    // Visual line chart settings
    const [zeroYAxisMin, setZeroYAxisMin] = useState(false)
    const insightCardReference = useRef<HTMLDivElement>(null)

    // Use deep copy check in case if a setting subject has re-created copy of
    // the insight config with same structure and values. To avoid insight data
    // re-fetching.
    const cachedInsight = useDistinctValue(insight)

    // Original insight filters values that are stored in setting subject with insight
    // configuration object, They are updated  whenever the user clicks update/save button
    const [originalInsightFilters, setOriginalInsightFilters] = useState(
        cachedInsight.filters ?? EMPTY_DRILLDOWN_FILTERS
    )

    // Live valid filters from filter form. They are updated whenever the user is changing
    // filter value in filters fields.
    const [filters, setFilters] = useState<SearchBasedBackendFilters>(originalInsightFilters)
    const [isFiltersOpen, setIsFiltersOpen] = useState(false)
    const debouncedFilters = useDebounce(useDistinctValue<SearchBasedBackendFilters>(filters), 500)

    // Loading the insight backend data
    const { data, loading, error } = useParallelRequests(
        useCallback(
            () =>
                getBackendInsight({
                    ...cachedInsight,
                    filters: debouncedFilters,
                }),
            [cachedInsight, debouncedFilters, getBackendInsight]
        )
    )

    // Handle insight delete action
    const { loading: isDeleting, delete: handleDelete } = useDeleteInsight({
        settingsCascade,
        platformContext,
    })

    const { create: creteInsightWithFilters } = useInsightFilterCreation({ platformContext })

    const handleFilterSave = async (filters: SearchBasedBackendFilters): Promise<SubmissionErrors> => {
        const subjectId = insight.visibility

        try {
            const settings = await getSubjectSettings(subjectId).toPromise()
            const insightWithNewFilters: SearchBackendBasedInsight = { ...insight, filters }
            const editedSettings = addInsightToSettings(settings.contents, insightWithNewFilters)

            await updateSubjectSettings(platformContext, subjectId, editedSettings).toPromise()

            telemetryService.log('CodeInsightsSearchBasedFilterUpdating')

            setOriginalInsightFilters(filters)
            setIsFiltersOpen(false)
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }

    const handleInsightFilterCreation = async (
        values: DrillDownInsightCreationFormValues
    ): Promise<SubmissionErrors> => {
        const { insightName } = values

        if (!dashboard) {
            return
        }

        try {
            await creteInsightWithFilters({
                insightName,
                filters,
                dashboard,
                originalInsight: insight,
            })

            telemetryService.log('CodeInsightsSearchBasedFilterInsightCreation')
            setOriginalInsightFilters(filters)
            setIsFiltersOpen(false)
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }

    return (
        <ViewCard
            {...otherProps}
            insight={{ id: insight.id, view: data?.view }}
            contextMenu={
                <InsightContextMenu
                    insightID={insight.id}
                    menuButtonClassName="ml-1 mr-n2 d-inline-flex"
                    zeroYAxisMin={zeroYAxisMin}
                    onToggleZeroYAxisMin={() => setZeroYAxisMin(!zeroYAxisMin)}
                    onDelete={() => handleDelete(insight)}
                />
            }
            actions={
                <DrillDownFiltersAction
                    isOpen={isFiltersOpen}
                    settings={settingsCascade.final ?? {}}
                    popoverTargetRef={insightCardReference}
                    initialFiltersValue={filters}
                    originalFiltersValue={originalInsightFilters}
                    onFilterChange={setFilters}
                    onFilterSave={handleFilterSave}
                    onInsightCreate={handleInsightFilterCreation}
                    onVisibilityChange={setIsFiltersOpen}
                />
            }
            innerRef={insightCardReference}
            className={classnames('be-insight-card', otherProps.className, {
                [styles.cardWithFilters]: isFiltersOpen,
            })}
        >
            {loading || isDeleting ? (
                <ViewLoadingContent
                    text={isDeleting ? 'Deleting code insight' : 'Loading code insight'}
                    subTitle={insight.id}
                    icon={DatabaseIcon}
                />
            ) : isErrorLike(error) ? (
                <ViewErrorContent error={error} title={insight.id} icon={DatabaseIcon}>
                    {error instanceof InsightStillProcessingError ? (
                        <div className="alert alert-info m-0">{error.message}</div>
                    ) : null}
                </ViewErrorContent>
            ) : (
                data && (
                    <LineChartSettingsContext.Provider value={{ zeroYAxisMin }}>
                        <ViewContent
                            telemetryService={telemetryService}
                            viewContent={data.view.content}
                            viewID={insight.id}
                            containerClassName="be-insight-card"
                            alert={
                                <BackendAlertOverlay
                                    hasNoData={!data.view.content.some(({ data }) => data.length > 0)}
                                    isFetchingHistoricalData={data.view.isFetchingHistoricalData}
                                />
                            }
                        />
                    </LineChartSettingsContext.Provider>
                )
            )}
            {
                // Passing children props explicitly to render any top-level content like
                // resize-handler from the react-grid-layout library
                otherProps.children
            }
        </ViewCard>
    )
}
