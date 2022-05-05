import React, { useMemo } from 'react'

import { EMPTY, from } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { useObservable } from '@sourcegraph/wildcard'

import { StaticView, ViewGrid } from '../../views'
import { isCodeInsightsEnabled } from '../utils/is-code-insights-enabled'

import { ExtensionViewsSectionCommonProps } from './types'

export interface ExtensionViewsDirectorySectionProps extends ExtensionViewsSectionCommonProps {
    where: 'directory'
    uri: string
}

/**
 * Renders extension views section for the directory page. Note that this component is used only for
 * OSS version. For Sourcegraph enterprise see `./enterprise/insights/sections` components.
 */
export const ExtensionViewsDirectorySection: React.FunctionComponent<
    React.PropsWithChildren<ExtensionViewsDirectorySectionProps>
> = props => {
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
        ) ?? []

    if (!showCodeInsights) {
        return null
    }

    return (
        <ViewGrid viewIds={extensionViews.map(view => view.id)} className={className}>
            {/* Render extension views for the directory page */}
            {extensionViews.map(view => (
                <StaticView key={view.id} content={view} telemetryService={props.telemetryService} />
            ))}
        </ViewGrid>
    )
}
