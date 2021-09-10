import classnames from 'classnames'
import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React, { useContext, useMemo, useState } from 'react'

import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { Settings } from '../../../../../../schema/settings.schema'
import {
    ViewCard,
    ViewLoadingContent,
    ViewErrorContent,
    ViewContent,
    LineChartSettingsContext,
} from '../../../../../../views'
import { InsightsApiContext } from '../../../../core/backend/api-provider'
import { LangStatsInsight } from '../../../../core/types'
import { SearchExtensionBasedInsight } from '../../../../core/types/insight/search-insight'
import { useDeleteInsight } from '../../../../hooks/use-delete-insight/use-delete-insight'
import { useParallelRequests } from '../../../../hooks/use-parallel-requests/use-parallel-request'
import { InsightContextMenu } from '../insight-context-menu/InsightContextMenu'

interface BuiltInInsightProps<D extends keyof ViewContexts>
    extends TelemetryProps,
        PlatformContextProps<'updateSettings'>,
        SettingsCascadeProps<Settings>,
        React.HTMLAttributes<HTMLElement> {
    insight: SearchExtensionBasedInsight | LangStatsInsight
    where: D
    context: ViewContexts[D]
}

/**
 * Historically we had a few insights that were worked via extension API
 * search-based, code-stats insight
 *
 * This component renders insight card that works almost like before with extensions
 * Component sends FE network request to get and process information but does that in
 * main work thread instead of using Extension API.
 */
export function BuiltInInsight<D extends keyof ViewContexts>(props: BuiltInInsightProps<D>): React.ReactElement {
    const { insight, telemetryService, settingsCascade, platformContext, where, context, ...otherProps } = props
    const { getBuiltInInsight } = useContext(InsightsApiContext)

    const { data, loading } = useParallelRequests(
        useMemo(() => () => getBuiltInInsight(insight, { where, context }), [
            getBuiltInInsight,
            insight,
            where,
            context,
        ])
    )

    // Visual line chart settings
    const [zeroYAxisMin, setZeroYAxisMin] = useState(false)
    const { delete: handleDelete, loading: isDeleting } = useDeleteInsight({ settingsCascade, platformContext })

    return (
        <ViewCard
            {...otherProps}
            insight={{ id: insight.id, view: data?.view }}
            className={classnames('extension-insight-card', otherProps.className)}
            contextMenu={
                <InsightContextMenu
                    insightID={insight.id}
                    menuButtonClassName="ml-1 mr-n2 d-inline-flex"
                    zeroYAxisMin={zeroYAxisMin}
                    onToggleZeroYAxisMin={() => setZeroYAxisMin(!zeroYAxisMin)}
                    onDelete={() => handleDelete(insight)}
                />
            }
        >
            {!data || loading || isDeleting ? (
                <ViewLoadingContent
                    text={isDeleting ? 'Deleting code insight' : 'Loading code insight'}
                    subTitle={insight.id}
                    icon={PuzzleIcon}
                />
            ) : isErrorLike(data.view) ? (
                <ViewErrorContent error={data.view} title={insight.id} icon={PuzzleIcon} />
            ) : (
                data.view && (
                    <LineChartSettingsContext.Provider value={{ zeroYAxisMin }}>
                        <ViewContent
                            telemetryService={telemetryService}
                            viewContent={data.view.content}
                            viewID={insight.id}
                            containerClassName="extension-insight-card"
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
