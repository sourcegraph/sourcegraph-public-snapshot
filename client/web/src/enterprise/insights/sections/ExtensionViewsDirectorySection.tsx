import React, { useMemo } from 'react'
import { EMPTY, from } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { ExtensionViewsSectionCommonProps } from '../../../insights/sections/types'
import { isCodeInsightsEnabled } from '../../../insights/utils/is-code-insights-enabled'
import { StaticView, ViewGrid } from '../../../views'
import { SmartInsight } from '../components/insights-view-grid/components/smart-insight/SmartInsight'
import { useAllInsights } from '../hooks/use-insight/use-insight'

export interface ExtensionViewsDirectorySectionProps extends ExtensionViewsSectionCommonProps {
    where: 'directory'
    uri: string
}

const EMPTY_EXTENSION_LIST: ViewProviderResult[] = []

/**
 * Renders extension views section for the directory page. Note that this component is used only for
 * Enterprise version. For Sourcegraph OSS see `./src/insights/sections` components.
 */
export const ExtensionViewsDirectorySection: React.FunctionComponent<ExtensionViewsDirectorySectionProps> = props => {
    const { settingsCascade, extensionsController, uri, className = '' } = props

    const showCodeInsights = isCodeInsightsEnabled(settingsCascade, { directory: true })

    const workspaceUri = useObservable(
        useMemo(
            () =>
                from(extensionsController.extHostAPI).pipe(
                    switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getWorkspaceRoots())),
                    map(workspaceRoots => workspaceRoots[0]?.uri)
                ),
            [extensionsController]
        )
    )

    const directoryPageContext = useMemo(
        () =>
            workspaceUri && {
                viewer: {
                    type: 'DirectoryViewer' as const,
                    directory: {
                        uri: new URL(uri),
                    },
                },
                workspace: {
                    uri: new URL(workspaceUri),
                },
            },
        [uri, workspaceUri]
    )

    // Read code insights views from the settings cascade
    const insights = useAllInsights({ settingsCascade })

    // Pull extension views with Extension API
    const extensionViews =
        useObservable(
            useMemo(
                () =>
                    showCodeInsights && workspaceUri
                        ? from(props.extensionsController.extHostAPI).pipe(
                              switchMap(extensionHostAPI =>
                                  wrapRemoteObservable(
                                      extensionHostAPI.getDirectoryViews({
                                          viewer: {
                                              type: 'DirectoryViewer',
                                              directory: { uri },
                                          },
                                          workspace: { uri: workspaceUri },
                                      })
                                  )
                              )
                          )
                        : EMPTY,
                [showCodeInsights, workspaceUri, uri, props.extensionsController]
            )
        ) ?? EMPTY_EXTENSION_LIST

    const allViewIds = useMemo(() => [...extensionViews, ...insights].map(view => view.id), [extensionViews, insights])

    if (!showCodeInsights) {
        return null
    }

    return (
        <ViewGrid viewIds={allViewIds} telemetryService={props.telemetryService} className={className}>
            {/* Render extension views for the directory page */}
            {extensionViews.map(view => (
                <StaticView key={view.id} view={view} telemetryService={props.telemetryService} />
            ))}

            {/* Render all code insights with proper directory page context */}
            {directoryPageContext
                ? insights.map(insight => (
                      <SmartInsight
                          key={insight.id}
                          insight={insight}
                          telemetryService={props.telemetryService}
                          platformContext={props.platformContext}
                          settingsCascade={settingsCascade}
                          where="directory"
                          context={directoryPageContext}
                      />
                  ))
                : []}
        </ViewGrid>
    )
}
