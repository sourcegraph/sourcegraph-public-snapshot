import classnames from 'classnames'
import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React, { useContext, useMemo } from 'react'

import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { Settings } from '../../../../../schema/settings.schema'
import { InsightsApiContext } from '../../../../core/backend/api-provider'
import { LangStatsInsight } from '../../../../core/types'
import { SearchExtensionBasedInsight } from '../../../../core/types/insight/search-insight'
import { useDeleteInsight } from '../../../../hooks/use-delete-insight/use-delete-insight'
import { useParallelRequests } from '../../../../hooks/use-parallel-requests/use-parallel-request'
import { InsightViewContent } from '../../../insight-view-content/InsightViewContent'
import { InsightErrorContent } from '../insight-card/components/insight-error-content/InsightErrorContent'
import { InsightLoadingContent } from '../insight-card/components/insight-loading-content/InsightLoadingContent'
import { InsightContentCard } from '../insight-card/InsightContentCard'

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

    const { delete: handleDelete, loading: isDeleting } = useDeleteInsight({ settingsCascade, platformContext })

    return (
        <InsightContentCard
            telemetryService={telemetryService}
            hasContextMenu={true}
            insight={{ id: insight.id, view: data?.view }}
            onDelete={() => handleDelete({ id: insight.id, title: insight.title })}
            {...otherProps}
            className={classnames('extension-insight-card', otherProps.className)}
        >
            {!data || loading || isDeleting ? (
                <InsightLoadingContent
                    text={isDeleting ? 'Deleting code insight' : 'Loading code insight'}
                    subTitle={insight.id}
                    icon={PuzzleIcon}
                />
            ) : isErrorLike(data.view) ? (
                <InsightErrorContent error={data.view} title={insight.id} icon={PuzzleIcon} />
            ) : (
                data.view && (
                    <InsightViewContent
                        telemetryService={telemetryService}
                        viewContent={data.view.content}
                        viewID={insight.id}
                        containerClassName="extension-insight-card"
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
