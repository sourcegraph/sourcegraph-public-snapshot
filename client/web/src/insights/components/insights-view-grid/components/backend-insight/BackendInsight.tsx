import classnames from 'classnames'
import DatabaseIcon from 'mdi-react/DatabaseIcon'
import FilterOutlineIcon from 'mdi-react/FilterOutlineIcon'
import React, { useCallback, useContext, useRef } from 'react'
import FocusLock from 'react-focus-lock'
import { UncontrolledPopover } from 'reactstrap'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { Settings } from '../../../../../schema/settings.schema'
import { InsightsApiContext } from '../../../../core/backend/api-provider'
import { SearchBackendBasedInsight } from '../../../../core/types'
import { useDeleteInsight } from '../../../../hooks/use-delete-insight/use-delete-insight'
import { useParallelRequests } from '../../../../hooks/use-parallel-requests/use-parallel-request'
import { InsightViewContent } from '../../../insight-view-content/InsightViewContent'
import { InsightErrorContent } from '../insight-card/components/insight-error-content/InsightErrorContent'
import { InsightLoadingContent } from '../insight-card/components/insight-loading-content/InsightLoadingContent'
import { InsightContentCard } from '../insight-card/InsightContentCard'

import styles from './BackendInsight.module.scss'
import { DrillDownFiltersPanel } from './components/drill-down-filters/DrillDownFiltersPanel'

interface BackendInsightProps
    extends TelemetryProps,
        SettingsCascadeProps<Settings>,
        PlatformContextProps<'updateSettings'>,
        React.HTMLAttributes<HTMLElement> {
    insight: SearchBackendBasedInsight
    drilldown?: boolean
}

/**
 * Renders BE search based insight. Fetches insight data by gql api handler.
 */
export const BackendInsight: React.FunctionComponent<BackendInsightProps> = props => {
    const { telemetryService, insight, platformContext, settingsCascade, drilldown, ...otherProps } = props
    const { getBackendInsightById } = useContext(InsightsApiContext)

    const { data, loading, error } = useParallelRequests(
        useCallback(() => getBackendInsightById(insight.id), [insight.id, getBackendInsightById])
    )

    const { loading: isDeleting, delete: handleDelete } = useDeleteInsight({
        settingsCascade,
        platformContext,
    })

    return (
        <InsightContentCard
            insight={{ id: insight.id, view: data?.view }}
            hasContextMenu={true}
            actions={drilldown && <DrillDownFilters active={true} />}
            telemetryService={telemetryService}
            onDelete={handleDelete}
            {...otherProps}
            className={classnames('be-insight-card', otherProps.className)}
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

interface DrillDownFiltersProps {
    active?: boolean
}

const DrillDownFilters: React.FunctionComponent<DrillDownFiltersProps> = props => {
    const { active } = props
    const targetButtonReference = useRef<HTMLButtonElement>(null)

    return (
        <>
            <button
                ref={targetButtonReference}
                type="button"
                className={classnames('btn btn-icon btn-secondary rounded-circle p-1', styles.filterButton, {
                    [styles.filterButtonActive]: active,
                })}
            >
                <FilterOutlineIcon size="1rem" />
            </button>

            <UncontrolledPopover
                placement="right-start"
                target={targetButtonReference}
                trigger="legacy"
                hideArrow={true}
                fade={false}
                popperClassName="border-0"
            >
                <FocusLock returnFocus={true}>
                    <DrillDownFiltersPanel className={classnames(styles.filterPanel)} />
                </FocusLock>
            </UncontrolledPopover>
        </>
    )
}
