import React, { useContext, useMemo } from 'react'
import { EMPTY, from } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { ExtensionViewsSectionCommonProps } from '../../../insights/sections/types'
import { isCodeInsightsEnabled } from '../../../insights/utils/is-code-insights-enabled'
import { StaticView, ViewGrid } from '../../../views'
import { SmartInsight } from '../components/insights-view-grid/components/smart-insight/SmartInsight'
import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'
import { CodeInsightsSettingsCascadeBackend } from '../core/backend/code-insights-setting-cascade-backend'
import { Insight } from '../core/types'

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
    const { platformContext, settingsCascade, extensionsController, uri, telemetryService, className = '' } = props

    const showCodeInsights = isCodeInsightsEnabled(settingsCascade, { directory: true })

    const api = useMemo(() => {
        console.log('recreate api context')

        return new CodeInsightsSettingsCascadeBackend(settingsCascade, platformContext)
    }, [platformContext, settingsCascade])

    if (!showCodeInsights) {
        return null
    }

    return (
        <CodeInsightsBackendContext.Provider value={api}>
            <ExtensionViewsDirectorySectionContent
                where="directory"
                uri={uri}
                extensionsController={extensionsController}
                platformContext={platformContext}
                settingsCascade={settingsCascade}
                telemetryService={telemetryService}
                className={className}
            />
        </CodeInsightsBackendContext.Provider>
    )
}

const EMPTY_INSIGHT_LIST: Insight[] = []

const ExtensionViewsDirectorySectionContent: React.FunctionComponent<ExtensionViewsDirectorySectionProps> = props => {
    const { extensionsController, uri, className } = props

    const { getInsights } = useContext(CodeInsightsBackendContext)

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
    const insights = useObservable(useMemo(() => getInsights(), [getInsights])) ?? EMPTY_INSIGHT_LIST

    // Pull extension views with Extension API
    const extensionViews =
        useObservable(
            useMemo(
                () =>
                    workspaceUri
                        ? from(extensionsController.extHostAPI).pipe(
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
                [workspaceUri, uri, extensionsController]
            )
        ) ?? EMPTY_EXTENSION_LIST

    const allViewIds = useMemo(() => [...extensionViews, ...insights].map(view => view.id), [extensionViews, insights])

    if (!directoryPageContext) {
        return null
    }

    return (
        <ViewGrid viewIds={allViewIds} telemetryService={props.telemetryService} className={className}>
            {/* Render extension views for the directory page */}
            {extensionViews.map(view => (
                <StaticView key={view.id} view={view} telemetryService={props.telemetryService} />
            ))}
            {/* Render all code insights with proper directory page context */}
            {insights.map(insight => (
                <SmartInsight
                    key={insight.id}
                    insight={insight}
                    telemetryService={props.telemetryService}
                    where="directory"
                    context={directoryPageContext}
                />
            ))}
        </ViewGrid>
    )
}
