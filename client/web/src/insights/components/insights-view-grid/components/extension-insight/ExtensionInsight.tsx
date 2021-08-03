import classnames from 'classnames'
import React, { useContext, useMemo } from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { Settings } from '../../../../../schema/settings.schema'
import { InsightsApiContext } from '../../../../core/backend/api-provider'
import { ViewInsightProviderSourceType } from '../../../../core/backend/types'
import { useDeleteInsight } from '../../../../hooks/use-delete-insight/use-delete-insight'
import { useParallelRequests } from '../../../../hooks/use-parallel-request'
import { InsightViewContent } from '../../../insight-view-content/InsightViewContent'
import { InsightErrorContent } from '../insight-card/components/insight-error-content/InsightErrorContent'
import { InsightLoadingContent } from '../insight-card/components/insight-loading-content/InsightLoadingContent'
import { getInsightViewIcon, InsightContentCard } from '../insight-card/InsightContentCard'

interface ExtensionInsightProps
    extends TelemetryProps,
        PlatformContextProps<'updateSettings'>,
        SettingsCascadeProps<Settings>,
        ExtensionsControllerProps,
        React.HTMLAttributes<HTMLElement> {
    viewId: string
}

export const ExtensionInsight: React.FunctionComponent<ExtensionInsightProps> = props => {
    const { viewId, telemetryService, settingsCascade, platformContext, extensionsController, ...otherProps } = props
    const { getExtensionViewById } = useContext(InsightsApiContext)

    const { data, loading } = useParallelRequests(
        useMemo(() => () => getExtensionViewById(viewId, extensionsController.extHostAPI), [
            extensionsController,
            getExtensionViewById,
            viewId,
        ])
    )

    const { delete: handleDelete, loading: isDeleting } = useDeleteInsight({ settingsCascade, platformContext })

    return (
        <InsightContentCard
            telemetryService={telemetryService}
            hasContextMenu={true}
            insight={{ id: viewId, view: data?.view }}
            onDelete={handleDelete}
            {...otherProps}
            className={classnames('extension-insight-card', otherProps.className)}
        >
            {!data || loading || isDeleting ? (
                <InsightLoadingContent
                    text={isDeleting ? 'Deleting code insight' : 'Loading code insight'}
                    subTitle={viewId}
                    icon={getInsightViewIcon(ViewInsightProviderSourceType.Extension)}
                />
            ) : isErrorLike(data.view) ? (
                <InsightErrorContent
                    error={data.view}
                    title={viewId}
                    icon={getInsightViewIcon(ViewInsightProviderSourceType.Extension)}
                />
            ) : (
                data.view && (
                    <InsightViewContent
                        telemetryService={telemetryService}
                        viewContent={data.view.content}
                        viewID={viewId}
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
